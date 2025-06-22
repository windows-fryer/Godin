package routes

import (
	"net/http"

	"github.com/golang/glog"
	"wednesday.wtf/godin/api/service"
)

func FileHandler(w http.ResponseWriter, r *http.Request) {
	serviceID := getServiceID(r)

	if cdnService, exists := service.CDNServices[serviceID]; exists {
		switch r.Method {
		case http.MethodGet:
			if err := cdnService.DownloadFile(w, r); err != nil {
				glog.Errorf("Error downloading file %s: %v", serviceID, err)

				WriteErrorResponse(w, http.StatusInternalServerError, "File Download Error", "An error occurred while downloading the file.")
			}
		case http.MethodPost:
			if err := cdnService.UploadFile(w, r); err != nil {
				glog.Errorf("Error creating file %s: %v", serviceID, err)

				WriteErrorResponse(w, http.StatusInternalServerError, "File Upload Error", "An error occurred while uploading the file.")
			}
		case http.MethodDelete:
			if err := cdnService.DeleteFile(w, r); err != nil {
				glog.Errorf("Error deleting file %s: %v", serviceID, err)

				WriteErrorResponse(w, http.StatusInternalServerError, "File Deletion Error", "An error occurred while deleting the file.")
			}
		default:
			WriteErrorResponse(w, http.StatusMethodNotAllowed, "Method Not Allowed", "The requested method is not supported for this service.")
		}
	} else {
		WriteErrorResponse(w, http.StatusNotFound, "Service Not Found", "The requested service does not exist.")
	}
}
