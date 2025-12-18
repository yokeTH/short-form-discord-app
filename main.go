package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog/log"
	"github.com/yokeTH/short-form-discord-app/internal/bot"
	"github.com/yokeTH/short-form-discord-app/internal/downloader/ig"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	if err := ig.InitializeProxies(); err != nil {
		log.Fatal().Err(err).Msg("Failed to setup proxies")
	}

	botCfg := bot.NewFromENV()
	b := bot.New(botCfg)

	b.Start(ctx, stop)
}
