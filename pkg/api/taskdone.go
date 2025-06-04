package api

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"go1f/pkg/db"
)

func taskDoneHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, ErrorResponse{Error: "Method not allowed"}, http.StatusMethodNotAllowed)
		return
	}

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

	if task.Repeat == "" {
		if err := db.DeleteTask(id); err != nil {
			log.Printf("Delete failed: %v", err)
			writeJSON(w, ErrorResponse{Error: err.Error()}, http.StatusInternalServerError)
			return
		}
	} else {
		now := time.Now().UTC()
		log.Printf("Calculating next date for task %d (date: %s, repeat: %s)",
			id, task.Date, task.Repeat)

		nextDateStr, err := NextDate(now, task.Date, task.Repeat)
		if err != nil {
			log.Printf("NextDate error: %v", err)
			writeJSON(w, ErrorResponse{Error: err.Error()}, http.StatusBadRequest)
			return
		}

		log.Printf("Updating task %d to new date: %s", id, nextDateStr)
		if err := db.UpdateTaskDate(id, nextDateStr); err != nil {
			log.Printf("Update failed: %v", err)
			writeJSON(w, ErrorResponse{Error: err.Error()}, http.StatusInternalServerError)
			return
		}
	}

	writeJSON(w, map[string]interface{}{}, http.StatusOK)
}
