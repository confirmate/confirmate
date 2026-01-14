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
	"context"
	"log/slog"
)

type contextKey string

const attrsKey contextKey = "log_attrs"

// WithAttrs stores log attributes in the context for automatic inclusion in all log messages.
// This is useful for request-scoped attributes like IDs that should appear in all logs.
func WithAttrs(ctx context.Context, attrs ...slog.Attr) context.Context {
	existing := attrsFromContext(ctx)
	return context.WithValue(ctx, attrsKey, append(existing, attrs...))
}

// attrsFromContext retrieves log attributes from the context.
func attrsFromContext(ctx context.Context) []slog.Attr {
	if attrs, ok := ctx.Value(attrsKey).([]slog.Attr); ok {
		return attrs
	}
	return nil
}
