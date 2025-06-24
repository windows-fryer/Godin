package service

import (
	"net/http"

	"wednesday.wtf/godin/api/discord"
)

type CDNBase interface {
	Start()

	Stop()
}

type CDNService interface {
	CreateService(w http.ResponseWriter, r *http.Request) error

	GetService(w http.ResponseWriter, r *http.Request) error

	DeleteService(w http.ResponseWriter, r *http.Request) error
}

type CDNFile interface {
	UploadFile(w http.ResponseWriter, r *http.Request) error

	ListFiles(w http.ResponseWriter, r *http.Request) error

	DownloadFile(w http.ResponseWriter, r *http.Request) error

	DeleteFile(w http.ResponseWriter, r *http.Request) error
}

type CDNSession interface {
	CreateSession(w http.ResponseWriter, r *http.Request) error

	GetSession(w http.ResponseWriter, r *http.Request) error

	DeleteSession(w http.ResponseWriter, r *http.Request) error
}

type CDNAPI interface {
	CDNBase
	CDNService
	CDNFile
	CDNSession
}

var CDNServices = map[string]CDNAPI{
	"eris": &discord.Discord,
}
