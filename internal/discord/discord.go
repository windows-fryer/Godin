package discord

import (
	"os"

	"github.com/bwmarrin/discordgo"
	"github.com/charmbracelet/log"
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
