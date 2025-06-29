package eris

import "go.uber.org/zap"

type Handler struct {
	log *zap.Logger
}

func NewHandler(log *zap.Logger) *Handler {
	return &Handler{
		log: log,
	}
}
