package eris

import (
	"bytes"
	"database/sql"
	"fmt"
	"io"
	"net/http"

	"wednesday.wtf/godin/internal/database"
	"wednesday.wtf/godin/pkg/responder"
	"wednesday.wtf/godin/pkg/splitutil"
)

func (h *Handler) getServiceID(sessionID string) (string, error) {
	res, err := h.db.Query("SELECT service_id FROM godin.eris.sessions WHERE session_id = $1", sessionID)

	if err != nil {
		return "", err
	}

	defer func(res *sql.Rows) {
		err := res.Close()

		if err != nil {
			panic(err)
		}
	}(res)

	var serviceID string

	res.Next()

	err = res.Scan(&serviceID)

	if err != nil {
		return "", err
	}

	return serviceID, nil
}

func (h *Handler) sessionExists(id string) (bool, error) {
	var found bool

	res := h.db.QueryRow(`SELECT EXISTS (SELECT 1 FROM godin.eris.sessions WHERE session_id = $1)`, id)

	if err := res.Scan(&found); err != nil {
		return false, err
	}

	return found, nil
}

type SessionMetadata struct {
	chunkSize int
	fileID    string
}

func (h *Handler) sessionMetadata(id string) (*SessionMetadata, error) {
	metadata := SessionMetadata{}

	res := h.db.QueryRow("SELECT file_chunk_size, file_id FROM godin.eris.sessions WHERE session_id = $1", id)

	if err := res.Scan(&metadata.chunkSize, &metadata.fileID); err != nil {
		return nil, err
	}

	return &metadata, nil
}

func (h *Handler) CreateFile(w http.ResponseWriter, r *http.Request) error {
	parsedURL, err := splitutil.SplitURL(r.URL.String(), []string{
		"session_id",
	}, 3)

	if err != nil {
		return err
	}

	sessionID := parsedURL["session_id"]

	sessionExists, err := h.sessionExists(sessionID)

	if err != nil {
		return err
	}

	if !sessionExists {
		return responder.NewError(401, "Invalid session ID")
	}

	sessionMetadata, err := h.sessionMetadata(sessionID)

	if err != nil {
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

	query, err := h.db.Query("SELECT guild_id FROM godin.eris.services WHERE service_id = $1 ORDER BY RANDOM() LIMIT 1", serviceID)

	if err != nil {
		return err
	}

	query.Next()

	var guildID string

	if err = query.Scan(&guildID); err != nil {
		return err
	}

	if err = query.Close(); err != nil {
		return err
	}

	query, err = h.db.Query("SELECT channel_id FROM godin.eris.channels WHERE guild_id = $1 ORDER BY RANDOM() LIMIT 1", guildID)

	if err != nil {
		return err
	}

	query.Next()

	var channelID string

	if err = query.Scan(&channelID); err != nil {
		return err
	}

	if err = query.Close(); err != nil {
		return err
	}

	rows, err := h.db.Query(
		"SELECT webhook_id, webhook_token FROM godin.eris.webhooks WHERE channel_id = $1 ORDER BY RANDOM()",
		channelID,
	)

	if err != nil {
		return err
	}

	defer func(rows *sql.Rows) {
		err := rows.Close()

		if err != nil {
			panic(err)
		}
	}(rows)

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
		return responder.NewError(500, fmt.Sprintf("No webhooks found for channel %s", channelID))
	}

	const MbToBytes = 1048576
	const HttpHeaderSize = 1024 // rough guess, 1KB lost should be fine...

	chunkSize := int64(sessionMetadata.chunkSize*MbToBytes - HttpHeaderSize)

	var partIndex int

	res, err := h.db.Query("SELECT COALESCE(MAX(part_index), 0) FROM godin.eris.file_parts WHERE file_id = $1", sessionMetadata.fileID)

	if err != nil {
		return err
	}

	defer func(res *sql.Rows) {
		err := res.Close()

		if err != nil {
			panic(err)
		}
	}(res)

	res.Next()

	if err := res.Scan(&partIndex); err != nil {
		return err
	}

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

		partIndex++

		// TODO: Extract this outside of the loop and push all file parts at once; We don't need to do a transaction for each part

		if _, err = h.db.Transaction(func(d *database.Database, tx *sql.Tx) (*sql.Result, error) {
			if _, err := tx.Exec("INSERT INTO godin.eris.file_parts (file_id, channel_id, message_id, message_url, message_expiration, part_index) VALUES ($1, $2, $3, $4, to_timestamp($5)::timestamptz, $6)", sessionMetadata.fileID, channelID, metadata.MessageID, metadata.AttachmentURL, metadata.AttachmentExpires, partIndex); err != nil {
				return nil, err
			}

			return nil, nil
		}); err != nil {
			return err
		}
	}

	return nil
}

func (h *Handler) GetFile(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (h *Handler) DeleteFile(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (h *Handler) PutFile(w http.ResponseWriter, r *http.Request) error {
	return nil
}
