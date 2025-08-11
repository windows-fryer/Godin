package eris

import (
	"database/sql"
	"encoding/json"
	"github.com/google/uuid"
	"net/http"
	"time"
	"wednesday.wtf/godin/internal/database"
	"wednesday.wtf/godin/internal/discord"

	"go.uber.org/zap"
	"wednesday.wtf/godin/pkg/responder"
	"wednesday.wtf/godin/pkg/splitutil"
)

func (h *Handler) maxUploadSize(serviceID string) (int, error) {
	res := h.db.QueryRow("SELECT guild_id, bot_token FROM godin.eris.services WHERE service_id = $1", serviceID)

	var guildID int
	var botToken string

	if err := res.Scan(&guildID, &botToken); err != nil {
		return 0, err
	}

	client, err := discord.New(botToken)

	if err != nil {
		return 0, err
	}

	return client.GuildUploadSize(guildID)
}

type createSessionRequest struct {
	FileName string `json:"file_name"`
}

type createSessionResponse struct {
	SessionID     string `json:"session_id"`
	MaxUploadSize int    `json:"max_upload_size"`
	Expires       int    `json:"expires"`
}

func (h *Handler) CreateSession(w http.ResponseWriter, r *http.Request) error {
	parsedURL, err := splitutil.SplitURL(r.URL.String(), []string{
		"service_id",
	}, 2)

	if err != nil {
		return err
	}

	serviceID := parsedURL["service_id"]

	request := createSessionRequest{}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return err
	}

	if request.FileName == "" {
		return responder.NewError(http.StatusBadRequest, "file_name is required")
	}

	fileID := uuid.NewString()

	if _, err := h.db.Transaction(func(d *database.Database, tx *sql.Tx) (*sql.Result, error) {
		if _, err := tx.Exec("INSERT INTO godin.eris.files (file_id, file_name) VALUES ($1, $2)", fileID, request.FileName); err != nil {
			return nil, err
		}

		return nil, nil
	}); err != nil {
		return err
	}

	sessionID := uuid.NewString()
	maxUploadSize, err := h.maxUploadSize(serviceID)
	expirationTime := time.Now().Add(5 * time.Minute)

	if err != nil {
		return err
	}

	if _, err := h.db.Transaction(func(d *database.Database, tx *sql.Tx) (*sql.Result, error) {
		if _, err := tx.Exec("INSERT INTO godin.eris.sessions (session_id, service_id, file_id, file_chunk_size) VALUES ($1, $2, $3, $4)", sessionID, serviceID, fileID, maxUploadSize); err != nil {
			return nil, err
		}

		return nil, nil
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
	return nil
}

func (h *Handler) DeleteSession(w http.ResponseWriter, r *http.Request) error {
	return nil
}
