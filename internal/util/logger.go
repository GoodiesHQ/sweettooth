package util

import (
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog"
)

func Logger(routine string) zerolog.Logger {
	return log.With().Str("routine", routine).Logger()
}