package api

import (
	"fmt"
	"net/http"
	"strings"
	"wednesday.wtf/godin/internal/database"

	"go.uber.org/zap"
	"wednesday.wtf/godin/internal/api/handler/eris"
	"wednesday.wtf/godin/internal/api/middleware"
	"wednesday.wtf/godin/internal/cdn"
	"wednesday.wtf/godin/pkg/responder"
)

type Router struct {
	log      *zap.Logger
	mux      *http.ServeMux
	services map[string]cdn.Handler
	db       *database.Database
}

func (r *Router) handleService(w http.ResponseWriter, req *http.Request, h cdn.Service) error {
	switch req.Method {
	case http.MethodPost:
		r.log.Debug("Creating service", zap.String("url", req.URL.String()))

		return h.CreateService(w, req)
	case http.MethodGet:
		r.log.Debug("Getting service", zap.String("url", req.URL.String()))

		return h.GetService(w, req)
	case http.MethodDelete:
		r.log.Debug("Deleting service", zap.String("url", req.URL.String()))

		return h.DeleteService(w, req)
	default:
		return responder.NewError(http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (r *Router) handleSession(w http.ResponseWriter, req *http.Request, h cdn.Session) error {
	switch req.Method {
	case http.MethodPost:
		r.log.Debug("Creating session", zap.String("url", req.URL.String()))

		return h.CreateSession(w, req)
	case http.MethodGet:
		r.log.Debug("Getting session", zap.String("url", req.URL.String()))

		return h.GetSession(w, req)
	case http.MethodDelete:
		r.log.Debug("Deleting session", zap.String("url", req.URL.String()))

		return h.DeleteSession(w, req)
	default:
		return responder.NewError(http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (r *Router) handleFile(w http.ResponseWriter, req *http.Request, h cdn.File) error {
	switch req.Method {
	case http.MethodPost:
		r.log.Debug("Creating file", zap.String("url", req.URL.String()))

		return h.CreateFile(w, req)
	case http.MethodGet:
		r.log.Debug("Getting file", zap.String("url", req.URL.String()))

		return h.GetFile(w, req)
	case http.MethodDelete:
		r.log.Debug("Deleting file", zap.String("url", req.URL.String()))

		return h.DeleteFile(w, req)
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
			return r.handleService(w, req, handler)
		case "session":
			return r.handleSession(w, req, handler)
		case "file":
			return r.handleFile(w, req, handler)
		default:
			return responder.NewError(http.StatusNotFound, "not found")
		}
	}
}

func (r *Router) createHandler(routeType string) {
	r.mux.HandleFunc(fmt.Sprintf("/v1/%s/", routeType), middleware.Error(r.log, r.dispatch(routeType)))
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	responder.RespondError(w, responder.NewError(http.StatusNotFound, "not found"))
}

func (r *Router) RegisterHandlers() {
	r.mux.HandleFunc("/", notFoundHandler)

	r.createHandler("service")
	r.createHandler("session")
	r.createHandler("file")
}

func NewRouter(log *zap.Logger, db *database.Database, mux *http.ServeMux) *Router {
	services := map[string]cdn.Handler{
		"eris": eris.NewHandler(log, db),
	}

	return &Router{
		log:      log,
		db:       db,
		mux:      mux,
		services: services,
	}
}
