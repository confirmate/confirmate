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
	"fmt"
	"io"
	"log/slog"
	"os"
	"sort"
	"strings"
	"time"
)

// contextHandler wraps an slog.Handler and automatically includes attributes from context.
type contextHandler struct {
	handler slog.Handler
	out     io.Writer
}

// newContextHandler creates a handler that includes context attributes.
func newContextHandler(h slog.Handler) *contextHandler {
	return &contextHandler{
		handler: h,
		out:     os.Stdout,
	}
}

func (h *contextHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

func (h *contextHandler) Handle(ctx context.Context, r slog.Record) error {
	// Collect context attrs
	ctxAttrs := attrsFromContext(ctx)

	// Collect record attrs (we need them twice: (a) detect multiline, (b) either print or forward)
	originalAttrs := make([]slog.Attr, 0, r.NumAttrs())
	r.Attrs(func(a slog.Attr) bool {
		originalAttrs = append(originalAttrs, a)
		return true
	})

	// Detect "format=multiline" on record attrs (so callers can opt-in per log line)
	multiline := false
	for _, a := range originalAttrs {
		// Removed the Resolve call as slog.Attr does not have this method
		if a.Key == "format" && a.Value.Kind() == slog.KindString && a.Value.String() == "multiline" {
			multiline = true
			break
		}
	}

	if multiline {
		// Print record ourselves (one attr per line), including context attrs first.
		// We intentionally do NOT call the wrapped handler in this mode.
		all := make([]slog.Attr, 0, len(ctxAttrs)+len(originalAttrs))
		all = append(all, ctxAttrs...)
		all = append(all, originalAttrs...)
		return h.handleMultiline(r, all)
	}

	// Default behaviour: prepend context attrs and pass through to wrapped handler
	if len(ctxAttrs) == 0 {
		return h.handler.Handle(ctx, r)
	}

	newRecord := slog.NewRecord(r.Time, r.Level, r.Message, r.PC)
	newRecord.AddAttrs(ctxAttrs...)
	newRecord.AddAttrs(originalAttrs...)

	return h.handler.Handle(ctx, newRecord)
}

func (h *contextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	nh := newContextHandler(h.handler.WithAttrs(attrs))
	nh.out = h.out
	return nh
}

func (h *contextHandler) WithGroup(name string) slog.Handler {
	nh := newContextHandler(h.handler.WithGroup(name))
	nh.out = h.out
	return nh
}

// WithOutput overrides where multiline output is written (defaults to os.Stdout).
func (h *contextHandler) WithOutput(w io.Writer) *contextHandler {
	if w != nil {
		h.out = w
	}
	return h
}

func (h *contextHandler) handleMultiline(r slog.Record, attrs []slog.Attr) error {
	w := h.out
	if w == nil {
		w = os.Stdout
	}

	ts := r.Time
	if ts.IsZero() {
		ts = time.Now()
	}

	// Desired style: "Mar 12 10:27:04.622 INF Message"
	timeStr := ts.Format("Jan _2 15:04:05.000")
	levelStr := strings.ToUpper(r.Level.String())

	if _, err := fmt.Fprintf(w, "%s %s %s\n", timeStr, levelStr, r.Message); err != nil {
		return err
	}

	// Split attrs into flat + groups; sort for stable order
	type kv struct {
		key string
		val slog.Value
	}
	var flat []kv
	type grp struct {
		key   string
		attrs []slog.Attr
	}
	var groups []grp

	for _, a := range attrs {
		// Hide the control attribute itself
		if a.Key == "format" && a.Value.Kind() == slog.KindString && a.Value.String() == "multiline" {
			continue
		}

		if a.Value.Kind() == slog.KindGroup {
			groups = append(groups, grp{key: a.Key, attrs: a.Value.Group()})
			continue
		}

		flat = append(flat, kv{key: a.Key, val: a.Value})
	}

	sort.SliceStable(flat, func(i, j int) bool { return flat[i].key < flat[j].key })
	sort.SliceStable(groups, func(i, j int) bool { return groups[i].key < groups[j].key })

	for _, item := range flat {
		if _, err := fmt.Fprintf(w, "  %s=%s\n", item.key, valueToString(item.val)); err != nil {
			return err
		}
	}

	for _, g := range groups {
		if err := h.printGroup(w, g.key, g.attrs); err != nil {
			return err
		}
	}

	return nil
}

func (h *contextHandler) printGroup(w io.Writer, prefix string, attrs []slog.Attr) error {
	// Flatten group to dotted keys: config.api_port=...
	type kv struct {
		key string
		val slog.Value
	}
	var flat []kv
	type grp struct {
		key   string
		attrs []slog.Attr
	}
	var groups []grp

	for _, a := range attrs {
		if a.Value.Kind() == slog.KindGroup {
			groups = append(groups, grp{key: a.Key, attrs: a.Value.Group()})
		} else {
			flat = append(flat, kv{key: a.Key, val: a.Value})
		}
	}

	sort.SliceStable(flat, func(i, j int) bool { return flat[i].key < flat[j].key })
	sort.SliceStable(groups, func(i, j int) bool { return groups[i].key < groups[j].key })

	for _, item := range flat {
		if _, err := fmt.Fprintf(w, "  %s.%s=%s\n", prefix, item.key, valueToString(item.val)); err != nil {
			return err
		}
	}

	for _, g := range groups {
		if err := h.printGroup(w, prefix+"."+g.key, g.attrs); err != nil {
			return err
		}
	}

	return nil
}

func valueToString(v slog.Value) string {
	v = v.Resolve()
	switch v.Kind() {
	case slog.KindString:
		return v.String()
	default:
		// for numbers, bools, slices, maps etc.
		return v.String()
	}
}
