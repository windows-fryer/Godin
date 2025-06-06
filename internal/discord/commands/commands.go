package commands

import "github.com/bwmarrin/discordgo"

func InteractionDefer(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	}); err != nil {
		return err
	}

	return nil
}

func InteractionFailed(s *discordgo.Session, i *discordgo.InteractionCreate, content string) error {
	if _, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: &content,
	}); err != nil {
		return err
	}

	return nil
}

func InteractionSuccess(s *discordgo.Session, i *discordgo.InteractionCreate, content string) error {
	if _, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: &content,
	}); err != nil {
		return err
	}

	return nil
}
