package discord

import (
	"net/http"

	"github.com/golang/glog"
)

func (service *DiscordAPIService) UploadFile(w http.ResponseWriter, r *http.Request) error {
	glog.V(2).Info("UploadFile endpoint used")
	return nil
}

func (service *DiscordAPIService) ListFiles(w http.ResponseWriter, r *http.Request) error {
	glog.V(2).Info("ListFiles endpoint used")
	return nil
}

func (service *DiscordAPIService) DownloadFile(w http.ResponseWriter, r *http.Request) error {
	glog.V(2).Info("DownloadFile endpoint used")
	return nil
}

func (service *DiscordAPIService) DeleteFile(w http.ResponseWriter, r *http.Request) error {
	glog.V(2).Info("DeleteFile endpoint used")
	return nil
}
