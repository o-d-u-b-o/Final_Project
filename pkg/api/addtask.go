package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"go1f/pkg/db"
)

type taskResponse struct {
	ID    int64  `json:"id,omitempty"`
	Error string `json:"error,omitempty"`
}

func addTaskHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("AddTask request: Method=%s, Path=%s", r.Method, r.URL.Path)

	if r.Method == http.MethodOptions {
		log.Println("Handling OPTIONS request")

		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		log.Printf("Invalid method: %s, expected POST", r.Method)
		writeJSON(w, taskResponse{Error: "Method not allowed"}, http.StatusMethodNotAllowed)
		return
	}

	var task db.Task
	var response taskResponse

	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		log.Printf("JSON decode error: %v", err)
		response.Error = "Invalid JSON format"
		writeJSON(w, response, http.StatusBadRequest)
		return
	}

	if task.Title == "" {
		log.Println("Empty title")
		response.Error = "Task title is required"
		writeJSON(w, response, http.StatusBadRequest)
		return
	}

	now := time.Now().UTC()
	if err := processTaskDate(&task, now); err != nil {
		log.Printf("Date processing error: %v", err)
		response.Error = err.Error()
		writeJSON(w, response, http.StatusBadRequest)
		return
	}

	id, err := db.AddTask(&task)
	if err != nil {
		log.Printf("Database error: %v", err)
		response.Error = "Failed to add task to database"
		writeJSON(w, response, http.StatusInternalServerError)
		return
	}

	log.Printf("Task added successfully, ID: %d", id)
	response.ID = id
	writeJSON(w, response, http.StatusOK)
}

func processTaskDate(task *db.Task, now time.Time) error {
	log.Printf("Input - Date: '%s', Title: '%s', Repeat: '%s'",
		task.Date, task.Title, task.Repeat)
	log.Printf("Current time: %s", now.Format("20060102"))

	// 1. Явная обработка "today"
	if strings.EqualFold(task.Date, "today") {
		task.Date = now.Format("20060102")
		log.Printf("Today case - Setting date to: %s", task.Date)
		return nil
	}

	// 2. Обработка пустой даты
	if task.Date == "" {
		task.Date = now.Format("20060102")
		log.Printf("Empty date case - Setting date to: %s", task.Date)
		return nil
	}

	// 3. Для явных дат
	date, err := time.Parse("20060102", task.Date)
	if err != nil {
		return fmt.Errorf("invalid date format, expected YYYYMMDD")
	}

	// Измененная логика: не считаем сегодняшнюю дату "прошедшей"
	if date.Before(now) && !isSameDay(date, now) {
		if task.Repeat != "" {
			next, err := NextDate(now, task.Date, task.Repeat)
			if err != nil {
				return err
			}
			task.Date = next
			log.Printf("Repeating task - New date: %s", task.Date)
		} else {
			task.Date = now.Format("20060102")
			log.Printf("Past date without repeat - Setting to today: %s", task.Date)
		}
	}

	log.Printf("Output - Final date: %s", task.Date)
	return nil
}

func isSameDay(t1, t2 time.Time) bool {
	return t1.Year() == t2.Year() &&
		t1.Month() == t2.Month() &&
		t1.Day() == t2.Day()
}
