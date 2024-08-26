package engine

import (
	"os"
	"time"

	"github.com/goodieshq/sweettooth/pkg/config"
	"github.com/natefinch/lumberjack"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func LoggingBasic() {
	logTerm := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339Nano,
		// TimeFormat: time.RFC3339,
	}

	log.Logger = log.Output(logTerm).With().Timestamp().Logger()
}

func setLogLevel(logger zerolog.Logger, levelName string) zerolog.Logger {
	level, err := zerolog.ParseLevel(levelName)
	if levelName == "" || err != nil || level == zerolog.NoLevel {
		log.Warn().Msgf("config log level '%s' is invalid, using default '%s'", levelName, config.DEFAULT_LOGLEVEL)
	}
	return logger.Level(level)
}

// note: this is not idempotent, it will add multiple key/values even if the key is the same
func AddLogKey(key string, val string) {
	log.Logger = log.Logger.With().Str(key, val).Logger()
}

func LoggingFile(filename string, level string) {
	logFile := &lumberjack.Logger{
		Filename:  filename, // target the logfile in the config module
		MaxSize:   25,       // 50 MB log file limit until rollover
		MaxAge:    365,      // store log file backups for 1 full year
		LocalTime: true,     // use local time instead of UTC for backup naming
		Compress:  true,     // use gzip compression for the backup files
	}

	// combine an stdout terminal writer and the log file
	log.Logger = setLogLevel(log.Output(
		zerolog.MultiLevelWriter(
			zerolog.ConsoleWriter{
				Out:        os.Stdout,
				TimeFormat: time.RFC3339,
			},
			logFile,
		),
	), level)
}
