package main

import (
	"os"
	"time"

	"github.com/goodieshq/sweettooth/internal/util"
	"github.com/goodieshq/sweettooth/pkg/api/server"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func loop() {
	// Initialize the logger for human-friendly output
	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: time.RFC3339,
	}).With().Caller().Logger()

	godotenv.Load()

	cfg, core := openDB()

	srv, err := server.NewSweetToothServer(cfg, core)
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
