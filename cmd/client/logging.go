package main

import (
	"os"
	"time"

	"github.com/natefinch/lumberjack"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func initLoggingTerm() {
	logTerm := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
	}

	log.Logger = log.Output(logTerm).With().Timestamp().Logger()
}

func initLoggingFile(filename string) {
	logFile := &lumberjack.Logger{
		Filename:  filename, // target the logfile in the config module
		MaxSize:   50,       // 50 MB log file limit until rollover
		MaxAge:    365,      // store log file backups for 1 full year
		LocalTime: true,     // use local time instead of UTC for backup naming
		Compress:  true,     // use gzip compression for the backup files
	}

	// combine the existing logger and the file logger
	log.Logger = log.Output(
		zerolog.MultiLevelWriter(
			zerolog.ConsoleWriter{
				Out:        os.Stdout,
				TimeFormat: time.RFC3339,
			},
			logFile,
		),
	)
}

func setLogLevel(levelName string) {
	level, err := zerolog.ParseLevel(levelName)
	if err != nil || level == zerolog.NoLevel {
		level, _ = zerolog.ParseLevel(CONFIG_DEFAULT_LOGLEVEL)
	}
	log.Logger = log.Logger.Level(level)
}

func setLogKey(key string, val string) {
	log.Logger = log.Logger.With().Str(key, val).Logger()
}
