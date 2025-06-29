package server

import (
	"net/http"

	"go.uber.org/zap"

	"wednesday.wtf/godin/internal/api"
	"wednesday.wtf/godin/internal/config"
)

type Server struct {
	log    *zap.Logger
	config *config.Config
	mux    *http.ServeMux
	router *api.Router
}

func New(log *zap.Logger, config *config.Config) *Server {
	mux := http.NewServeMux()
	router := api.NewRouter(log, mux)

	router.RegisterHandlers()

	return &Server{
		log:    log,
		config: config,
		mux:    mux,
		router: router,
	}
}

func (s *Server) Start() error {
	return http.ListenAndServe(s.config.ServerAddress, s.mux)
}
