package discord

import (
	"os"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/golang/glog"
	"github.com/wordgen/wordgen"
	"golang.org/x/text/language"
)

func createRandomResourceName(format string, count int) string {
	words := []string{
		"Nexus",
		"Haven",
		"Citadel",
		"Collective",
		"Realm",
		"Legion",
		"Order",
		"Syndicate",
		"Den",
		"Sanctuary",
		"Outpost",
		"Conclave",
		"Keep",
		"Dominion",
		"Vanguard",
		"Alliance",
		"Echo",
		"Archive",
		"Lair",
		"Vault",
		"Forge",
		"Crucible",
		"Bastion",
		"Asylum",
		"Assembly",
		"Brotherhood",
		"Enclave",
		"Summit",
		"Temple",
		"Empire",
		"Chamber",
		"Faction",
		"Circle",
		"Guildhall",
		"Coalition",
		"Monastery",
		"Dynasty",
		"Watch",
		"Arcadia",
		"Syndra",
		"Frontier",
		"Refuge",
		"Genesis",
		"Obsidian",
		"Echelon",
		"Accord",
		"Horizon",
		"Sanctum",
		"Paragon",
		"Citadel",
	}

	generator := wordgen.NewGenerator()

	generator.Words = words

	if format == "title" {
		generator.Separator = " "
	} else {
		generator.Separator = "-"
	}

	generator.Casing = format
	generator.Count = count
	generator.Language = language.English

	name, _ := generator.Generate()

	return name
}

type DiscordAPIService struct {
	discordClient *discordgo.Session
}

var Service = DiscordAPIService{}

func (service *DiscordAPIService) Start() {
	discordSession, err := discordgo.New("Bot " + os.Getenv("DISCORD_TOKEN"))

	if err != nil {
		glog.Fatalf("Error creating Discord session: %v", err)

		return
	}

	discordSession.Identify.Intents = discordgo.IntentsAllWithoutPrivileged | discordgo.IntentMessageContent
	discordSession.Client.Timeout = time.Minute * 10

	discordSession.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		service.discordClient = discordSession
	})

	if err := discordSession.Open(); err != nil {
		glog.Fatalf("Error opening Discord session: %v", err)

		return
	}
}

func (service *DiscordAPIService) Stop() {
	if service.discordClient != nil {
		return
	}

	if err := service.discordClient.Close(); err != nil {
		glog.Errorf("Error closing Discord session: %v", err)
	}
}
