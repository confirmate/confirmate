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
	"testing"

	"confirmate.io/core/util/assert"
)

func TestWithAttrs(t *testing.T) {
	type args struct {
		ctx   context.Context
		attrs []slog.Attr
	}

	base := context.Background()
	withExisting := WithAttrs(base, slog.String("a", "1"))

	tests := []struct {
		name string
		args args
		want assert.Want[context.Context]
	}{
		{
			name: "no attrs returns original ctx",
			args: args{ctx: base},
			want: func(t *testing.T, got context.Context, args ...any) bool {
				wantCtx := args[0].(context.Context)
				assert.True(t, got == wantCtx)
				return assert.Nil(t, attrsFromContext(got))
			},
		},
		{
			name: "stores attrs",
			args: args{ctx: base, attrs: []slog.Attr{slog.String("request_id", "abc")}},
			want: func(t *testing.T, got context.Context, args ...any) bool {
				wantAttrs := args[1].([]slog.Attr)
				return assert.Equal(t, wantAttrs, attrsFromContext(got), assert.CompareAllUnexported())
			},
		},
		{
			name: "appends to existing attrs",
			args: args{ctx: withExisting, attrs: []slog.Attr{slog.String("b", "2")}},
			want: func(t *testing.T, got context.Context, _ ...any) bool {
				return assert.Equal(
					t,
					[]slog.Attr{slog.String("a", "1"), slog.String("b", "2")},
					attrsFromContext(got),
					assert.CompareAllUnexported(),
				)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := WithAttrs(tt.args.ctx, tt.args.attrs...)
			tt.want(t, got, tt.args.ctx, tt.args.attrs)
		})
	}
}

func TestFindAttr(t *testing.T) {
	tests := []struct {
		name  string
		attrs []slog.Attr
		key   string
		want  assert.Want[bool]
	}{
		{
			name:  "found string attr",
			attrs: []slog.Attr{slog.String("k", "v")},
			key:   "k",
			want: func(t *testing.T, got bool, msgAndArgs ...any) bool {
				if !assert.True(t, got) {
					return false
				}
				a := msgAndArgs[0].(*slog.Attr)
				return assert.Equal(t, "v", a.Value.String())
			},
		},
		{
			name:  "not found",
			attrs: []slog.Attr{slog.String("k", "v")},
			key:   "missing",
			want: func(t *testing.T, got bool, _ ...any) bool {
				return assert.False(t, got)
			},
		},
		{
			name:  "group attr present",
			attrs: []slog.Attr{slog.Group("g", slog.String("inner", "x"))},
			key:   "g",
			want: func(t *testing.T, got bool, msgAndArgs ...any) bool {
				if !assert.True(t, got) {
					return false
				}
				a := msgAndArgs[0].(*slog.Attr)
				v := a.Value.Resolve()
				return assert.Equal(t, slog.KindGroup, v.Kind())
			},
		},
		{
			name:  "empty slice",
			attrs: nil,
			key:   "k",
			want: func(t *testing.T, got bool, _ ...any) bool {
				return assert.False(t, got)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a, ok := FindAttr(tt.attrs, tt.key)
			tt.want(t, ok, a)
		})
	}
}
