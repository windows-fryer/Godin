package discord

import (
	"fmt"
	"io"
	"net/url"
	"strconv"

	"github.com/bwmarrin/discordgo"
	"github.com/google/uuid"
)

const ownedResourcePrefix = "godin-storage-"

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

	endpoints := make(map[int][]*WebhookData)

	for range channelCount {
		channelName := fmt.Sprintf("%s%s", ownedResourcePrefix, uuid.NewString())
		channel, err := c.session.GuildChannelCreate(guild.ID, channelName, discordgo.ChannelTypeGuildText)

		if err != nil {
			return nil, err
		}

		channelID, err := strconv.Atoi(channel.ID)

		if err != nil {
			return nil, err
		}

		endpoints[channelID] = make([]*WebhookData, webhookCount)

		for i := range webhookCount {
			webhookName := fmt.Sprintf("%s%s", ownedResourcePrefix, uuid.NewString())
			webhook, err := c.session.WebhookCreate(channel.ID, webhookName, "")

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

func (c *Client) GuildUploadSize(guildID int) (int, error) {
	guild, err := c.session.Guild(strconv.Itoa(guildID))

	if err != nil {
		return 0, err
	}

	switch guild.PremiumTier {
	case discordgo.PremiumTier3:
		return 100, nil
	case discordgo.PremiumTier2:
		return 50, nil
	default:
		return 10, nil
	}
}

type FileResponse struct {
	MessageID         string
	AttachmentExpires int
	AttachmentURL     string
	AttachmentSize    int
}

func (c *Client) UploadChunk(webhookID string, webhookToken string, r *io.Reader) (*FileResponse, error) {
	message, err := c.session.WebhookExecute(webhookID, webhookToken, true, &discordgo.WebhookParams{
		Files: []*discordgo.File{{
			Name:   uuid.NewString(),
			Reader: *r,
		}},
	})

	if err != nil {
		return nil, err
	}

	attachment := message.Attachments[0]
	attachmentUrl := attachment.URL

	attachmentUrlParsed, err := url.Parse(attachmentUrl)

	if err != nil {
		return nil, err
	}

	attachmentExpires, err := strconv.ParseInt(attachmentUrlParsed.Query().Get("ex"), 16, 64)

	if err != nil {
		return nil, err
	}

	return &FileResponse{
		MessageID:         message.ID,
		AttachmentExpires: int(attachmentExpires),
		AttachmentURL:     attachmentUrl,
		AttachmentSize:    attachment.Size,
	}, nil
}
