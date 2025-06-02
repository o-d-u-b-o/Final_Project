package api

import (
	"net/http"
	"time"
)

func Init() {
	http.HandleFunc("/api/nextdate", nextDateHandler)
	http.HandleFunc("/api/task", taskHandler)
	http.HandleFunc("/api/addtask", addTaskHandler)
	http.HandleFunc("/api/tasks", getTaskListHandler)
}

func nextDateHandler(w http.ResponseWriter, r *http.Request) {
	// Обработка запроса nextdate
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Получаем параметры
	nowStr := r.FormValue("now")
	dateStr := r.FormValue("date")
	repeat := r.FormValue("repeat")

	// Если now не указан, используем текущую дату
	var now time.Time
	if nowStr == "" {
		now = time.Now()
	} else {
		var err error
		now, err = time.Parse(dateFormat, nowStr)
		if err != nil {
			http.Error(w, "Invalid now parameter", http.StatusBadRequest)
			return
		}
	}

	// Вычисляем следующую дату
	result, err := NextDate(now, dateStr, repeat)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(result))
}
