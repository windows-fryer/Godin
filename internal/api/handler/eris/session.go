package eris

import (
	"net/http"

	"go.uber.org/zap"
	"wednesday.wtf/godin/pkg/splitutil"
)

func (h *Handler) CreateSession(w http.ResponseWriter, r *http.Request) error {
	parsedURL, err := splitutil.SplitURL(r.URL.String(), []string{
		"service_id",
	}, 2)

	if err != nil {
		return err
	}

	serverID := parsedURL["service_id"]

	h.log.Debug("Creating Session", zap.String("server_id", serverID))

	return nil
}

func (h *Handler) GetSession(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (h *Handler) DeleteSession(w http.ResponseWriter, r *http.Request) error {
	return nil
}
