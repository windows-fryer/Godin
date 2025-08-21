package eris

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"wednesday.wtf/godin/internal/database"
	"wednesday.wtf/godin/internal/discord"
	"wednesday.wtf/godin/pkg/responder"
)

type createServiceRequest struct {
	BotToken string `json:"bot_token"`

	GuildID           int `json:"guild_id"`
	GuildChannelCount int `json:"guild_channel_count"`
	GuildWebhookCount int `json:"guild_webhook_count"`
}

type createServiceResponse struct {
	ServiceID string `json:"service_id"`
}

func (h *Handler) guildExists(guildID int) bool {
	var found bool

	res := h.db.QueryRow(`SELECT EXISTS (SELECT 1 FROM godin.eris.guilds WHERE guild_id = $1)`, guildID)

	if err := res.Scan(&found); err != nil {
		return false
	}

	return found
}

func (h *Handler) serviceExists(id string) (bool, error) {
	var found bool

	res := h.db.QueryRow(`SELECT EXISTS (SELECT 1 FROM godin.eris.services WHERE service_id = $1)`, id)

	if err := res.Scan(&found); err != nil {
		return false, err
	}

	return found, nil
}

func (h *Handler) CreateService(w http.ResponseWriter, r *http.Request) error {
	request := createServiceRequest{}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return err
	}

	if ok := h.guildExists(request.GuildID); ok {
		return responder.NewError(http.StatusBadRequest, "guild already exists")
	}

	client, err := discord.New(request.BotToken)

	if err != nil {
		return err
	}

	endpoints, err := client.InitializeGuild(strconv.Itoa(request.GuildID), request.GuildChannelCount, request.GuildWebhookCount)

	if err != nil {
		return err
	}

	h.log.Debug("Endpoints Created", zap.Any("endpoints", endpoints))

	if _, err := h.db.Transaction(func(d *database.Database, tx *sql.Tx) (*sql.Result, error) {
		if _, err := tx.Exec(`INSERT INTO godin.eris.guilds (guild_id) VALUES ($1)`, request.GuildID); err != nil {
			return nil, err
		}

		return nil, nil
	}); err != nil {
		return err
	}

	serviceID := uuid.NewString()

	if _, err := h.db.Transaction(func(d *database.Database, tx *sql.Tx) (*sql.Result, error) {
		if _, err := tx.Exec(`INSERT INTO godin.eris.services (service_id, guild_id, bot_token) VALUES ($1, $2, $3)`, serviceID, request.GuildID, request.BotToken); err != nil {
			return nil, err
		}

		return nil, nil
	}); err != nil {
		return err
	}

	if _, err := h.db.Transaction(func(d *database.Database, tx *sql.Tx) (*sql.Result, error) {
		channelsStmt, err := tx.Prepare(`INSERT INTO godin.eris.channels (guild_id, channel_id) VALUES ($1, $2)`)

		if err != nil {
			return nil, err
		}

		webhooksStmt, err := tx.Prepare(`INSERT INTO godin.eris.webhooks (webhook_id, channel_id, webhook_token) VALUES ($1, $2, $3)`)

		if err != nil {
			return nil, err
		}

		for channelID, webhooks := range endpoints {
			if _, err := channelsStmt.Exec(request.GuildID, channelID); err != nil {
				return nil, err
			}

			for _, webhookData := range webhooks {
				if _, err = webhooksStmt.Exec(webhookData.ID, channelID, webhookData.Token); err != nil {
					return nil, err
				}
			}
		}

		return nil, nil
	}); err != nil {
		return err
	}

	return responder.Respond(w, http.StatusCreated, createServiceResponse{
		ServiceID: serviceID,
	})
}

func (h *Handler) GetService(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (h *Handler) DeleteService(w http.ResponseWriter, r *http.Request) error {
	return nil
}
