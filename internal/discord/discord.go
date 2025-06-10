package discord

import (
	"maps"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/charmbracelet/log"
	"github.com/godin/internal/database"
	"github.com/godin/internal/discord/commands/guild"
)

var DiscordClient *discordgo.Session

func handleSlashCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.ApplicationCommandData().Name {
	case "guild":
		switch i.ApplicationCommandData().Options[0].Name {
		case "create":
			if err := guild.CreateGuild(s, i); err != nil {
				log.Error("Error handling guild create command: ", err)
			}
		}
	case "file":
		switch i.ApplicationCommandData().Options[0].Name {
		case "upload":
			if err := guild.UploadFile(s, i); err != nil {
				log.Error("Error handling file upload command: ", err)
			}
		}
	default:
		log.Warn("Unknown command: ", i.ApplicationCommandData().Options[0].Name)
	}
}

func initializeSlashCommands(discordSession *discordgo.Session) error {
	commands := []*discordgo.ApplicationCommand{
		guild.CreateGuildCommand,
		guild.UploadFileCommand,
	}

	for _, cmd := range commands {
		if _, err := discordSession.ApplicationCommandCreate(discordSession.State.User.ID, "", cmd); err != nil {
			return err
		}
	}

	discordSession.AddHandler(handleSlashCommand)

	return nil
}

func connectToDiscord() (*discordgo.Session, error) {
	discordSession, err := discordgo.New("Bot " + os.Getenv("DISCORD_TOKEN"))

	if err != nil {
		return nil, err
	}

	discordSession.Identify.Intents = discordgo.IntentsAllWithoutPrivileged | discordgo.IntentMessageContent
	discordSession.Client.Timeout = 0 // Disable timeout for long-running operations, like file uploads (lol)

	err = discordSession.Open()

	if err != nil {
		return nil, err
	}

	return discordSession, nil
}

func UpdateMessageAttachments(fileID string) error {
	file, err := database.GetVFSFile(fileID)

	if err != nil {
		return err
	}

	if file == nil {
		return nil
	}

	channelSortedParts := make(map[int64]map[int64]*database.VFSFilePart)
	latestChannelMessages := make(map[int64]*database.VFSFilePart)

	for _, part := range file.FileParts {
		if time.Now().Unix()-part.PartTimestamp < 60*60*24 {
			continue
		}

		if _, exists := channelSortedParts[part.PartChannelID]; !exists {
			channelSortedParts[part.PartChannelID] = make(map[int64]*database.VFSFilePart)
		}

		channelSortedParts[part.PartChannelID][part.PartID] = &part

		if latest, exists := latestChannelMessages[part.PartChannelID]; !exists || part.PartTimestamp > latest.PartTimestamp {
			latestChannelMessages[part.PartChannelID] = &part
		}
	}

	for channelID, parts := range channelSortedParts {
		channelIDStr := strconv.Itoa(int(channelID))

		remainingParts := make(map[int64]*database.VFSFilePart)

		maps.Copy(remainingParts, parts)

		var beforeID string

		for len(remainingParts) > 0 {
			messages, err := DiscordClient.ChannelMessages(channelIDStr, 100, beforeID, "", "")

			if err != nil {
				log.Error("Failed to fetch messages for channel", "channel_id", channelID, "error", err)

				break
			}

			if len(messages) == 0 {
				log.Warn("No more messages found in channel, but some parts were not updated", "channel_id", channelID, "remaining_parts", len(remainingParts))

				break
			}

			for _, msg := range messages {
				IDInt, err := strconv.Atoi(msg.ID)
				if err != nil {
					log.Error("Failed to convert message ID to int", "message_id", msg.ID, "error", err)

					continue
				}

				partMessage, exists := remainingParts[int64(IDInt)]

				if !exists {
					continue
				}

				if len(msg.Attachments) == 0 {
					log.Warn("No attachments found for message", "message_id", msg.ID, "channel_id", channelID)
					delete(remainingParts, int64(IDInt))

					continue
				}

				messageAttachment := msg.Attachments[0]

				if messageAttachment.URL != "" {
					partMessage.PartAttachmentURL = strings.Split(messageAttachment.URL, "https://cdn.discordapp.com/attachments/")[1]
					partMessage.PartTimestamp = int64(time.Now().Unix())

					log.Info("Updating VFS file part with new attachment URL", "file_id", fileID, "part_id", partMessage.PartID, "attachment_url", partMessage.PartAttachmentURL)
				}

				delete(remainingParts, int64(IDInt))
			}

			if len(messages) > 0 {
				beforeID = messages[len(messages)-1].ID
			}

			if len(remainingParts) == 0 {
				log.Info("All parts updated for channel", "channel_id", channelID)
				break
			}
		}
	}

	if err := database.UpdateVFSFile(file); err != nil {
		log.Error("Failed to update VFS file with new attachments", "file_id", fileID, "error", err)
	}

	return nil
}

func Start() {
	log.Info("Discord started")

	go func() {
		discordClient, err := connectToDiscord()

		if err != nil {
			panic("Failed to connect to Discord: " + err.Error())
		}

		if err := initializeSlashCommands(discordClient); err != nil {
			panic("Failed to initialize slash commands: " + err.Error())
		}

		DiscordClient = discordClient
	}()
}

func Stop() {
	log.Info("Discord stopped")

	if err := DiscordClient.Close(); err != nil {
		log.Error("Error closing Discord session: %v", err)
	}
}
