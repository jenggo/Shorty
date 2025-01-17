package main

import (
	"io"
	"os"
	"os/signal"
	"shorty/app"
	"shorty/config"
	"shorty/pkg"
	"syscall"

	zlogsentry "github.com/archdx/zerolog-sentry"
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	// Init config
	if err := cleanenv.ReadConfig("config.yaml", &config.Use); err != nil {
		log.Fatal().Err(err).Send()
	}

	// Set logger (zerolog)
	zerolog.SetGlobalLevel(zerolog.Level(config.Use.App.LogLevel))
	var writeLog io.Writer = zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: "[Mon] [2006-01-02] [15:04:05]",
	}

	if config.Use.App.Sentry != "" {
		host, _ := os.Hostname()
		w, err := zlogsentry.New(
			config.Use.App.Sentry,
			zlogsentry.WithRelease(config.AppVersion),
			zlogsentry.WithServerName(host),
			zlogsentry.WithSampleRate(1.0),
		)
		if err != nil {
			log.Fatal().Err(err).Msg("error initializing Sentry client")
		}

		writeLog = zerolog.MultiLevelWriter(w, writeLog)
	}

	log.Logger = zerolog.New(writeLog).With().Timestamp().Logger()

	// Run server
	server, err := app.RunServer()
	if err != nil {
		log.Error().Err(err).Send()
	}

	// Open Redis connection for every DB
	pkg.Redis, err = pkg.NewRedis()
	if err != nil {
		log.Error().Err(err).Send()
	}

	// Run cleanup objects for expired shorty
	if config.Use.S3.Enable && config.Use.S3.CleanupInterval > 0 {
		pkg.Redis.StartCleanupScheduler()
	}

	pkg.RedisAuth, err = pkg.NewRedis(config.Use.Redis.DB.Auth)
	if err != nil {
		log.Error().Err(err).Send()
	}

	defer func() {

		// Shutdown server
		if err := server.Shutdown(); err != nil {
			log.Error().Err(err).Send()
		}

		// Close redis connection
		pkg.Redis.Close()
		pkg.RedisAuth.Close()
	}()

	// Handle graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
}
