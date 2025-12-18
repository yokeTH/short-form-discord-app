package main

import (
	"flag"

	"github.com/rs/zerolog/log"

	"github.com/bwmarrin/discordgo"
	"github.com/yokeTH/short-form-discord-app/internal/command"
)

var (
	GuildID  = flag.String("guild", "", "Test guild ID. If not passed - bot registers commands globally")
	BotToken = flag.String("token", "", "Bot access token")
)

var s *discordgo.Session

func init() { flag.Parse() }

func init() {
	var err error
	s, err = discordgo.New("Bot " + *BotToken)
	if err != nil {
		log.Fatal().Err(err).Msg("Invalid bot parameters")
	}
}

var (
	commands = []*discordgo.ApplicationCommand{
		&command.IGCommand,
	}
)

func main() {
	s.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Info().
			Str("username", s.State.User.Username).
			Str("discriminator", s.State.User.Discriminator).
			Msg("Logged in as")
	})
	err := s.Open()
	if err != nil {
		log.Fatal().Err(err).Msg("Cannot open the session")
	}

	log.Info().Msg("Adding commands...")
	registeredCommands := make([]*discordgo.ApplicationCommand, len(commands))
	for i, v := range commands {
		cmd, err := s.ApplicationCommandCreate(s.State.User.ID, *GuildID, v)
		if err != nil {
			log.Panic().Err(err).Str("command", v.Name).Msg("Cannot create command")
		}
		registeredCommands[i] = cmd
	}

	defer s.Close()

	log.Info().Msg("Gracefully shutting down.")
}
