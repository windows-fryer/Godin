package discord

import (
	"net/http"

	"github.com/golang/glog"
)

func (service *DiscordAPIService) CreateSession(w http.ResponseWriter, r *http.Request) error {
	glog.V(2).Info("CreateSession endpoint used")
	return nil
}

func (service *DiscordAPIService) GetSession(w http.ResponseWriter, r *http.Request) error {
	glog.V(2).Info("GetSession endpoint used")
	return nil
}

func (service *DiscordAPIService) DeleteSession(w http.ResponseWriter, r *http.Request) error {
	glog.V(2).Info("DeleteSession endpoint used")
	return nil
}
