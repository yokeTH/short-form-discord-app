package bot

import (
	"context"

	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/yokeTH/short-form-discord-app/internal/command"
)

func (b *bot) Start(ctx context.Context, stop context.CancelFunc) {
	go func() {
		if err := b.session.Open(); err != nil {
			log.Error().Err(err).Msg("Cannot open Discord session")
			stop()
		}
	}()

	deps := command.NewCommandRouterDependency(b.config.AppID)
	router := command.NewCommandRouter(deps)

	b.session.AddHandler(router)

	log.Info().Msg("Bot is now running.")

	<-ctx.Done()

	log.Info().Msg("Shutdown signal received. Cleaning up...")

	if err := b.shutdown(); err != nil {
		log.Error().Err(err).Msg("Error during shutdown")
	} else {
		log.Info().Msg("Bot shutdown completed.")
	}
}

func (b *bot) shutdown() error {
	if err := b.session.Close(); err != nil {
		return fmt.Errorf("failed to close Discord session: %w", err)
	}
	return nil
}
