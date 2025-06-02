package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"go1f/pkg/db"
)

func taskHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		handleGetTask(w, r)
	case http.MethodPut:
		handleUpdateTask(w, r)
	case http.MethodDelete:
		handleDeleteTask(w, r)
	default:
		writeJSON(w, ErrorResponse{Error: "Method not allowed"}, http.StatusMethodNotAllowed)
	}
}

func handleGetTask(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		writeJSON(w, ErrorResponse{Error: "ID is required"}, http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeJSON(w, ErrorResponse{Error: "Invalid ID format"}, http.StatusBadRequest)
		return
	}

	task, err := db.GetTask(id)
	if err != nil {
		writeJSON(w, ErrorResponse{Error: err.Error()}, http.StatusNotFound)
		return
	}

	writeJSON(w, task, http.StatusOK)
}

func handleUpdateTask(w http.ResponseWriter, r *http.Request) {
	var task db.Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		writeJSON(w, ErrorResponse{Error: "Invalid JSON format"}, http.StatusBadRequest)
		return
	}

	if task.ID == 0 {
		writeJSON(w, ErrorResponse{Error: "Task ID is required"}, http.StatusBadRequest)
		return
	}

	if task.Title == "" {
		writeJSON(w, ErrorResponse{Error: "Title is required"}, http.StatusBadRequest)
		return
	}

	// Обработка даты
	if task.Date == "" {
		task.Date = time.Now().Format("20060102")
	} else {
		if _, err := time.Parse("20060102", task.Date); err != nil {
			writeJSON(w, ErrorResponse{Error: "Invalid date format, use YYYYMMDD"}, http.StatusBadRequest)
			return
		}
	}

	if err := db.UpdateTask(&task); err != nil {
		writeJSON(w, ErrorResponse{Error: err.Error()}, http.StatusInternalServerError)
		return
	}

	writeJSON(w, struct{}{}, http.StatusOK)
}

// Добавляем новый обработчик для удаления задач
func handleDeleteTask(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		writeJSON(w, ErrorResponse{Error: "ID is required"}, http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeJSON(w, ErrorResponse{Error: "Invalid ID format"}, http.StatusBadRequest)
		return
	}

	if err := db.DeleteTask(id); err != nil {
		writeJSON(w, ErrorResponse{Error: err.Error()}, http.StatusInternalServerError)
		return
	}

	writeJSON(w, struct{}{}, http.StatusOK)
}
