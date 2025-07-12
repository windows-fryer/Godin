package eris

import (
	"go.uber.org/zap"
	"wednesday.wtf/godin/internal/database"
)

type Handler struct {
	log *zap.Logger
	db  *database.Database
}

func NewHandler(log *zap.Logger, db *database.Database) *Handler {
	return &Handler{
		log: log,
		db:  db,
	}
}
