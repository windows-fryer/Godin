package routes

import (
	"net/http"

	"github.com/golang/glog"
	"wednesday.wtf/godin/api/service"
)

func SessionHandler(w http.ResponseWriter, r *http.Request) {
	serviceID := getServiceID(r)

	if cdnService, exists := service.CDNServices[serviceID]; exists {
		switch r.Method {
		case http.MethodPost:
			if err := cdnService.CreateSession(w, r); err != nil {
				glog.Errorf("Error creating session for service %s: %v", serviceID, err)

				WriteErrorResponse(w, http.StatusInternalServerError, "Session Creation Error", "An error occurred while creating the session.")
			}
		case http.MethodGet:
			if err := cdnService.GetSession(w, r); err != nil {
				glog.Errorf("Error retrieving session for service %s: %v", serviceID, err)

				WriteErrorResponse(w, http.StatusInternalServerError, "Session Retrieval Error", "An error occurred while retrieving the session.")
			}
		case http.MethodDelete:
			if err := cdnService.DeleteSession(w, r); err != nil {
				glog.Errorf("Error deleting session for service %s: %v", serviceID, err)

				WriteErrorResponse(w, http.StatusInternalServerError, "Session Deletion Error", "An error occurred while deleting the session.")
			}
		default:
			WriteErrorResponse(w, http.StatusMethodNotAllowed, "Method Not Allowed", "The requested method is not supported for this service.")
		}
	} else {
		WriteErrorResponse(w, http.StatusNotFound, "Service Not Found", "The requested service does not exist.")
	}
}
