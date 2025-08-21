package cdn

import "net/http"

type Service interface {
	CreateService(w http.ResponseWriter, r *http.Request) error
	GetService(w http.ResponseWriter, r *http.Request) error
	DeleteService(w http.ResponseWriter, r *http.Request) error
}

type File interface {
	CreateFile(w http.ResponseWriter, r *http.Request) error
	PutFile(w http.ResponseWriter, r *http.Request) error
	GetFile(w http.ResponseWriter, r *http.Request) error
	DeleteFile(w http.ResponseWriter, r *http.Request) error
}

type Session interface {
	CreateSession(w http.ResponseWriter, r *http.Request) error
	GetSession(w http.ResponseWriter, r *http.Request) error
	DeleteSession(w http.ResponseWriter, r *http.Request) error
}

type Handler interface {
	Service
	File
	Session
}
