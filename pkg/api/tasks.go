package api

import (
	"go1f/pkg/db"
	"net/http"
)

// TasksResponse - структура для ответа со списком задач
type TasksResponse struct {
	Tasks []*db.Task `json:"tasks"`
}

func getTaskListHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, ErrorResponse{Error: "Method not allowed"}, http.StatusMethodNotAllowed)
		return
	}

	search := r.URL.Query().Get("search")
	tasks, err := db.Tasks(50, search)
	if err != nil {
		writeJSON(w, ErrorResponse{Error: err.Error()}, http.StatusInternalServerError)
		return
	}

	if tasks == nil {
		tasks = make([]*db.Task, 0)
	}

	writeJSON(w, TasksResponse{Tasks: tasks}, http.StatusOK)
}
