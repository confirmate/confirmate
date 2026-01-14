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

import "log/slog"

// Log levels for Confirmate.
// We re-export standard slog levels and add a custom TRACE level for very detailed logging.
const (
	// LevelTrace is a custom log level below DEBUG for very detailed logging (e.g., SQL queries).
	// This is set to -8 to be below slog.LevelDebug (-4).
	LevelTrace = slog.LevelDebug - 4 // -8

	// Standard slog levels (re-exported for convenience)
	LevelDebug = slog.LevelDebug // -4
	LevelInfo  = slog.LevelInfo  // 0
	LevelWarn  = slog.LevelWarn  // 4
	LevelError = slog.LevelError // 8
)

// ParseLevel converts a string to a slog.Level, supporting our custom TRACE level.
// Valid values: TRACE, DEBUG, INFO, WARN, WARNING, ERROR
// Returns an error if the level string is not recognized.
func ParseLevel(levelStr string) (slog.Level, error) {
	switch levelStr {
	case "TRACE":
		return LevelTrace, nil
	case "DEBUG":
		return LevelDebug, nil
	case "INFO":
		return LevelInfo, nil
	case "WARN", "WARNING":
		return LevelWarn, nil
	case "ERROR":
		return LevelError, nil
	default:
		return LevelInfo, &InvalidLevelError{Level: levelStr}
	}
}

// InvalidLevelError is returned when ParseLevel receives an invalid level string.
type InvalidLevelError struct {
	Level string
}

func (e *InvalidLevelError) Error() string {
	return "unknown log level: " + e.Level + " (valid: TRACE, DEBUG, INFO, WARN, ERROR)"
}
