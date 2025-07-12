package discord

import (
	"github.com/bwmarrin/discordgo"
	"github.com/google/uuid"
	"strconv"
)

type Client struct {
	session *discordgo.Session
}

func New(botToken string) (*Client, error) {
	session, err := discordgo.New("Bot " + botToken)

	if err != nil {
		return nil, err
	}

	return &Client{
		session: session,
	}, nil
}

type WebhookData struct {
	ID    int
	Token string
}
type GuildEndpoints map[int][]*WebhookData

func (c *Client) InitializeGuild(guildID string, channelCount int, webhookCount int) (GuildEndpoints, error) {
	guild, err := c.session.Guild(guildID)

	if err != nil {
		return nil, err
	}

	channels, err := c.session.GuildChannels(guild.ID)

	if err != nil {
		return nil, err
	}

	for _, channel := range channels {
		if channel.Name == guild.Name {
			continue
		}

		if _, err := c.session.ChannelDelete(channel.ID); err != nil {
			return nil, err
		}
	}

	endpoints := make(map[int][]*WebhookData)

	for range channelCount {
		channelUUID := uuid.NewString()
		channel, err := c.session.GuildChannelCreate(guild.ID, channelUUID, discordgo.ChannelTypeGuildText)

		if err != nil {
			return nil, err
		}

		channelID, err := strconv.Atoi(channel.ID)

		if err != nil {
			return nil, err
		}

		endpoints[channelID] = make([]*WebhookData, webhookCount)

		for i := range webhookCount {
			webhookUUID := uuid.NewString()
			webhook, err := c.session.WebhookCreate(channel.ID, webhookUUID, "")

			if err != nil {
				return nil, err
			}

			webhookID, err := strconv.Atoi(webhook.ID)

			if err != nil {
				return nil, err
			}

			endpoints[channelID][i] = &WebhookData{
				ID:    webhookID,
				Token: webhook.Token,
			}
		}
	}

	return endpoints, nil
}
