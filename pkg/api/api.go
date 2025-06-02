package api

import (
	"log"
	"net/http"
	"time"
)

func Init() {
	corsHandler := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next(w, r)
		}
	}

	http.HandleFunc("/api/nextdate", corsHandler(nextDateHandler))
	http.HandleFunc("/api/task", corsHandler(taskHandler))
	http.HandleFunc("/api/addtask", corsHandler(addTaskHandler))
	http.HandleFunc("/api/tasks", corsHandler(getTaskListHandler))
	http.HandleFunc("/api/signin", corsHandler(signinHandler))
	http.HandleFunc("/api/task/done", corsHandler(taskDoneHandler))

	log.Println("API handlers initialized with CORS support")
}

func nextDateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	nowStr := r.FormValue("now")
	dateStr := r.FormValue("date")
	repeat := r.FormValue("repeat")

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
