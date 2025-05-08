package logger

import (
	"log/syslog"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Initialize sets up the global logger
func Initialize(level string, useSyslog bool, appName string) {
	// Parse log level
	logLevel, err := zerolog.ParseLevel(level)
	if err != nil {
		logLevel = zerolog.InfoLevel
	}

	// Configure global logger
	zerolog.SetGlobalLevel(logLevel)
	zerolog.TimeFieldFormat = time.RFC3339

	var output zerolog.ConsoleWriter
	if useSyslog {
		// Create syslog writer
		syslogWriter, err := syslog.New(syslog.LOG_DAEMON|syslog.LOG_INFO, appName)
		if err != nil {
			// Fallback to console if syslog fails
			output = zerolog.ConsoleWriter{
				Out:        os.Stderr,
				TimeFormat: time.RFC3339,
			}
			log.Logger = zerolog.New(output).With().Timestamp().Logger()
			log.Error().Err(err).Msg("Failed to initialize syslog, falling back to console output")
			return
		}

		// Create zerolog writer that writes to syslog
		output = zerolog.ConsoleWriter{
			Out:        syslogWriter,
			TimeFormat: time.RFC3339,
			NoColor:    true,
		}
	} else {
		// Use console writer for development
		output = zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		}
	}

	log.Logger = zerolog.New(output).With().Timestamp().Logger()
}

// Get returns the global logger instance
func Get() *zerolog.Logger {
	return &log.Logger
}
