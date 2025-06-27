package session

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"wednesday.wtf/godin/internal/database"
	"wednesday.wtf/godin/internal/resource"
)

const (
	ExistsServiceQuery = "SELECT EXISTS(SELECT 1 FROM eris.services WHERE service_id = $1)"

	InsertFileQuery    = "INSERT INTO eris.files (file_id, file_name, file_size) VALUES ($1, $2, $3)"
	InsertSessionQuery = "INSERT INTO eris.sessions (session_id, file_id, service_id) VALUES ($1, $2, $3)"
)

type CreateSessionPayload struct {
	FileName string `json:"file_name"`
	FileSize int64  `json:"file_size"`
}

type CreateSessionResponse struct {
	SessionID    string `json:"session_id"`
	MaxChunkSize int64  `json:"max_chunk_size"`
}

func serviceExists(tx *sql.Tx, serviceID *uuid.UUID) (bool, error) {
	var exists bool

	if err := tx.QueryRow(ExistsServiceQuery, serviceID).Scan(&exists); err != nil {
		return false, err
	}

	return exists, nil
}

func createFile(tx *sql.Tx, payload *CreateSessionPayload) (string, error) {
	fileID := uuid.NewString()

	if _, err := tx.Exec(InsertFileQuery, fileID, payload.FileName, payload.FileSize); err != nil {
		return "", err
	}

	return fileID, nil
}

func createSession(tx *sql.Tx, payload *CreateSessionPayload, serviceID *uuid.UUID) (string, error) {
	fileID, err := createFile(tx, payload)

	if err != nil {
		return "", err
	}

	sessionID := uuid.NewString()

	if _, err := tx.Exec(InsertSessionQuery, sessionID, fileID, serviceID); err != nil {
		return "", err
	}

	return sessionID, nil
}

func (service *DiscordAPISession) CreateSession(w http.ResponseWriter, r *http.Request) error {
	serviceID, err := uuid.Parse(r.URL.String()[len("/v1/session/eris/"):])

	if err != nil {
		return err
	}

	tx, err := database.PostgresClient.BeginTx(r.Context(), nil)

	if err != nil {
		return err
	}

	defer tx.Rollback()

	exists, err := serviceExists(tx, &serviceID)

	if err != nil {
		return err
	}

	if !exists {
		return errors.New("service does not exist")
	}

	var payload CreateSessionPayload

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		return err
	}

	sessionID, err := createSession(tx, &payload, &serviceID)

	if err != nil {
		return err
	}

	// TODO: Get max chunk size from service!
	response := CreateSessionResponse{
		SessionID:    sessionID,
		MaxChunkSize: 1024 * 1024 * 10,
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return resource.GenerateJSONResponse(w, http.StatusCreated, response)
}
