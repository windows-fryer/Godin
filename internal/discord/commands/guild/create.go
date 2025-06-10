package guild

import (
	"strconv"

	"github.com/bwmarrin/discordgo"
	"github.com/charmbracelet/log"
	"github.com/godin/internal/database"
	"github.com/godin/internal/discord/commands"
)

var CreateGuildCommand = &discordgo.ApplicationCommand{
	Name:        "guild",
	Description: "Manage guilds",
	Type:        discordgo.ChatApplicationCommand,

	Options: []*discordgo.ApplicationCommandOption{{
		Name:        "create",
		Description: "Create a new guild",
		Type:        discordgo.ApplicationCommandOptionSubCommand,

		Options: []*discordgo.ApplicationCommandOption{{
			Name:        "channel_count",
			Description: "Number of channels to create",
			Type:        discordgo.ApplicationCommandOptionInteger,
			Required:    false,
		}, {
			Name:        "webhook_count",
			Description: "Number of webhooks to create in each channel",
			Type:        discordgo.ApplicationCommandOptionInteger,
			Required:    false,
		}},
	}},
}

func deleteChannels(s *discordgo.Session, guildID string) error {
	channels, err := s.GuildChannels(guildID)

	if err != nil {
		return err
	}

	for _, channel := range channels {
		if channel.Name != "test" {
			if _, err := s.ChannelDelete(channel.ID); err != nil {
				return err
			}
		}
	}

	return nil
}

func createChannels(s *discordgo.Session, guildID string, channelCount int, webhookCount int) ([]database.GuildChannel, error) {
	createdChannels := make([]database.GuildChannel, channelCount*webhookCount)

	for i := range channelCount {
		channel, err := s.GuildChannelCreate(guildID, "vfs-"+strconv.Itoa(i), discordgo.ChannelTypeGuildText)

		if err != nil {
			return nil, err
		}

		for j := range webhookCount {
			webhook, err := s.WebhookCreate(channel.ID, "Godin VFS", "")

			if err != nil {
				return nil, err
			}

			channelIDInt, err := strconv.Atoi(channel.ID)

			if err != nil {
				return nil, err
			}

			webhookIDInt, err := strconv.Atoi(webhook.ID)

			if err != nil {
				return nil, err
			}

			createdChannels[i*webhookCount+j] = database.GuildChannel{
				ChannelID:    int64(channelIDInt),
				WebhookID:    int64(webhookIDInt),
				WebhookToken: webhook.Token,
			}
		}
	}

	return createdChannels, nil
}

func CreateGuild(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	if err := commands.InteractionDefer(s, i); err != nil {
		return commands.InteractionFailed(s, i, "Failed to defer interaction: "+err.Error())
	}

	if err := deleteChannels(s, i.GuildID); err != nil {
		commands.InteractionFailed(s, i, "Failed to delete channels!")
	}

	channelCount := 10

	if option := i.ApplicationCommandData().GetOption("create").GetOption("channel_count"); option != nil {
		channelCount = int(option.IntValue())
	}

	webhookCount := 5

	if option := i.ApplicationCommandData().GetOption("create").GetOption("webhook_count"); option != nil {
		webhookCount = int(option.IntValue())
	}

	channels, err := createChannels(s, i.GuildID, channelCount, webhookCount)

	if err != nil {
		commands.InteractionFailed(s, i, "Failed to create channels and webhooks: "+err.Error())

		return err
	}

	guildIDToInt, _ := strconv.Atoi(i.GuildID)

	log.Info("Creating new guild", "channels", channels)

	newGuild := &database.Guild{
		GuildID:     guildIDToInt,
		VFSChannels: channels,
	}

	if err := database.CreateGuild(newGuild); err != nil {
		commands.InteractionFailed(s, i, "Failed to create guild!")

		return err
	}

	commands.InteractionSuccess(s, i, "Guild created successfully!")

	return nil
}
