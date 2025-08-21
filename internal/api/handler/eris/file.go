package eris

import (
	"database/sql"
	"io"
	"net/http"

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

	res := h.db.QueryRow(`SELECT file_chunk_size, file_id FROM godin.eris.sessions WHERE session_id = $1`, id)

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

	//sessionMetadata, err := h.sessionMetadata(sessionID)

	//if err != nil {
	//	return err
	//}

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

	query, err = h.db.Query("SELECT webhook_id, webhook_token FROM godin.eris.webhooks WHERE channel_id = $1 ORDER BY RANDOM() LIMIT 1", channelID)

	if err != nil {
		return err
	}

	query.Next()

	var webhookID string
	var webhookToken string

	if err = query.Scan(&webhookID, &webhookToken); err != nil {
		return err
	}

	if err = query.Close(); err != nil {
		return err
	}

	body := io.Reader(r.Body)

	_, err = discordClient.client.UploadChunk(webhookID, webhookToken, &body)

	if err != nil {
		return err
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
