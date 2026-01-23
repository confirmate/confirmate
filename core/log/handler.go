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

// contextHandler wraps an slog.Handler and automatically includes attributes from context.
type contextHandler struct {
	handler slog.Handler
}

// newContextHandler creates a handler that includes context attributes.
func newContextHandler(h slog.Handler) *contextHandler {
	return &contextHandler{handler: h}
}

func (h *contextHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

func (h *contextHandler) Handle(ctx context.Context, r slog.Record) error {
	// Check if context has attributes to prepend
	ctxAttrs := attrsFromContext(ctx)
	if len(ctxAttrs) == 0 {
		// No context attributes, pass through directly
		return h.handler.Handle(ctx, r)
	}

	// Prepend context attributes to the record (so they appear first)
	// Pre-allocate slice with exact capacity needed
	originalAttrs := make([]slog.Attr, 0, r.NumAttrs())
	r.Attrs(func(a slog.Attr) bool {
		originalAttrs = append(originalAttrs, a)
		return true
	})

	// Create new record with context attributes first, then original
	newRecord := slog.NewRecord(r.Time, r.Level, r.Message, r.PC)
	newRecord.AddAttrs(ctxAttrs...)
	newRecord.AddAttrs(originalAttrs...)

	return h.handler.Handle(ctx, newRecord)
}

func (h *contextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return newContextHandler(h.handler.WithAttrs(attrs))
}

func (h *contextHandler) WithGroup(name string) slog.Handler {
	return newContextHandler(h.handler.WithGroup(name))
}
