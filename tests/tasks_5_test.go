package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func addTask(t *testing.T, task task) string {
	ret, err := postJSON("api/addtask", map[string]any{
		"date":    task.date,
		"title":   task.title,
		"comment": task.comment,
		"repeat":  task.repeat,
	}, http.MethodPost)
	assert.NoError(t, err)
	assert.NotNil(t, ret["id"])
	id := fmt.Sprint(ret["id"])
	assert.NotEmpty(t, id)
	return id
}

func getTasks(t *testing.T, search string) []map[string]string {
	url := "api/tasks"
	if Search {
		url += "?search=" + search
	}
	body, err := requestJSON(url, nil, http.MethodGet)
	assert.NoError(t, err)

	var m map[string][]map[string]string
	err = json.Unmarshal(body, &m)
	assert.NoError(t, err)
	return m["tasks"]
}

func TestTasks(t *testing.T) {
	db := openDB(t)
	defer db.Close()

	now := time.Now()
	_, err := db.Exec("DELETE FROM scheduler")
	assert.NoError(t, err)

	// Проверяем пустую базу
	tasks := getTasks(t, "")
	assert.Equal(t, 0, len(tasks))

	// Добавляем задачи
	addTask(t, task{
		date:    now.Format(`20060102`),
		title:   "Просмотр фильма",
		comment: "с попкорном",
		repeat:  "",
	})

	now = now.AddDate(0, 0, 1)
	date := now.Format(`20060102`)
	addTask(t, task{
		date:    date,
		title:   "Сходить в бассейн",
		comment: "",
		repeat:  "",
	})

	addTask(t, task{
		date:    date,
		title:   "Оплатить коммуналку",
		comment: "",
		repeat:  "d 30",
	})

	// Проверяем количество задач
	tasks = getTasks(t, "")
	assert.Equal(t, 3, len(tasks))
}
