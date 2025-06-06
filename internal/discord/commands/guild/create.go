package guild

import (
	"strconv"

	"github.com/bwmarrin/discordgo"
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

func createChannels(s *discordgo.Session, guildID string, count int) error {
	createdChannels := make([]map[string]any, count)

	for i := range count {
		channel, err := s.GuildChannelCreate(guildID, "vfs-"+strconv.Itoa(i), discordgo.ChannelTypeGuildText)

		if err != nil {
			return err
		}

		webhook, err := s.WebhookCreate(channel.ID, "Godin VFS", "")

		if err != nil {
			return err
		}

		channelIDInt, err := strconv.Atoi(channel.ID)
		if err != nil {
			return err
		}
		webhookIDInt, err := strconv.Atoi(webhook.ID)
		if err != nil {
			return err
		}

		createdChannels[i] = map[string]any{
			"_id":           channelIDInt,
			"webhook_id":    webhookIDInt,
			"webhook_token": webhook.Token,
		}
	}

	if err := database.AddVFSDirectories(guildID, createdChannels); err != nil {
		return err
	}

	return nil
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

	if err := database.CreateGuild(i.GuildID); err != nil {
		commands.InteractionFailed(s, i, "Failed to create guild!")

		return err
	}

	if err := createChannels(s, i.GuildID, channelCount); err != nil {
		commands.InteractionFailed(s, i, "Failed to create channels!")
	}

	commands.InteractionSuccess(s, i, "Guild created successfully!")

	return nil
}
