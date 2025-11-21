package logger

import (
	"github.com/rs/zerolog"
)

// Logger is the global logger instance
var Logger zerolog.Logger

func init() {
	Logger = SetupLogger("development")
	Logger.Info().Msg("Logger Initialized")
}
