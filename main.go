package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	setupLog()

	// Recover Telegram token from env
	TELEGRAM_TOKEN := os.Getenv("TELEGRAM_TOKEN")
	switch TELEGRAM_TOKEN {
	case "":
		log.Fatal().Msg("TELEGRAM_TOKEN not set")
	}

	log.Info().Msgf("TELEGRAM_TOKEN: %s", TELEGRAM_TOKEN)

	// Recever Webhook URL from env
	TELEGRAM_WEBHOOK := os.Getenv("TELEGRAM_WEBHOOK")
	switch TELEGRAM_WEBHOOK {
	case "":
		log.Fatal().Msg("TELEGRAM_WEBHOOK not set")
	}

	log.Info().Msgf("TELEGRAM_WEBHOOK: %s", TELEGRAM_WEBHOOK)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)

	defer func() { cancel() }()

	opts := []bot.Option{
		bot.WithDebug(),
		bot.WithDefaultHandler(defaultHandler),
	}

	tgbot, err := bot.New(TELEGRAM_TOKEN, opts...)
	switch {
	case err != nil:
		log.Panic().Msgf("ERROR: %v", err)
	}

	_, err = tgbot.SetWebhook(ctx, &bot.SetWebhookParams{URL: TELEGRAM_WEBHOOK})

	go tgbot.StartWebhook(ctx)

	BOT_PORT := os.Getenv("BOT_PORT")
	switch BOT_PORT {
	case "":
		log.Info().Msg("BOT_PORT not set. Assuming deafault port :3000")
		BOT_PORT = "3000"
	}

	log.Info().Msgf("BOT_PORT: %s", BOT_PORT)

	go func() {
		err = http.ListenAndServe(":"+BOT_PORT, tgbot.WebhookHandler())
		switch {
		case err != nil:
			log.Fatal().Msgf("ERROR: %v", err)
		}
	}()

	<-ctx.Done()
	log.Info().Msg("BrightMindBot is shutting down")
}

func defaultHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	log.Info().Msg("defaultHandler")
}

func setupLog() {
	// always log in UTC, with accurate timestamps
	zerolog.TimestampFunc = func() time.Time {
		return time.Now().UTC()
	}
	zerolog.TimeFieldFormat = time.RFC3339Nano
	// NodeJS/bunyan uses "msg" for MessageFieldName, but that's bad for LogDNA,
	// so don't do that here; do make error logging consistent with NodeJS however
	zerolog.ErrorFieldName = "err"

	switch os.Getenv("ENVIRONMENT") {
	case "dev":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
		// zerolog.SetGlobalLevel(zerolog.DebugLevel)

	case "prod":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
		zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	}
	log.Logger = zerolog.New(os.Stdout).
		With().
		Timestamp().
		Logger()
	log.Info().Msg("Logging initialized")
}
