package bot

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

type bot struct {
	session *discordgo.Session
	config  *config
}

func New(config *config) *bot {
	session, err := discordgo.New("Bot " + config.Token)
	if err != nil {
		panic(fmt.Sprintf("Error creating Discord session, %v", err))
	}

	return &bot{
		session: session,
		config:  config,
	}
}
