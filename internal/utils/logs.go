package utils

import (
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func GetLogs(types ...string) {

}

func SetupLogger(config *ConfigStructure) {
	devMode := os.Getenv("DEV_MODE")
	var _logger zerolog.Logger

	if config.LogOutput == "file" || config.LogOutput == "both" {
		logfile, err := os.OpenFile(config.LogFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)

		if err != nil {
			log.Fatal().Err(err).Msg("Failed to open log file")
		}
		if config.LogOutput == "both" {
			_logger = zerolog.New(zerolog.MultiLevelWriter(os.Stdout, logfile))
		} else {
			_logger = zerolog.New(logfile)
		}
	}

	if config.LogOutput == "console" {
		_logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})
	}

	_logger.Level(zerolog.Level(config.LogLevel))

	if devMode == "true" {
		_logger.
			With().
			Timestamp().
			Caller().
			Logger()
	}
	log.Logger = _logger
}
