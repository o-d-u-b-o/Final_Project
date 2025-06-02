package server

import (
	"log"
	"net/http"
	"os"

	"go1f/pkg/api"
)

const (
	defaultPort = "7540"
	webDir      = "./web"
)

func Start() {
	// Инициализация API обработчиков
	api.Init()

	port := getPort()

	fs := http.FileServer(http.Dir(webDir))
	http.Handle("/", fs)

	log.Printf("Starting server on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func getPort() string {
	if port := os.Getenv("TODO_PORT"); port != "" {
		return port
	}
	return defaultPort
}
