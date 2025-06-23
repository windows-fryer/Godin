package discord

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/golang/glog"
	"github.com/google/uuid"
	"wednesday.wtf/godin/internal/database"
)

const (
	InsertGuildQuery   = "INSERT INTO eris.guilds (guild_id) VALUES ($1)"
	InsertServiceQuery = "INSERT INTO eris.services (service_id, bot_token, guild_id) VALUES ($1, $2, $3)"
)

type CreateServicePayload struct {
	GuildID  int    `json:"guild_id"`
	BotToken string `json:"bot_token"`
}

type CreateServiceResponse struct {
	ServiceID string `json:"service_id"`
}

func (service *DiscordAPIService) insertGuild(tx *sql.Tx, guildID int) error {
	_, err := tx.Exec(InsertGuildQuery, guildID)

	return err
}

func (service *DiscordAPIService) insertService(tx *sql.Tx, serviceID, botToken string, guildID int) error {
	_, err := tx.Exec(InsertServiceQuery, serviceID, botToken, guildID)

	return err
}

func writeJSONResponse(w http.ResponseWriter, status int, data any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	return json.NewEncoder(w).Encode(data)
}

func (service *DiscordAPIService) CreateService(w http.ResponseWriter, r *http.Request) error {
	if r.Header.Get("Content-Type") != "application/json" {
		return errors.New("invalid content type, expected application/json")
	}

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

	serviceID := uuid.NewString()

	tx, err := database.PostgresClient.Begin()

	if err != nil {
		return err
	}

	defer tx.Rollback()

	if err := service.insertGuild(tx, payload.GuildID); err != nil {
		return err
	}

	if err := service.insertService(tx, serviceID, payload.BotToken, payload.GuildID); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	response := CreateServiceResponse{
		ServiceID: serviceID,
	}

	if err := writeJSONResponse(w, http.StatusCreated, response); err != nil {
		glog.Errorf("Failed to write JSON response: %v", err)
	}

	return nil
}

func (service *DiscordAPIService) GetService(w http.ResponseWriter, r *http.Request) error {
	glog.V(2).Info("GetService endpoint used")
	return nil
}

func (service *DiscordAPIService) DeleteService(w http.ResponseWriter, r *http.Request) error {
	glog.V(2).Info("DeleteService endpoint used")
	return nil
}
