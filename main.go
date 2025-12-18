package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/yokeTH/short-form-discord-app/internal/bot"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	botCfg := bot.NewFromENV()
	b := bot.New(botCfg)

	b.Start(ctx, stop)
}
