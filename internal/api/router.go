package api

import (
	"net/http"

	"go.uber.org/zap"
	"wednesday.wtf/godin/internal/api/handler/eris"
	"wednesday.wtf/godin/internal/api/middleware"
	"wednesday.wtf/godin/internal/cdn"
	"wednesday.wtf/godin/internal/database"
	"wednesday.wtf/godin/pkg/responder"
)

type Router struct {
	log      *zap.Logger
	mux      *http.ServeMux
	services map[string]cdn.Handler
	db       *database.Database
}

func (r *Router) serviceHandler(serviceName string) (cdn.Handler, error) {
	handler, exists := r.services[serviceName]
	if !exists {
		return nil, responder.NewError(http.StatusNotFound, "service provider not found")
	}

	return handler, nil
}

func (r *Router) handleService(w http.ResponseWriter, req *http.Request) error {
	h, err := r.serviceHandler(req.PathValue("provider"))
	if err != nil {
		return err
	}

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

func (r *Router) handleSession(w http.ResponseWriter, req *http.Request) error {
	h, err := r.serviceHandler(req.PathValue("provider"))
	if err != nil {
		return err
	}

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

func (r *Router) handleFile(w http.ResponseWriter, req *http.Request) error {
	h, err := r.serviceHandler(req.PathValue("provider"))
	if err != nil {
		return err
	}

	switch req.Method {
	case http.MethodPost:
		r.log.Debug("Creating file", zap.String("url", req.URL.String()))
		return h.CreateFile(w, req)
	case http.MethodPut:
		r.log.Debug("Completing file", zap.String("url", req.URL.String()))
		return h.PutFile(w, req)
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

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	responder.RespondError(w, responder.NewError(http.StatusNotFound, "not found"))
}

func (r *Router) RegisterHandlers() {
	r.mux.HandleFunc("/", notFoundHandler)
	r.mux.HandleFunc("/v1/service/{provider}/", middleware.Error(r.log, r.handleService))
	r.mux.HandleFunc("/v1/session/{provider}/{service_id}/", middleware.Error(r.log, r.handleSession))
	r.mux.HandleFunc("/v1/file/{provider}/{session_id}/", middleware.Error(r.log, r.handleFile))
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
