package service

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/bwmarrin/discordgo"
	"github.com/google/uuid"
	"wednesday.wtf/godin/internal/database"
	"wednesday.wtf/godin/internal/resource"
)

const (
	InsertGuildQuery   = "INSERT INTO eris.guilds (guild_id) VALUES ($1)"
	InsertServiceQuery = "INSERT INTO eris.services (service_id, bot_token, guild_id) VALUES ($1, $2, $3)"
	InsertChannelQuery = "INSERT INTO eris.channels (channel_id, guild_id) VALUES ($1, $2)"
	InsertWebhookQuery = "INSERT INTO eris.webhooks (webhook_id, channel_id, webhook_token) VALUES ($1, $2, $3)"
)

type CreateServicePayload struct {
	GuildID  int    `json:"guild_id"`
	BotToken string `json:"bot_token"`

	ChannelCount int `json:"channel_count,omitempty"`
	WebhookCount int `json:"webhook_count,omitempty"`
}

type CreateServiceResponse struct {
	ServiceID string `json:"service_id"`
}

func deleteChannels(discordClient *discordgo.Session, guildID string) error {
	channels, err := discordClient.GuildChannels(guildID)

	if err != nil {
		return err
	}

	for _, channel := range channels {
		if channel.Name == "general" {
			continue
		}

		if _, err := discordClient.ChannelDelete(channel.ID); err != nil {
			return err
		}
	}

	return nil
}

func createChannels(tx *sql.Tx, discordClient *discordgo.Session, guildID string, channelCount int, webhookCount int) error {
	for range channelCount {
		channel, err := discordClient.GuildChannelCreate(guildID, resource.GenerateResourceName("lower", 2), discordgo.ChannelTypeGuildText)

		if err != nil {
			return err
		}

		if _, err := tx.Exec(InsertChannelQuery, channel.ID, guildID); err != nil {
			return err
		}

		for range webhookCount {
			webhook, err := discordClient.WebhookCreate(channel.ID, resource.GenerateResourceName("lower", 2), "")

			if err != nil {
				return err
			}

			if _, err := tx.Exec(InsertWebhookQuery, webhook.ID, channel.ID, webhook.Token); err != nil {
				return err
			}
		}
	}

	return nil
}

func createGuild(tx *sql.Tx, payload CreateServicePayload, serviceID string) error {
	if _, err := tx.Exec(InsertGuildQuery, payload.GuildID); err != nil {
		return err
	}

	channelCount := payload.ChannelCount
	webhookCount := payload.WebhookCount

	if payload.ChannelCount <= 0 {
		channelCount = 10
	}

	if payload.WebhookCount <= 0 {
		webhookCount = 1
	}

	discordClient, err := discordgo.New("Bot " + payload.BotToken)

	if err != nil {
		return err
	}

	guildIDStr := strconv.Itoa(payload.GuildID)

	deleteChannels(discordClient, guildIDStr)

	if err := createChannels(tx, discordClient, guildIDStr, channelCount, webhookCount); err != nil {
		return err
	}

	if _, err := tx.Exec(InsertServiceQuery, serviceID, payload.BotToken, payload.GuildID); err != nil {
		return err
	}

	return nil
}

func (service *DiscordAPIService) CreateService(w http.ResponseWriter, r *http.Request) error {
	var payload CreateServicePayload

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		return err
	}

	if payload.GuildID <= 0 {
		return errors.New("invalid guild_id")
	}

	if payload.BotToken == "" {
		return errors.New("bot_token is required")
	}

	tx, err := database.PostgresClient.BeginTx(r.Context(), nil)

	if err != nil {
		return err
	}

	defer tx.Rollback()

	serviceID := uuid.NewString()

	if err := createGuild(tx, payload, serviceID); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	response := CreateServiceResponse{
		ServiceID: serviceID,
	}

	return resource.GenerateJSONResponse(w, http.StatusCreated, response)
}
