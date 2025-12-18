package command

import (
	"os"
	"time"

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

	content := "This is a bot for downloading Instagram videos and more!"

	embed := &discordgo.MessageEmbed{
		Title:       "Bot Information",
		Description: content,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:  "Repository",
				Value: "[GitHub Repository](github.com/yokeTH/short-form-discord-app)",
			},
			{
				Name:  "Usage Example",
				Value: "`/ig <instagram_url>`",
			},
			{
				Name:  "Git Hash",
				Value: gitHash,
			},
		},
		Color: 0x5865F2,
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Created by yoketh",
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
			Flags:  discordgo.MessageFlagsEphemeral,
		},
	})

	if err != nil {
		log.Error().Err(err).Msg("Failed to send info response")
	}
}
