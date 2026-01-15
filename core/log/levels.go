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
	"strings"
)

// Level is a log level that extends slog.Level with support for a custom TRACE level. It implements
// [encoding.TextUnmarshaler] for automatic unmarshaling in configuration files and [slog.Leveler]
// for use with slog handlers.
type Level slog.Level

// Log levels for Confirmate. We re-export standard slog levels and add a custom TRACE level for
// very detailed logging.
const (
	// LevelTrace is a custom log level below DEBUG for very detailed logging (e.g., SQL queries).
	// This is set to -8 to be below slog.LevelDebug (-4).
	LevelTrace Level = -8
)

// Standard slog levels (re-exported for convenience)
const (
	LevelDebug Level = Level(slog.LevelDebug) // -4
	LevelInfo  Level = Level(slog.LevelInfo)  // 0
	LevelWarn  Level = Level(slog.LevelWarn)  // 4
	LevelError Level = Level(slog.LevelError) // 8
)

// UnmarshalText implements encoding.TextUnmarshaler. It handles the custom TRACE level and
// delegates to [slog.Level.UnmarshalText] for standard levels.
func (l *Level) UnmarshalText(data []byte) error {
	text := string(data)

	// Handle custom TRACE level
	if strings.ToUpper(text) == "TRACE" {
		*l = LevelTrace
		return nil
	}

	// Delegate to slog.Level's UnmarshalText for standard levels
	var slogLevel slog.Level
	if err := slogLevel.UnmarshalText(data); err != nil {
		return err
	}
	*l = Level(slogLevel)
	return nil
}

// MarshalText implements encoding.TextMarshaler.
func (l Level) MarshalText() ([]byte, error) {
	return []byte(l.String()), nil
}

// String returns the string representation of the level.
func (l Level) String() string {
	if l == LevelTrace {
		return "TRACE"
	}

	return slog.Level(l).String()
}

// Level returns the underlying [slog.Level], implementing the [slog.Leveler] interface. This allows
// our custom Level type to be used wherever [slog.Leveler] is expected.
func (l Level) Level() slog.Level {
	return slog.Level(l)
}
