package eris

import (
	"go.uber.org/zap"
	"wednesday.wtf/godin/internal/database"
	"wednesday.wtf/godin/internal/discord"
)

type Handler struct {
	log *zap.Logger
	db  *database.Database
}

func NewHandler(log *zap.Logger, db *database.Database) *Handler {
	return &Handler{
		log: log,
		db:  db,
	}
}

type DiscordClientData struct {
	client *discord.Client

	guild    int
	botToken string
}

func (h *Handler) getDiscordClient(serviceID string) (*DiscordClientData, error) {
	res := h.db.QueryRow("SELECT guild_id, bot_token FROM godin.eris.services WHERE service_id = $1", serviceID)

	var guildID int
	var botToken string

	if err := res.Scan(&guildID, &botToken); err != nil {
		return nil, err
	}

	client, err := discord.New(botToken)

	if err != nil {
		return nil, err
	}

	return &DiscordClientData{
		client: client,

		guild:    guildID,
		botToken: botToken,
	}, nil
}

func (h *Handler) maxUploadSize(serviceID string) (int, error) {
	data, err := h.getDiscordClient(serviceID)

	if err != nil {
		return 0, err
	}

	return data.client.GuildUploadSize(data.guild)
}
