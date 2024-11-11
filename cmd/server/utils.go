package main

import (
	"context"
	"errors"
	"os"
	"strconv"
	"time"

	"github.com/goodieshq/sweettooth/internal/server/core_pgx"
	"github.com/goodieshq/sweettooth/pkg/api/server"
	"github.com/rs/zerolog/log"
)

func openDB() (server.SweetToothServerConfig, *core_pgx.CorePGX) {

	pgHost := os.Getenv("POSTGRES_HOST")
	if pgHost == "" {
		pgHost = "localhost"
	}

	pgPort := os.Getenv("POSTGRES_PORT")
	if pgPort == "" {
		pgPort = "5432"
	}

	p, err := strconv.Atoi(pgPort)
	if err != nil || p <= 0x00000 || p >= 0x10000 {
		log.Fatal().Str("invalid", pgPort).Msg("invalid postgres port")
	}

	pgUser := os.Getenv("POSTGRES_USER")
	pgPass := os.Getenv("POSTGRES_PASSWORD")

	if pgUser == "" || pgPass == "" {
		log.Fatal().Msg("POSTGRES_USER and POSTGRES_PASSWORD are required")
	}

	pgName := os.Getenv("POSTGRES_DB")
	if pgName == "" {
		pgName = "sweettooth"
	}

	pgConnStr := "postgres://" +
		pgUser + ":" + pgPass + "@" + pgHost + ":" + pgPort + "/" + pgName

	if pgConnStr == "" {
		log.Fatal().Err(errors.New("missing database connection url")).Send()
	}

	secret := os.Getenv("SWEETTOOTH_SECRET") // secret token for JWT HMAC validation
	if secret == "" {
		log.Fatal().Err(errors.New("missing server secret")).Send()
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	core, err := core_pgx.NewCorePGX(ctx, pgConnStr)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize the database connection")
	}
	return server.SweetToothServerConfig{
		Secret: secret,
	}, core
}
