package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
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

	now := time.Now()
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
	// Если дата не указана - используем сегодня
	if task.Date == "" {
		task.Date = now.Format(dateFormat)
		return nil
	}

	// Парсим дату
	date, err := time.Parse(dateFormat, task.Date)
	if err != nil {
		return fmt.Errorf("invalid date format, expected YYYYMMDD")
	}

	// Если дата в прошлом и есть правило повторения
	if date.Before(now) && task.Repeat != "" {
		next, err := NextDate(now, task.Date, task.Repeat)
		if err != nil {
			return err
		}
		task.Date = next
	} else if date.Before(now) {
		// Если дата в прошлом и нет правила повторения - используем сегодня
		task.Date = now.Format(dateFormat)
	}

	return nil
}
