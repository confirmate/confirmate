// Copyright 2016-2025 Fraunhofer AISEC
//
// SPDX-License-Identifier: Apache-2.0
//
//                                 /$$$$$$  /$$                                     /$$
//                               /$$__  $$|__/                                    | $$
//   /$$$$$$$  /$$$$$$  /$$$$$$$ | $$  \__/ /$$  /$$$$$$  /$$$$$$/$$$$   /$$$$$$  /$$$$$$    /$$$$$$
//  /$$_____/ /$$__  $$| $$__  $$| $$$$    | $$ /$$__  $$| $$_  $$_  $$ |____  $$|_  $$_/   /$$__  $$
// | $$      | $$  \ $$| $$  \ $$| $$_/    | $$| $$  \__/| $$ \ $$ \ $$  /$$$$$$$  | $$    | $$$$$$$$
// | $$      | $$  | $$| $$  | $$| $$      | $$| $$      | $$ | $$ | $$ /$$__  $$  | $$ /$$| $$_____/
// |  $$$$$$$|  $$$$$$/| $$  | $$| $$      | $$| $$      | $$ | $$ | $$|  $$$$$$$  |  $$$$/|  $$$$$$$
// \_______/ \______/ |__/  |__/|__/      |__/|__/      |__/ |__/ |__/ \_______/   \___/   \_______/
//
// This file is part of Confirmate Core.

package log

import (
	"log/slog"
	"os"

	"github.com/lmittmann/tint"
	"github.com/mattn/go-isatty"
)

var (
	// logger is the default logger instance for Confirmate.
	logger *slog.Logger

	// colorEnabled tracks whether ANSI colors are supported by the output.
	colorEnabled bool
)

func init() {
	// Detect color support based on terminal capabilities
	colorEnabled = isatty.IsTerminal(os.Stdout.Fd())

	// Initialize with INFO level by default, wrapped with context handler
	logger = slog.New(newContextHandler(tint.NewHandler(os.Stdout, &tint.Options{
		Level:   LevelInfo,
		NoColor: !colorEnabled,
	})))
	slog.SetDefault(logger)
}

// Configure configures the default logger with the specified level string.
// Valid values: TRACE, DEBUG, INFO, WARN, WARNING, ERROR
// Returns an error if the level string is not recognized.
func Configure(levelStr string) error {
	var level Level
	err := level.UnmarshalText([]byte(levelStr))
	if err != nil {
		return err
	}

	// Create new handler with the specified level, wrapped with context handler
	logger = slog.New(newContextHandler(tint.NewHandler(os.Stdout, &tint.Options{
		Level:   level,
		NoColor: !colorEnabled,
	})))
	slog.SetDefault(logger)

	slog.Debug("Log level configured", slog.String("level", levelStr))
	return nil
}

// ColorEnabled returns whether ANSI color codes are supported by the output.
// This is automatically detected based on whether stdout is a terminal.
func ColorEnabled() bool {
	return colorEnabled
}

// Err is a re-export of tint.Err for convenient error formatting in log attributes.
// Usage: slog.Error("message", log.Err(err))
var Err = tint.Err
