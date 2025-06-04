package discord

import (
	"os"

	"github.com/bwmarrin/discordgo"
	"github.com/charmbracelet/log"
)

var DiscordClient *discordgo.Session

func connectToDiscord() (*discordgo.Session, error) {
	discordSession, err := discordgo.New("Bot " + os.Getenv("DISCORD_TOKEN"))

	if err != nil {
		return nil, err
	}

	err = discordSession.Open()

	if err != nil {
		return nil, err
	}

	return discordSession, nil
}

func Start() {
	log.Info("[DISCORD] Started")

	go func() {
		discordClient, err := connectToDiscord()

		if err != nil {
			panic("Failed to connect to Discord: " + err.Error())
		}

		DiscordClient = discordClient
	}()
}

func Stop() {
	log.Info("[DISCORD] Stopped")

	if err := DiscordClient.Close(); err != nil {
		log.Error("[DISCORD] Error closing Discord session: %v", err)
	}
}
