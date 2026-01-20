// Copyright 2016-2025 Fraunhofer AISEC
//
// SPDX-License-Identifier: Apache-2.0
//                                 /$$$$$$  /$$                                     /$$
//                               /$$__  $$|__/                                    | $$
//   /$$$$$$$  /$$$$$$  /$$$$$$$ | $$  \__/ /$$  /$$$$$$  /$$$$$$/$$$$   /$$$$$$  /$$$$$$    /$$$$$$
//  /$$_____/ /$$__  $$| $$__  $$| $$$$    | $$ /$$__  $$| $$_  $$_  $$ |____  $$|_  $$_/   /$$__  $$
// | $$      | $$  \ $$| $$  \ $$| $$_/    | $$| $$  \__/| $$ \ $$ \ $$  /$$$$$$$  | $$    | $$$$$$$$
// | $$      | $$  | $$| $$  | $$| $$      | $$| $$      | $$ | $$ | $$ /$$__  $$  | $$ /$$| $$_____/
// |  $$$$$$$|  $$$$$$/| $$  | $$| $$      | $$| $$      | $$ | $$ | $$|  $$$$$$$  |  $$$$/|  $$$$$$$
// \_______/ \______/ |__/  |__/|__/      |__/|__/      |__/ |__/ |__/ \_______/   \___/   \_______/

// This file is part of Confirmate Core.
package log_test

import (
	"encoding/json"
	"testing"

	"confirmate.io/core/log"
	"confirmate.io/core/util/assert"
)

func TestLevel_UnmarshalText(t *testing.T) {
	type args struct {
		text string
	}

	wantAnyError := func(t *testing.T, err error, _ ...any) bool {
		return assert.Error(t, err)
	}

	wantLevel := func(want log.Level, wantString string, wantInt int) assert.Want[log.Level] {
		return func(t *testing.T, got log.Level, _ ...any) bool {
			assert.Equal(t, want, got)
			assert.Equal(t, wantString, got.String())
			return assert.Equal(t, wantInt, int(got))
		}
	}

	tests := []struct {
		name    string
		args    args
		want    assert.Want[log.Level]
		wantErr assert.WantErr
	}{
		{
			name:    "INFO",
			args:    args{text: "INFO"},
			want:    wantLevel(log.LevelInfo, "INFO", 0),
			wantErr: assert.NoError,
		},
		{
			name:    "TRACE",
			args:    args{text: "TRACE"},
			want:    wantLevel(log.LevelTrace, "TRACE", -8),
			wantErr: assert.NoError,
		},
		{
			name:    "INFO+2",
			args:    args{text: "INFO+2"},
			want:    wantLevel(log.Level(2), "INFO+2", 2),
			wantErr: assert.NoError,
		},
		{
			name:    "WARN-1",
			args:    args{text: "WARN-1"},
			want:    wantLevel(log.Level(3), "INFO+3", 3),
			wantErr: assert.NoError,
		},
		{
			name:    "invalid",
			args:    args{text: "NOPE"},
			want:    assert.AnyValue[log.Level],
			wantErr: wantAnyError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				got log.Level
				err error
			)

			err = got.UnmarshalText([]byte(tt.args.text))
			tt.wantErr(t, err)
			tt.want(t, got)
		})
	}
}

func TestLevel_JSONUnmarshal(t *testing.T) {
	type Config struct {
		LogLevel log.Level `json:"log_level"`
	}

	tests := []struct {
		name    string
		json    string
		want    assert.Want[Config]
		wantErr assert.WantErr
	}{
		{
			name: "DEBUG",
			json: `{"log_level": "DEBUG"}`,
			want: func(t *testing.T, got Config, _ ...any) bool {
				return assert.Equal(t, log.LevelDebug, got.LogLevel)
			},
			wantErr: assert.NoError,
		},
		{
			name: "TRACE",
			json: `{"log_level": "TRACE"}`,
			want: func(t *testing.T, got Config, _ ...any) bool {
				return assert.Equal(t, log.LevelTrace, got.LogLevel)
			},
			wantErr: assert.NoError,
		},
		{
			name: "INFO+2",
			json: `{"log_level": "INFO+2"}`,
			want: func(t *testing.T, got Config, _ ...any) bool {
				return assert.Equal(t, log.Level(2), got.LogLevel)
			},
			wantErr: assert.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				got Config
				err error
			)

			err = json.Unmarshal([]byte(tt.json), &got)
			tt.wantErr(t, err)
			tt.want(t, got)
		})
	}
}
