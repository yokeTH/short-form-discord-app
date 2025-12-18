package command

import (
	"github.com/bwmarrin/discordgo"
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
		switch i.ApplicationCommandData().Name {
		case IGCommand.Name:
			IGHandler(s, i, deps)
		}
	}
}
