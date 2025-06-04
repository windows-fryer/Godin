package routes

import (
	"net/http"
)

func DownloadHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", "attachment; filename=\"downloaded_file\"")

	w.Write([]byte("This is a placeholder for the downloaded file content."))
}
