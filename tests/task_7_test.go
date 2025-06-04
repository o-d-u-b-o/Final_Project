package tests

import (
	"encoding/json"
	"net/http"
	"runtime/debug"
	"strconv"
	"testing"
	"time"

	"github.com/o-d-u-b-o/Final_Project/pkg/db"
	"github.com/stretchr/testify/assert"
)

func notFoundTask(t *testing.T, id string) { // Убрали dbConn, если не используется внутри
	body, err := requestJSON("api/task?id="+id, nil, http.MethodGet)
	if !assert.NoError(t, err) {
		return
	}

	var m map[string]interface{}
	if !assert.NoError(t, json.Unmarshal(body, &m)) {
		return
	}

	_, ok := m["error"]
	assert.True(t, ok, "Expected error for non-existent task")
}

func TestDone(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Test panicked: %v\n%s", r, debug.Stack())
		}
	}()

	t.Log("Starting TestDone")
	defer t.Log("Finished TestDone")

	// Инициализация БД
	dbConn := openDB(t) // Должен возвращать *sqlx.DB
	if !assert.NotNil(t, dbConn, "DB connection is nil") {
		return
	}
	defer dbConn.Close()

	// Устанавливаем соединение для всех функций db
	db.SetDBConnection(dbConn)

	now := time.Now()

	// Создаем тестовую задачу
	taskData := task{
		date:  now.Format(`20060102`),
		title: "Свести баланс",
	}

	id := addTask(t, taskData)
	if !assert.NotEmpty(t, id, "Task ID should not be empty") {
		return
	}
	t.Logf("Created task with ID: %s", id)

	// Преобразуем ID
	taskID, err := strconv.ParseInt(id, 10, 64)
	if !assert.NoError(t, err, "Failed to parse task ID") {
		return
	}

	// Отмечаем задачу выполненной
	ret, err := postJSON("api/task/done?id="+id, nil, http.MethodPost)
	if !assert.NoError(t, err, "postJSON failed") {
		return
	}
	assert.Empty(t, ret, "Response should be empty")

	// Вызываем GetTask без dbConn, так как он использует глобальное соединение
	_, err = db.GetTask(taskID) // Теперь передаем только ID
	assert.Error(t, err, "Task should be deleted")

	// Тестируем повторяющуюся задачу
	recurringTask := task{
		title:  "Проверить работу /api/task/done",
		repeat: "d 3",
		date:   now.Format(`20060102`),
	}

	id = addTask(t, recurringTask)
	if !assert.NotEmpty(t, id, "Recurring task ID should not be empty") {
		return
	}
	t.Logf("Created recurring task with ID: %s", id)

	taskID, err = strconv.ParseInt(id, 10, 64)
	if !assert.NoError(t, err, "Failed to parse recurring task ID") {
		return
	}

	for i := 0; i < 3; i++ {
		ret, err := postJSON("api/task/done?id="+id, nil, http.MethodPost)
		if !assert.NoError(t, err, "postJSON failed") {
			return
		}
		assert.Empty(t, ret, "Response should be empty")

		task, err := db.GetTask(taskID) // Без dbConn
		if !assert.NoError(t, err, "Failed to get task") || !assert.NotNil(t, task, "Task is nil") {
			return
		}

		now = now.AddDate(0, 0, 3)
		expectedDate := now.Format(`20060102`)
		assert.Equal(t, expectedDate, task.Date,
			"Task date should be updated to %s, got %s", expectedDate, task.Date)
	}
}

func TestDelTask(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Test panicked: %v\n%s", r, debug.Stack())
		}
	}()

	dbConn := openDB(t)
	if !assert.NotNil(t, dbConn, "DB connection is nil") {
		return
	}
	defer dbConn.Close()

	// Устанавливаем соединение для db-пакета
	db.SetDBConnection(dbConn)

	// Создаем тестовую задачу
	id := addTask(t, task{
		title:  "Временная задача",
		repeat: "d 3",
	})
	if !assert.NotEmpty(t, id, "Task ID should not be empty") {
		return
	}

	// Удаляем задачу
	ret, err := postJSON("api/task?id="+id, nil, http.MethodDelete)
	if !assert.NoError(t, err, "postJSON failed") {
		return
	}
	assert.Empty(t, ret, "Response should be empty")

	// Проверяем, что задача не найдена
	notFoundTask(t, id)

	// Тестируем обработку ошибок
	ret, err = postJSON("api/task", nil, http.MethodDelete)
	if !assert.NoError(t, err, "postJSON failed") {
		return
	}
	assert.NotEmpty(t, ret, "Should return error for missing ID")

	ret, err = postJSON("api/task?id=invalid", nil, http.MethodDelete)
	if !assert.NoError(t, err, "postJSON failed") {
		return
	}
	assert.NotEmpty(t, ret, "Should return error for invalid ID format")
}
