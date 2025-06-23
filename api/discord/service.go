package discord

import (
	"net/http"

	"github.com/golang/glog"
)

func (service *DiscordAPIService) CreateService(w http.ResponseWriter, r *http.Request) error {
	glog.V(2).Info("CreateService endpoint used")
	return nil
}

func (service *DiscordAPIService) GetService(w http.ResponseWriter, r *http.Request) error {
	glog.V(2).Info("GetService endpoint used")
	return nil
}

func (service *DiscordAPIService) DeleteService(w http.ResponseWriter, r *http.Request) error {
	glog.V(2).Info("DeleteService endpoint used")
	return nil
}
