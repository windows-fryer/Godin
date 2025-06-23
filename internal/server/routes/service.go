package routes

import (
	"net/http"
	"strings"

	"github.com/golang/glog"
	"wednesday.wtf/godin/api/service"
)

func getServiceID(r *http.Request) string {
	urlSplit := strings.Split(strings.Trim(r.URL.Path, "/"), "/")

	if len(urlSplit) < 3 {
		return ""
	}

	return urlSplit[2]
}

func ServiceHandler(w http.ResponseWriter, r *http.Request) {
	serviceID := getServiceID(r)

	if cdnService, exists := service.CDNServices[serviceID]; exists {
		switch r.Method {
		case http.MethodGet:
			if err := cdnService.GetService(w, r); err != nil {
				glog.Errorf("Error retrieving service %s: %v", serviceID, err)

				WriteErrorResponse(w, http.StatusInternalServerError, "Service Retrieval Error", "An error occurred while retrieving the service.")
			}
		case http.MethodPost:
			if err := cdnService.CreateService(w, r); err != nil {
				glog.Errorf("Error creating service %s: %v", serviceID, err)

				WriteErrorResponse(w, http.StatusInternalServerError, "Service Creation Error", "An error occurred while creating the service.")
			}
		case http.MethodDelete:
			if err := cdnService.DeleteService(w, r); err != nil {
				glog.Errorf("Error deleting service %s: %v", serviceID, err)

				WriteErrorResponse(w, http.StatusInternalServerError, "Service Deletion Error", "An error occurred while deleting the service.")
			}
		default:
			WriteErrorResponse(w, http.StatusMethodNotAllowed, "Method Not Allowed", "The requested method is not supported for this service.")
		}
	} else {
		WriteErrorResponse(w, http.StatusNotFound, "Service Not Found", "The requested service does not exist.")
	}
}
