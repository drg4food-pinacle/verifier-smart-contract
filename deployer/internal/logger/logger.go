package logger

import (
	"fmt"
	"os"
	"strings"

	"github.com/rs/zerolog"
)

// SetupLogger initializes the logger with the specified mode (development or production)
func SetupLogger(mode string) zerolog.Logger {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	consoleWriter := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: "2006/01/02 15:04:05"}
	consoleWriter.FormatLevel = func(i any) string {
		level := strings.ToUpper(fmt.Sprintf("%-6s", i))
		switch i {
		case "info":
			return "\033[32m" + level + "\033[0m" // green
		case "warn":
			return "\033[33m" + level + "\033[0m" // yellow
		case "debug":
			return "\033[33m" + level + "\033[0m" // yellow
		case "error":
			return "\033[31m" + level + "\033[0m" // red
		case "fatal":
			return "\033[31m" + level + "\033[0m" // red
		default:
			return level
		}
	}

	if mode == "development" {
		zerolog.SetGlobalLevel(zerolog.DebugLevel) // For development, use Debug level
		return zerolog.New(consoleWriter).With().Timestamp().Logger()
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel) // For production, use Info level
		return zerolog.New(os.Stderr).With().Timestamp().Logger()
	}
}
