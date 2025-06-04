package server

import (
	"log"
	"net/http"

	"github.com/godin/internal/server/routes"
)

const SERVER_URL = "localhost:8080"

func initializeRoutes() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	})

	http.HandleFunc("/api/v1/download/", routes.DownloadHandler)
}

func startServer() {
	if err := http.ListenAndServe(SERVER_URL, nil); err != nil {
		panic("Failed to start server: " + err.Error())
	}
}

func Start() {
	go func() {
		initializeRoutes()

		log.Printf("Starting server on %s...", SERVER_URL)
		startServer()
	}()
}
