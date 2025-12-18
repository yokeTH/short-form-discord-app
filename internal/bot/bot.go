package bot

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"
)

type bot struct {
	session *discordgo.Session
	config  *config
}

func New(config *config) *bot {
	log.Info().Msg("Creating new Discord bot session")
	session, err := discordgo.New("Bot " + config.Token)
	if err != nil {
		log.Error().Err(err).Msg("Error creating Discord session")
		panic(fmt.Sprintf("Error creating Discord session, %v", err))
	}

	log.Info().Msg("Discord bot session created successfully")
	return &bot{
		session: session,
		config:  config,
	}
}
