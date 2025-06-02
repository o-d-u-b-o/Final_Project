package main

import (
	"go1f/pkg/db"
	"go1f/pkg/server"
	"log"
)

func main() {
	// Инициализация БД (без параметров)
	if err := db.Init(); err != nil {
		log.Fatalf("Database initialization failed: %v", err)
	}
	defer db.Close()

	// Запуск сервера
	server.Start()
}
