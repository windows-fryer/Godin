package eris

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"wednesday.wtf/godin/internal/database"
	"wednesday.wtf/godin/pkg/responder"
)

func (h *Handler) getServiceID(sessionID string) (string, error) {
	res := h.db.QueryRow("SELECT service_id FROM eris.sessions WHERE session_id = $1", sessionID)

	var serviceID string
	if err := res.Scan(&serviceID); err != nil {
		return "", err
	}

	return serviceID, nil
}

func (h *Handler) sessionExists(id string) (bool, error) {
	var found bool

	res := h.db.QueryRow(`SELECT EXISTS (SELECT 1 FROM eris.sessions WHERE session_id = $1 AND status = 'open' AND expiration_time > NOW())`, id)

	if err := res.Scan(&found); err != nil {
		return false, err
	}

	return found, nil
}

type SessionMetadata struct {
	chunkSize      int
	fileID         string
	expirationTime time.Time
}

func (h *Handler) sessionMetadata(id string) (*SessionMetadata, error) {
	metadata := SessionMetadata{}

	res := h.db.QueryRow("SELECT file_chunk_size, file_id, expiration_time FROM eris.sessions WHERE session_id = $1 AND status = 'open'", id)

	if err := res.Scan(&metadata.chunkSize, &metadata.fileID, &metadata.expirationTime); err != nil {
		return nil, err
	}

	return &metadata, nil
}

func (h *Handler) CreateFile(w http.ResponseWriter, r *http.Request) error {
	sessionID := r.PathValue("session_id")
	if sessionID == "" {
		return responder.NewError(http.StatusBadRequest, "session_id is required")
	}

	sessionExists, err := h.sessionExists(sessionID)
	if err != nil {
		return err
	}

	if !sessionExists {
		return responder.NewError(http.StatusNotFound, "upload session not found or expired")
	}

	sessionMetadata, err := h.sessionMetadata(sessionID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return responder.NewError(http.StatusNotFound, "upload session not found or expired")
		}
		return err
	}

	serviceID, err := h.getServiceID(sessionID)
	if err != nil {
		return err
	}

	discordClient, err := h.getDiscordClient(serviceID)
	if err != nil {
		return err
	}

	var guildID string
	if err := h.db.QueryRowContext(r.Context(), "SELECT guild_id FROM eris.services WHERE service_id = $1", serviceID).Scan(&guildID); err != nil {
		return err
	}

	var channelID string
	if err := h.db.QueryRowContext(r.Context(), "SELECT channel_id FROM eris.channels WHERE guild_id = $1 AND status = 'active' ORDER BY RANDOM() LIMIT 1", guildID).Scan(&channelID); err != nil {
		return err
	}

	rows, err := h.db.QueryContext(
		r.Context(),
		"SELECT webhook_id, webhook_token FROM eris.webhooks WHERE channel_id = $1 AND status = 'active' ORDER BY RANDOM()",
		channelID,
	)
	if err != nil {
		return err
	}
	defer rows.Close()

	var webhooks [][2]string
	for rows.Next() {
		var webhookID, webhookToken string
		if err := rows.Scan(&webhookID, &webhookToken); err != nil {
			return err
		}
		webhooks = append(webhooks, [2]string{webhookID, webhookToken})
	}

	if err := rows.Err(); err != nil {
		return err
	}

	if len(webhooks) == 0 {
		return responder.NewError(http.StatusInternalServerError, fmt.Sprintf("no active webhooks found for channel %s", channelID))
	}

	const mbToBytes = 1048576
	const httpHeaderSize = 1024 // Conservative allowance for multipart overhead.

	chunkSize := int64(sessionMetadata.chunkSize*mbToBytes - httpHeaderSize)
	if chunkSize <= 0 {
		return responder.NewError(http.StatusInternalServerError, "invalid upload chunk size")
	}

	partIndex := -1
	if err := h.db.QueryRowContext(r.Context(), "SELECT COALESCE(MAX(part_index), -1) FROM eris.file_parts WHERE file_id = $1", sessionMetadata.fileID).Scan(&partIndex); err != nil {
		return err
	}
	partIndex++

	if err := h.db.TransactionContext(r.Context(), func(d *database.Database, tx *sql.Tx) error {
		_, err := tx.ExecContext(r.Context(), "UPDATE eris.files SET status = 'uploading', updated_at = NOW() WHERE file_id = $1 AND status IN ('pending', 'uploading')", sessionMetadata.fileID)
		return err
	}); err != nil {
		return err
	}

	uploadedParts := 0
	for {
		chunkReader := io.LimitReader(r.Body, chunkSize)
		firstByte := make([]byte, 1)

		n, err := chunkReader.Read(firstByte)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		fullChunkReader := io.MultiReader(bytes.NewReader(firstByte[:n]), chunkReader)

		webhook := webhooks[partIndex%len(webhooks)]
		webhookID, webhookToken := webhook[0], webhook[1]

		metadata, err := discordClient.client.UploadChunk(webhookID, webhookToken, &fullChunkReader)
		if err != nil {
			return err
		}

		currentPartIndex := partIndex
		partIndex++
		uploadedParts++

		if err = h.db.TransactionContext(r.Context(), func(d *database.Database, tx *sql.Tx) error {
			if _, err := tx.ExecContext(r.Context(), "INSERT INTO eris.file_parts (file_id, channel_id, message_id, message_url, message_expiration, part_index, size_bytes) VALUES ($1, $2, $3, $4, to_timestamp($5)::timestamptz, $6, $7)", sessionMetadata.fileID, channelID, metadata.MessageID, metadata.AttachmentURL, metadata.AttachmentExpires, currentPartIndex, metadata.AttachmentSize); err != nil {
				return err
			}

			if _, err := tx.ExecContext(r.Context(), "UPDATE eris.files SET uploaded_bytes = COALESCE(uploaded_bytes, 0) + $1, updated_at = NOW() WHERE file_id = $2", metadata.AttachmentSize, sessionMetadata.fileID); err != nil {
				return err
			}

			return nil
		}); err != nil {
			return err
		}
	}

	return responder.Respond(w, http.StatusCreated, map[string]any{
		"file_id":        sessionMetadata.fileID,
		"uploaded_parts": uploadedParts,
	})
}

func (h *Handler) GetFile(w http.ResponseWriter, r *http.Request) error {
	return responder.NewError(http.StatusNotImplemented, "not implemented")
}

func (h *Handler) DeleteFile(w http.ResponseWriter, r *http.Request) error {
	return responder.NewError(http.StatusNotImplemented, "not implemented")
}

func (h *Handler) PutFile(w http.ResponseWriter, r *http.Request) error {
	sessionID := r.PathValue("session_id")
	if sessionID == "" {
		return responder.NewError(http.StatusBadRequest, "session_id is required")
	}

	if err := h.db.TransactionContext(r.Context(), func(d *database.Database, tx *sql.Tx) error {
		var fileID string
		if err := tx.QueryRowContext(r.Context(), "SELECT file_id FROM eris.sessions WHERE session_id = $1 AND status = 'open' AND expiration_time > NOW() FOR UPDATE", sessionID).Scan(&fileID); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return responder.NewError(http.StatusNotFound, "upload session not found or expired")
			}
			return err
		}

		if _, err := tx.ExecContext(r.Context(), "UPDATE eris.files SET status = 'complete', completed_at = NOW(), updated_at = NOW() WHERE file_id = $1", fileID); err != nil {
			return err
		}

		if _, err := tx.ExecContext(r.Context(), "UPDATE eris.sessions SET status = 'complete', updated_at = NOW() WHERE session_id = $1", sessionID); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return err
	}

	return responder.Respond(w, http.StatusOK, map[string]string{"status": "complete"})
}
