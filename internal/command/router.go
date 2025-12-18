package command

import (
	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"
)

type commandRouterDependency struct {
	appID string
}

func NewCommandRouterDependency(appID string) *commandRouterDependency {
	return &commandRouterDependency{
		appID: appID,
	}
}

func NewCommandRouter(deps *commandRouterDependency) func(*discordgo.Session, *discordgo.InteractionCreate) {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		cmdName := i.ApplicationCommandData().Name
		log.Info().Str("command", cmdName).Msg("Routing command")
		switch cmdName {
		case IGCommand.Name:
			log.Info().Msg("Routing to IGHandler")
			IGHandler(s, i, deps)
		default:
			log.Warn().Str("command", cmdName).Msg("Unknown command received")
		}
	}
}
