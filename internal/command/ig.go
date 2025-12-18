package command

import (
	"github.com/bwmarrin/discordgo"
	"github.com/yokeTH/short-form-discord-app/internal/downloader/ig"
)

var IGCommand = discordgo.ApplicationCommand{
	Name: "Ig",
	Type: discordgo.MessageApplicationCommand,
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
	if i.Type != discordgo.InteractionApplicationCommand {
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
		return
	}

	videoData, err := ig.DownloadInstragramVideo(url)
	if err != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Failed to download Instagram video: " + err.Error(),
			},
		})
		return
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Here is your Instagram video:",
			Files: []*discordgo.File{
				{
					Name:        "video.mp4",
					ContentType: "video/mp4",
					Reader:      videoData,
				},
			},
		},
	})
}
