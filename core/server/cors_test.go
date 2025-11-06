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

package server

import "testing"

func Test_corsConfig_OriginAllowed(t *testing.T) {
	type fields struct {
		cfg Config
	}
	type args struct {
		origin string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name:   "Allow non-browser origin",
			fields: fields{},
			args: args{
				origin: "", // origin is only explicitly set by a browser
			},
			want: true,
		},
		{
			name: "Allowed origin",
			fields: fields{
				cfg: Config{
					CORS: CORS{
						AllowedOrigins: []string{"confirmate.io", "localhost"},
					},
				},
			},
			args: args{
				origin: "confirmate.io",
			},
			want: true,
		},
		{
			name: "Disallowed origin",
			fields: fields{
				cfg: Config{
					CORS: CORS{
						AllowedOrigins: []string{"confirmate.io", "localhost"},
					},
				},
			},
			args: args{
				origin: "confirmate.com",
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cors := &Server{
				cfg: tt.fields.cfg,
			}
			if got := cors.OriginAllowed(tt.args.origin); got != tt.want {
				t.Errorf("corsConfig.OriginAllowed() = %v, want %v", got, tt.want)
			}
		})
	}
}
