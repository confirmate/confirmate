package logconfig

import (
	"log/slog"
	"os"

	"github.com/lmittmann/tint"
)

var (
	logger *slog.Logger
)

// TODO(all): Maybe move package and component output to the beginning or end of the log line

// InitializeLogger initializes the logger
func InitializeLogger() {
	logger = slog.New(tint.NewHandler(os.Stdout, nil))
	logger = logger.With("package", "collector")
	slog.SetDefault(logger)
}

// GetLogger returns the logger
func GetLogger() *slog.Logger {
	if logger == nil {
		// Initialize logger if not already initialized
		InitializeLogger()
	}
	return logger
}
