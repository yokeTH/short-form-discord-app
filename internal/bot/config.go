package bot

import (
	"fmt"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
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
		fmt.Println("Unable to load .env file:", err)
	}

	if err := env.Parse(config); err != nil {
		panic(fmt.Sprintf("Unable to parse env vars: %s", err))
	}

	return config
}
