package eris

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"wednesday.wtf/godin/internal/database"
	"wednesday.wtf/godin/pkg/responder"
)

type createSessionRequest struct {
	FileName string `json:"file_name"`
}

type createSessionResponse struct {
	SessionID     string `json:"session_id"`
	MaxUploadSize int    `json:"max_upload_size"`
	Expires       int    `json:"expires"`
}

func (h *Handler) CreateSession(w http.ResponseWriter, r *http.Request) error {
	serviceID := r.PathValue("service_id")
	if serviceID == "" {
		return responder.NewError(http.StatusBadRequest, "service_id is required")
	}

	exists, err := h.serviceExists(serviceID)
	if err != nil {
		return err
	}

	if !exists {
		return responder.NewError(http.StatusNotFound, "service not found")
	}

	request := createSessionRequest{}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return err
	}

	if request.FileName == "" {
		return responder.NewError(http.StatusBadRequest, "file_name is required")
	}

	maxUploadSize, err := h.maxUploadSize(serviceID)
	if err != nil {
		return err
	}

	fileID := uuid.NewString()
	sessionID := uuid.NewString()
	expirationTime := time.Now().Add(5 * time.Minute)

	if err := h.db.TransactionContext(r.Context(), func(d *database.Database, tx *sql.Tx) error {
		if _, err := tx.ExecContext(r.Context(), "INSERT INTO eris.files (file_id, service_id, file_name, status) VALUES ($1, $2, $3, 'pending')", fileID, serviceID, request.FileName); err != nil {
			return err
		}

		if _, err := tx.ExecContext(r.Context(), "INSERT INTO eris.sessions (session_id, service_id, file_id, file_chunk_size, expiration_time, status) VALUES ($1, $2, $3, $4, $5, 'open')", sessionID, serviceID, fileID, maxUploadSize, expirationTime); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return err
	}

	response := createSessionResponse{
		SessionID:     sessionID,
		MaxUploadSize: maxUploadSize,
		Expires:       int(expirationTime.Unix()),
	}

	h.log.Debug("Session Created", zap.Any("response", response))

	return responder.Respond(w, http.StatusCreated, response)
}

func (h *Handler) GetSession(w http.ResponseWriter, r *http.Request) error {
	return responder.NewError(http.StatusNotImplemented, "not implemented")
}

func (h *Handler) DeleteSession(w http.ResponseWriter, r *http.Request) error {
	return responder.NewError(http.StatusNotImplemented, "not implemented")
}
