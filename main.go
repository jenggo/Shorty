package main

import (
	"os"
	"os/signal"
	"shorty/app"
	"shorty/config"
	"shorty/pkg"
	"syscall"

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
	log.Logger = zerolog.New(zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: "[Mon] [2006-01-02] [15:04:05]",
	}).With().Timestamp().Logger()

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

	pkg.RedisAuth, err = pkg.NewRedis(config.Use.Redis.DB.Auth)
	if err != nil {
		log.Error().Err(err).Send()
	}

	defer func() {
		// Close redis connection
		pkg.Redis.Close()
		pkg.RedisAuth.Close()

		// Shutdown server
		if err := server.Shutdown(); err != nil {
			log.Error().Err(err).Send()
		}
	}()

	// Handle graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
}
