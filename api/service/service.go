package service

import (
	"net/http"

	"wednesday.wtf/godin/api/discord"
)

type CDNService interface {
	Start()

	Stop()

	CreateSession(w http.ResponseWriter, r *http.Request) error

	GetSession(w http.ResponseWriter, r *http.Request) error

	DeleteSession(w http.ResponseWriter, r *http.Request) error

	UploadFile(w http.ResponseWriter, r *http.Request) error

	ListFiles(w http.ResponseWriter, r *http.Request) error

	DownloadFile(w http.ResponseWriter, r *http.Request) error

	DeleteFile(w http.ResponseWriter, r *http.Request) error

	CreateService(w http.ResponseWriter, r *http.Request) error

	GetService(w http.ResponseWriter, r *http.Request) error

	DeleteService(w http.ResponseWriter, r *http.Request) error
}

var CDNServices = map[string]CDNService{
	"eris": &discord.Service,
}
