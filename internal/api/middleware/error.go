package middleware

import (
	"errors"
	"net/http"

	"go.uber.org/zap"
	"wednesday.wtf/godin/pkg/responder"
)

type Handler func(w http.ResponseWriter, r *http.Request) error

func Error(log *zap.Logger, next Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := next(w, r); err != nil {
			var customErr *responder.Error

			if errors.As(err, &customErr) {
				log.Debug("Request failed",
					zap.String("error", customErr.Message),
					zap.Int("status_code", customErr.Code),
					zap.String("method", r.Method),
					zap.String("url", r.URL.String()),
				)
			} else {
				log.Error("Request failed",
					zap.Error(err),
					zap.String("method", r.Method),
				)
			}

			responder.RespondError(w, err)
		}
	}
}
