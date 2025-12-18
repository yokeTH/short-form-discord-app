package bot

import (
	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
)

type config struct {
	AppID string `env:"APP_ID,required"`
	Token string `env:"BOT_TOKEN,required"`
}

func NewConfig(token, appID string) *config {
	return &config{
		Token: token,
		AppID: appID,
	}
}

func NewFromENV() *config {
	config := &config{}

	if err := godotenv.Load(); err != nil {
		log.Error().Err(err).Msg("Unable to load .env file")
	}

	if err := env.Parse(config); err != nil {
		log.Panic().Err(err).Msg("Unable to parse env vars")
	}

	log.Info().Msg("Environment variables loaded successfully")
	return config
}
