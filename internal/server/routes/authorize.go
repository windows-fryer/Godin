package routes

import (
	"net/http"

	"github.com/golang/glog"
)

func getServiceID(r *http.Request) string {
	path := r.URL.Path[len("/v1/authorize/"):]

	return path
}

func AuthorizeHandler(w http.ResponseWriter, r *http.Request) {
	serviceID := getServiceID(r)

	if serviceID == "" {
		WriteErrorResponse(w, http.StatusBadRequest, "Invalid Service ID", "The service ID is missing or invalid.")

		return
	}

	glog.V(2).Infof("Authorize request for service ID: %s", serviceID)

	// sessionToken, err := database.CreateSession(serviceID)

	// if err != nil {
	// 	glog.Errorf("Failed to create session for service ID %s: %v", serviceID, err)

	// 	WriteErrorResponse(w, http.StatusInternalServerError, "Internal Server Error", "Failed to create session.")

	// 	return
	// }
}
