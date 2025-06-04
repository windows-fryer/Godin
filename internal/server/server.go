package server

import (
	"net/http"

	"github.com/charmbracelet/log"
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
	log.Info("[SERVER] Started")

	go func() {
		initializeRoutes()

		startServer()
	}()
}

func Stop() {
	log.Info("[SERVER] Stopped")
}
