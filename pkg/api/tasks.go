package api

import (
	"go1f/pkg/db"
	"net/http"
	"strconv"
)

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

	// Преобразуем задачи в нужный формат
	var response []map[string]string
	for _, t := range tasks {
		response = append(response, map[string]string{
			"id":      strconv.FormatInt(t.ID, 10),
			"date":    t.Date,
			"title":   t.Title,
			"comment": t.Comment,
			"repeat":  t.Repeat,
		})
	}

	writeJSON(w, map[string]interface{}{"tasks": response}, http.StatusOK)
}
