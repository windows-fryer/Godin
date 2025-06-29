package api

import (
	"net/http"
	"strings"

	"go.uber.org/zap"
	"wednesday.wtf/godin/internal/api/handler/eris"
	"wednesday.wtf/godin/internal/api/middleware"
	"wednesday.wtf/godin/internal/cdn"
	"wednesday.wtf/godin/pkg/responder"
)

type Router struct {
	log      *zap.Logger
	services map[string]cdn.Handler
}

func handleService(w http.ResponseWriter, r *http.Request, h cdn.Service) error {
	switch r.Method {
	case http.MethodPost:
		return h.CreateService(w, r)
	case http.MethodGet:
		return h.GetService(w, r)
	case http.MethodDelete:
		return h.DeleteService(w, r)
	default:
		return responder.NewError(http.StatusMethodNotAllowed, "method not allowed")
	}
}

func handleSession(w http.ResponseWriter, r *http.Request, h cdn.Session) error {
	switch r.Method {
	case http.MethodPost:
		return h.CreateSession(w, r)
	case http.MethodGet:
		return h.GetSession(w, r)
	case http.MethodDelete:
		return h.DeleteSession(w, r)
	default:
		return responder.NewError(http.StatusMethodNotAllowed, "method not allowed")
	}
}

func handleFile(w http.ResponseWriter, r *http.Request, h cdn.File) error {
	switch r.Method {
	case http.MethodPost:
		return h.CreateFile(w, r)
	case http.MethodGet:
		return h.GetFile(w, r)
	case http.MethodDelete:
		return h.DeleteFile(w, r)
	default:
		return responder.NewError(http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (r *Router) dispatch(routeType string) middleware.Handler {
	return func(w http.ResponseWriter, req *http.Request) error {
		parts := strings.Split(strings.Trim(req.URL.Path, "/"), "/")

		if len(parts) < 3 {
			return responder.NewError(http.StatusBadRequest, "invalid URL path")
		}

		serviceName := parts[2]

		handler, exists := r.services[serviceName]

		if !exists {
			return responder.NewError(http.StatusNotFound, "not found")
		}

		switch routeType {
		case "service":
			return handleService(w, req, handler)
		case "session":
			return handleSession(w, req, handler)
		case "file":
			return handleFile(w, req, handler)
		default:
			return responder.NewError(http.StatusNotFound, "not found")
		}
	}
}

func (r *Router) RegisterHandlers(mux *http.ServeMux) {
	mux.HandleFunc("/v1/service/", middleware.Error(r.log, r.dispatch("service")))
	mux.HandleFunc("/v1/session/", middleware.Error(r.log, r.dispatch("session")))
	mux.HandleFunc("/v1/file/", middleware.Error(r.log, r.dispatch("file")))
}

func NewRouter(log *zap.Logger) *Router {
	services := map[string]cdn.Handler{
		"eris": eris.NewHandler(log),
	}

	return &Router{
		log:      log,
		services: services,
	}
}
