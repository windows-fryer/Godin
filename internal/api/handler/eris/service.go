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

	res := h.db.QueryRow(`SELECT EXISTS (SELECT 1 FROM eris.guilds WHERE guild_id = $1)`, guildID)

	if err := res.Scan(&found); err != nil {
		return false
	}

	return found
}

func (h *Handler) serviceExists(id string) (bool, error) {
	var found bool

	res := h.db.QueryRow(`SELECT EXISTS (SELECT 1 FROM eris.services WHERE service_id = $1)`, id)

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

	if request.BotToken == "" {
		return responder.NewError(http.StatusBadRequest, "bot_token is required")
	}

	if request.GuildID == 0 {
		return responder.NewError(http.StatusBadRequest, "guild_id is required")
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

	serviceID := uuid.NewString()

	if err := h.db.TransactionContext(r.Context(), func(d *database.Database, tx *sql.Tx) error {
		if _, err := tx.ExecContext(r.Context(), `INSERT INTO eris.guilds (guild_id) VALUES ($1)`, request.GuildID); err != nil {
			return err
		}

		if _, err := tx.ExecContext(r.Context(), `INSERT INTO eris.services (service_id, guild_id, bot_token, status) VALUES ($1, $2, $3, 'active')`, serviceID, request.GuildID, request.BotToken); err != nil {
			return err
		}

		channelsStmt, err := tx.PrepareContext(r.Context(), `INSERT INTO eris.channels (guild_id, channel_id, status) VALUES ($1, $2, 'active')`)
		if err != nil {
			return err
		}
		defer channelsStmt.Close()

		webhooksStmt, err := tx.PrepareContext(r.Context(), `INSERT INTO eris.webhooks (webhook_id, channel_id, webhook_token, status) VALUES ($1, $2, $3, 'active')`)
		if err != nil {
			return err
		}
		defer webhooksStmt.Close()

		for channelID, webhooks := range endpoints {
			if _, err := channelsStmt.ExecContext(r.Context(), request.GuildID, channelID); err != nil {
				return err
			}

			for _, webhookData := range webhooks {
				if _, err = webhooksStmt.ExecContext(r.Context(), webhookData.ID, channelID, webhookData.Token); err != nil {
					return err
				}
			}
		}

		return nil
	}); err != nil {
		return err
	}

	return responder.Respond(w, http.StatusCreated, createServiceResponse{
		ServiceID: serviceID,
	})
}

func (h *Handler) GetService(w http.ResponseWriter, r *http.Request) error {
	return responder.NewError(http.StatusNotImplemented, "not implemented")
}

func (h *Handler) DeleteService(w http.ResponseWriter, r *http.Request) error {
	return responder.NewError(http.StatusNotImplemented, "not implemented")
}
