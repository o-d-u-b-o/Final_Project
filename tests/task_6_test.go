package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/o-d-u-b-o/Final_Project/pkg/db"
	"github.com/stretchr/testify/assert"
)

func TestTask(t *testing.T) {
	db := openDB(t)
	defer db.Close()

	now := time.Now()
	task := task{
		date:    now.Format(`20060102`),
		title:   "Созвон в 16:00",
		comment: "Обсуждение планов",
		repeat:  "d 5",
	}

	// Добавляем задачу
	id := addTask(t, task)

	// Проверяем получение задачи
	body, err := requestJSON("api/task?id="+id, nil, http.MethodGet)
	assert.NoError(t, err)

	var resp struct {
		ID      string `json:"id"`
		Date    string `json:"date"`
		Title   string `json:"title"`
		Comment string `json:"comment"`
		Repeat  string `json:"repeat"`
	}
	err = json.Unmarshal(body, &resp)
	assert.NoError(t, err)

	assert.Equal(t, id, resp.ID)
	assert.Equal(t, task.date, resp.Date)
	assert.Equal(t, task.title, resp.Title)
	assert.Equal(t, task.comment, resp.Comment)
	assert.Equal(t, task.repeat, resp.Repeat)
}

type fulltask struct {
	id string
	task
}

func postJSONWithContentType(url string, data []byte, contentType string, method string) (map[string]interface{}, error) {
	req, err := http.NewRequest(method, url, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}

func TestEditTask(t *testing.T) {
	dbConn := openDB(t) // переименуем переменную, чтобы не путать с пакетом db
	defer dbConn.Close()

	now := time.Now()

	tsk := task{
		date:    now.Format(`20060102`),
		title:   "Заказать пиццу",
		comment: "в 17:00",
		repeat:  "",
	}

	id := addTask(t, tsk)

	tbl := []fulltask{
		{"", task{"20240129", "Тест", "", ""}},
		{"abc", task{"20240129", "Тест", "", ""}},
		{"7645346343", task{"20240129", "Тест", "", ""}},
		{id, task{"20240129", "", "", ""}},
		{id, task{"20240192", "Qwerty", "", ""}},
		{id, task{"28.01.2024", "Заголовок", "", ""}},
		{id, task{"20240212", "Заголовок", "", "ooops"}},
	}
	for _, v := range tbl {
		m, err := postJSON("api/task", map[string]any{
			"id":      v.id,
			"date":    v.date,
			"title":   v.title,
			"comment": v.comment,
			"repeat":  v.repeat,
		}, http.MethodPut)
		assert.NoError(t, err)

		var errVal string
		e, ok := m["error"]
		if ok {
			errVal = fmt.Sprint(e)
		}
		assert.NotEqual(t, len(errVal), 0, "Ожидается ошибка для значения %v", v)
	}

	updateTask := func(newVals map[string]any) {
		// Преобразуем данные в JSON
		jsonData, err := json.Marshal(newVals)
		assert.NoError(t, err)

		// Отправляем с правильным Content-Type
		mupd, err := postJSONWithContentType("api/task", jsonData, "application/json", http.MethodPut)
		assert.NoError(t, err)

		if e, ok := mupd["error"]; ok {
			t.Fatalf("Unexpected error: %v", e)
		}

		// Преобразуем ID
		taskID, err := strconv.ParseInt(id, 10, 64)
		assert.NoError(t, err)

		// Получаем задачу
		task, err := db.GetTask(taskID)
		assert.NoError(t, err)
		assert.NotNil(t, task) // Проверяем что задача не nil

		// Проверяем обновленные значения
		assert.Equal(t, id, strconv.FormatInt(task.ID, 10))
		assert.Equal(t, newVals["title"].(string), task.Title)
		assert.Equal(t, newVals["comment"].(string), task.Comment)
		assert.Equal(t, newVals["repeat"].(string), task.Repeat)

		// Проверяем дату
		taskDate, err := time.Parse("20060102", task.Date)
		assert.NoError(t, err)
		assert.False(t, taskDate.Before(time.Now()))
	}

	updateTask(map[string]any{
		"id":      id,
		"date":    now.Format(`20060102`),
		"title":   "Заказать хинкали",
		"comment": "в 18:00",
		"repeat":  "d 7",
	})
}
