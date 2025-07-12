package server

import (
	"net/http"
	"wednesday.wtf/godin/internal/database"

	"go.uber.org/zap"

	"wednesday.wtf/godin/internal/api"
	"wednesday.wtf/godin/internal/config"
)

type Server struct {
	log    *zap.Logger
	db     *database.Database
	config *config.Config
	mux    *http.ServeMux
	router *api.Router
}

func New(log *zap.Logger, db *database.Database, config *config.Config) *Server {
	mux := http.NewServeMux()
	router := api.NewRouter(log, db, mux)

	router.RegisterHandlers()

	return &Server{
		log:    log,
		config: config,
		db:     db,
		mux:    mux,
		router: router,
	}
}

func (s *Server) Start() error {
	return http.ListenAndServe(s.config.ServerAddress, s.mux)
}
