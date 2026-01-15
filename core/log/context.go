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

// contextKey is a unique type for context keys to avoid collisions.
type contextKey string

const (
	// attrsKey is the context key for storing log attributes.
	attrsKey contextKey = "log_attrs"
)

// WithAttrs stores log attributes in the context for automatic inclusion in all log messages.
// This is useful for request-scoped attributes like IDs that should appear in all logs.
// Attributes are prepended to log output, appearing before message-specific attributes.
func WithAttrs(ctx context.Context, attrs ...slog.Attr) context.Context {
	if len(attrs) == 0 {
		return ctx
	}

	existing := attrsFromContext(ctx)
	if len(existing) == 0 {
		// No existing attrs, just set the new ones
		return context.WithValue(ctx, attrsKey, attrs)
	}

	// Append to existing attrs - pre-allocate with exact capacity
	combined := make([]slog.Attr, 0, len(existing)+len(attrs))
	combined = append(combined, existing...)
	combined = append(combined, attrs...)
	return context.WithValue(ctx, attrsKey, combined)
}

// attrsFromContext retrieves log attributes from the context.
func attrsFromContext(ctx context.Context) []slog.Attr {
	if attrs, ok := ctx.Value(attrsKey).([]slog.Attr); ok {
		return attrs
	}
	return nil
}
