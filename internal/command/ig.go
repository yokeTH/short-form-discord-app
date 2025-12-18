package command

import (
	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"
	"github.com/yokeTH/short-form-discord-app/internal/downloader/ig"
)

var IGCommand = discordgo.ApplicationCommand{
	Name:        "ig",
	Description: "Download an Instagram video by URL",
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
	Options: []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "url",
			Description: "The URL to process",
			Required:    true,
		},
	},
}

func IGHandler(s *discordgo.Session, i *discordgo.InteractionCreate, deps *commandRouterDependency) {
	log.Info().Msg("IGHandler invoked")
	if i.Type != discordgo.InteractionApplicationCommand {
		log.Warn().Msg("Interaction type is not ApplicationCommand, returning")
		return
	}

	data := i.ApplicationCommandData()
	var url string
	for _, option := range data.Options {
		if option.Name == "url" && option.Type == discordgo.ApplicationCommandOptionString {
			url = option.StringValue()
			break
		}
	}
	if url == "" {
		log.Warn().Msg("No URL provided in command options, returning")
		return
	}

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
	if err != nil {
		log.Error().Err(err).Msg("Failed to defer interaction response")
		return
	}

	videoData, err := ig.DownloadInstragramVideo(url)
	if err != nil {
		log.Error().Err(err).Str("url", url).Msg("Failed to download Instagram video")
		s.FollowupMessageCreate(i.Interaction, false, &discordgo.WebhookParams{
			Content: "Failed to download Instagram video: " + err.Error(),
		})
		return
	}

	log.Info().Str("url", url).Msg("Instagram video downloaded successfully, sending to user")
	s.FollowupMessageCreate(i.Interaction, false, &discordgo.WebhookParams{
		Files: []*discordgo.File{
			{
				Name:        "video.mp4",
				ContentType: "video/mp4",
				Reader:      videoData,
			},
		},
		Components: []discordgo.MessageComponent{
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.Button{
						Label: "Original Post",
						Style: discordgo.LinkButton,
						URL:   url,
					},
				},
			},
		},
	})
}
