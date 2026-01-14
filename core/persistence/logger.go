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

package persistence

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"confirmate.io/core/log"
	"gorm.io/gorm/logger"
)

// slogGormLogger integrates GORM's logger with slog.
// SQL queries are only logged at DEBUG level to reduce noise in production.
type slogGormLogger struct{}

// newSlogGormLogger creates a new GORM logger that uses slog.
func newSlogGormLogger() logger.Interface {
	return &slogGormLogger{}
}

// LogMode is a no-op since we control logging via slog's level configuration.
func (l *slogGormLogger) LogMode(level logger.LogLevel) logger.Interface {
	return l
}

// Info logs informational messages at DEBUG level.
func (l *slogGormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	slog.DebugContext(ctx, fmt.Sprintf(msg, data...))
}

// Warn logs warning messages.
func (l *slogGormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	slog.WarnContext(ctx, fmt.Sprintf(msg, data...))
}

// Error logs error messages.
func (l *slogGormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	slog.ErrorContext(ctx, fmt.Sprintf(msg, data...))
}

// Trace logs SQL queries at TRACE level only.
// This ensures SQL queries don't clutter DEBUG or INFO-level logs in production.
func (l *slogGormLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	// Only log if TRACE level is enabled
	if !slog.Default().Enabled(ctx, log.LevelTrace) {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()

	if err != nil {
		slog.LogAttrs(ctx, log.LevelTrace, "SQL query failed",
			slog.Duration("elapsed", elapsed),
			slog.String("sql", sql),
			slog.Int64("rows", rows),
			slog.String("error", err.Error()),
		)
	} else {
		slog.LogAttrs(ctx, log.LevelTrace, "SQL query",
			slog.Duration("elapsed", elapsed),
			slog.String("sql", sql),
			slog.Int64("rows", rows),
		)
	}
}
