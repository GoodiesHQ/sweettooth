package main

import (
	"context"
	"errors"
	"os"
	"time"

	"github.com/goodieshq/sweettooth/pkg/api/server"
	"github.com/goodieshq/sweettooth/pkg/util"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func loop() {
	// Initialize the logger for human-friendly output
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr}).With().Caller().Logger()

	godotenv.Load()

	connStr := os.Getenv("SWEETTOOTH_DB_URL") // backend database (postgres only for now)
	if connStr == "" {
		log.Fatal().Err(errors.New("missing database connection url")).Send()
	}

	secret := os.Getenv("SWEETTOOTH_SECRET") // secret token for JWT HMAC validation
	if secret == "" {
		log.Fatal().Err(errors.New("missing server secret")).Send()
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	core, err := server.NewCorePGX(ctx, connStr)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize the database connection")
	}

	srv, err := server.NewSweetToothServer(server.SweetToothServerConfig{
		Secret: secret,
	}, core)
	if err != nil {
		log.Fatal().Err(err).Send()
	}

	defer core.Close()

	srv.Run()

}

func loopRecoverable() {
	defer util.Recoverable(false)
	loop() // main server application loop
}

func main() {
	for {
		loopRecoverable()
		util.Countdown("restarting server in", 5, "s...")
	}
}
