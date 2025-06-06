package guild

import (
	"github.com/bwmarrin/discordgo"
	"github.com/godin/internal/discord/commands"
)

var UploadFileCommand = &discordgo.ApplicationCommand{
	Name:        "file",
	Description: "Manage files",
	Type:        discordgo.ChatApplicationCommand,

	Options: []*discordgo.ApplicationCommandOption{{
		Name:        "upload",
		Description: "Upload a file",
		Type:        discordgo.ApplicationCommandOptionSubCommand,

		Options: []*discordgo.ApplicationCommandOption{{
			Name:        "file",
			Description: "The file to upload",
			Type:        discordgo.ApplicationCommandOptionAttachment,
			Required:    true,
		}},
	}},
}

func UploadFile(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	if err := commands.InteractionDefer(s, i); err != nil {
		return commands.InteractionFailed(s, i, "Failed to defer interaction: "+err.Error())
	}

	attachmentId := i.ApplicationCommandData().GetOption("upload").GetOption("file").Value.(string)
	attachment := i.ApplicationCommandData().Resolved.Attachments[attachmentId]

	if attachment == nil {
		return commands.InteractionFailed(s, i, "No file provided in the command.")
	}

	return commands.InteractionSuccess(s, i, "File uploaded successfully!")
}
