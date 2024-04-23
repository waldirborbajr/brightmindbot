package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/waldirborbajr/brightmindbot/internal/env"
)

func main() {
	setupLog()

	// Recover Telegram token from env
	TELEGRAM_TOKEN := env.MustGetString("TELEGRAM_TOKEN")

	// Recever Webhook URL from env
	TELEGRAM_WEBHOOK := env.MustGetString("TELEGRAM_WEBHOOK")

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)

	defer func() { cancel() }()

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", healthzHandler)

	opts := []bot.Option{
		bot.WithDebug(),
		bot.WithDefaultHandler(defaultHandler),
	}

	tgbot, err := bot.New(TELEGRAM_TOKEN, opts...)
	switch {
	case err != nil:
		log.Panic().Msgf("ERROR: %v", err)
	}

	// Handle commands
	tgbot.RegisterHandler(bot.HandlerTypeMessageText, "/start", bot.MatchTypeExact, startHandler)
	tgbot.RegisterHandler(bot.HandlerTypeMessageText, "/help", bot.MatchTypeExact, helpHandler)

	_, err = tgbot.SetWebhook(ctx, &bot.SetWebhookParams{URL: TELEGRAM_WEBHOOK})

	go tgbot.StartWebhook(ctx)

	BOT_PORT := env.GetString("PORT", "3000")

	log.Info().Msgf("BOT_PORT: %s", BOT_PORT)

	sslbotServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", 8443),
		Handler: tgbot.WebhookHandler(),
		TLSConfig: &tls.Config{
			GetCertificate: func(*tls.ClientHelloInfo) (*tls.Certificate, error) {
				// Always get latest localhost.crt and localhost.key
				// ex: keeping certificates file somewhere in global location where created certificates updated and this closure function can refer that
				cert, err := tls.LoadX509KeyPair("localhost.crt", "localhost.key")
				if err != nil {
					return nil, err
				}
				return &cert, nil
			},
		},
	}

	botServer := &http.Server{
		// Addr:    ":" + BOT_PORT,
		Addr:    net.JoinHostPort("localhost", BOT_PORT),
		Handler: tgbot.WebhookHandler(),
	}

	mainServer := &http.Server{
		Addr:    net.JoinHostPort("localhost", "1469"),
		Handler: mux,
	}

	go func() {
		// err = http.ListenAndServe(":"+BOT_PORT, tgbot.WebhookHandler())

		if os.Getenv("SSL_ENABLED") == "true" {
			err = sslbotServer.ListenAndServeTLS("localhost.crt", "localhost.key")
		} else {
			err = botServer.ListenAndServe()
		}

		switch {
		case err != nil:
			log.Fatal().Msgf("ERROR: %v", err)
		}
	}()

	go func() {
		log.Info().Msgf("listening on %v\n", mainServer.Addr)
		err = mainServer.ListenAndServe()
		switch {
		case err != nil:
			log.Fatal().Msgf("ERROR: %v", err)
		}
	}()

	<-ctx.Done()

	log.Info().Msg("BrightMindBot is shutting down")
}

// defaultHandler is used for handling updates that don't have a specific handler
func defaultHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	log.Info().Msg("defaultHandler")
}

func startHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    update.Message.Chat.ID,
		Text:      "Starting",
		ParseMode: models.ParseModeMarkdown,
	})
}

func helpHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    update.Message.Chat.ID,
		Text:      "Cry for help!",
		ParseMode: models.ParseModeMarkdown,
	})
}

func healthzHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "It is alive")
}

// setupLog initializes the global logger
func setupLog() {
	// always log in UTC, with accurate timestamps
	zerolog.TimestampFunc = func() time.Time {
		return time.Now().UTC()
	}
	zerolog.TimeFieldFormat = time.RFC3339Nano
	// NodeJS/bunyan uses "msg" for MessageFieldName, but that's bad for LogDNA,
	// so don't do that here; do make error logging consistent with NodeJS however
	zerolog.ErrorFieldName = "err"

	switch env.MustGetString("ENVIRONMENT") {
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
