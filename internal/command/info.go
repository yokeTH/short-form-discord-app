package command

import (
	"os"

	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"
)

var InfoCommand = discordgo.ApplicationCommand{
	Name:        "info",
	Description: "Show information about the bot",
	Type:        discordgo.ChatApplicationCommand,
	Contexts: &[]discordgo.InteractionContextType{
		discordgo.InteractionContextBotDM,
		discordgo.InteractionContextGuild,
		discordgo.InteractionContextPrivateChannel,
	},
	IntegrationTypes: &[]discordgo.ApplicationIntegrationType{
		discordgo.ApplicationIntegrationUserInstall,
		discordgo.ApplicationIntegrationGuildInstall,
	},
}

func InfoHandler(s *discordgo.Session, i *discordgo.InteractionCreate, deps *commandRouterDependency) {
	log.Info().Msg("InfoHandler invoked")
	if i.Type != discordgo.InteractionApplicationCommand {
		log.Warn().Msg("Interaction type is not ApplicationCommand, returning")
		return
	}

	gitHash := os.Getenv("GITHASH")
	if gitHash == "" {
		gitHash = "unknown"
	}

	content := "This is a bot for downloading Instagram videos and more!\n\nCreated by yoketh.\nGit Hash: " + gitHash

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})

	if err != nil {
		log.Error().Err(err).Msg("Failed to send info response")
	}
}
