package main

import (
	"flag"
	"os"

	"github.com/rs/zerolog/log"

	"github.com/bwmarrin/discordgo"
	"github.com/yokeTH/short-form-discord-app/internal/command"
)

var (
	BotToken = flag.String("token", "", "Bot access token")
	Remove   = flag.Bool("remove", false, "Remove commands instead of registering them")
)

var s *discordgo.Session

func init() {
	flag.Parse()
	token := *BotToken
	if token == "" {
		token = os.Getenv("DISCORD_BOT_TOKEN")
	}
	if token == "" {
		log.Fatal().Msg("Bot token not provided via --token flag or DISCORD_BOT_TOKEN environment variable")
	}
	var err error
	s, err = discordgo.New("Bot " + token)
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

	defer s.Close()

	if *Remove {
		log.Info().Msg("Removing commands...")
		for _, v := range commands {
			err := s.ApplicationCommandDelete(s.State.User.ID, "", v.ID)
			if err != nil {
				log.Panic().Err(err).Str("command", v.Name).Msg("Cannot remove command")
			} else {
				log.Info().Str("command", v.Name).Msg("Removed command")
			}
		}
	} else {
		log.Info().Msg("Adding commands...")
		registeredCommands := make([]*discordgo.ApplicationCommand, len(commands))
		for i, v := range commands {
			cmd, err := s.ApplicationCommandCreate(s.State.User.ID, "", v)
			if err != nil {
				log.Panic().Err(err).Str("command", v.Name).Msg("Cannot create command")
			}
			registeredCommands[i] = cmd
		}
	}

	log.Info().Msg("Gracefully shutting down.")
}
