package main

import (
	"log"
	"net/http"

	"go1f/pkg/api"
	"go1f/pkg/db"
)

func main() {
	// Инициализация базы данных
	if err := db.Init(); err != nil {
		log.Fatalf("Database initialization failed: %v", err)
	}
	defer db.Close()

	// Инициализация API обработчиков
	api.Init()

	// Статические файлы
	fs := http.FileServer(http.Dir("./web"))
	http.Handle("/", fs)

	log.Println("Server starting on :7540...")
	log.Fatal(http.ListenAndServe(":7540", nil))
}
