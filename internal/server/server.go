package server

import (
	"net/http"
	"os"

	"github.com/golang/glog"
	"wednesday.wtf/godin/internal/server/routes"
)

func initializeRoutes() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		writeErrorResponse(w, http.StatusNotFound, "Route Not Found", "The requested route was not handleds.")
	})

	http.HandleFunc("/v1/upload", routes.UploadHandler)
	http.HandleFunc("/v1/download", routes.DownloadHandler)
}

// Initialize the server package by registering the routes.
func serveRequests() error {
	serverAddres := os.Getenv("SERVER_ADDRESS")

	if err := http.ListenAndServe(serverAddres, nil); err != nil {
		return err
	}

	return nil
}

// Start initializes the server and starts listening for requests.
func Start() {
	initializeRoutes()

	if err := serveRequests(); err != nil {
		glog.Fatalf("Error starting the server: %v", err)
	}
}

func Stop() {
}
