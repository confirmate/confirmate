// Copyright 2016-2026 Fraunhofer AISEC
//
// SPDX-License-Identifier: Apache-2.0
//
//                                 /$$$$$$  /$$                                     /$$
//                               /$$__  $$|__/                                    | $$
//   /$$$$$$$  /$$$$$$  /$$$$$$$ | $$  \__/ /$$  /$$$$$$  /$$$$$$/$$$$   /$$$$$$  /$$$$$$    /$$$$$$
//  /$$_____/ /$$__  $$| $$__  $$| $$$$    | $$ /$$__  $$| $$_  $$_  $$ |____  $$|_  $$_/   /$$__  $$
// | $$      | $$  \ $$| $$  \ $$| $$_/    | $$| $$      | $$ | $$ | $$ /$$__  $$  | $$ /$$| $$_____/
// |  $$$$$$$|  $$$$$$/| $$  | $$| $$      | $$| $$      | $$ | $$ | $$|  $$$$$$$  |  $$$$/|  $$$$$$$
// \_______/ \______/ |__/  |__/|__/      |__/|__/      |__/ |__/ |__/ \_______/   \___/   \_______/
//
// This file is part of Confirmate Core.

package server

import (
	"os"
	"path/filepath"
	"testing"

	"confirmate.io/core/util/assert"
)

func TestLoadDemoSeedFile(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		writeFile bool
		wantErr   assert.WantErr
		wantUsers int
	}{
		{
			name:      "file does not exist",
			writeFile: false,
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.ErrorContains(t, err, "reading demo seed file")
			},
		},
		{
			name:      "invalid json",
			content:   "{not valid json",
			writeFile: true,
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.ErrorContains(t, err, "parsing demo seed file")
			},
		},
		{
			name:      "valid file with users",
			content:   `{"users": [{"username": "alice", "password": "alice", "firstName": "Alice", "lastName": "Adams"}]}`,
			writeFile: true,
			wantErr:   assert.NoError,
			wantUsers: 1,
		},
		{
			name:      "valid file with no users",
			content:   `{"users": []}`,
			writeFile: true,
			wantErr:   assert.NoError,
			wantUsers: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "seed.json")
			if tt.writeFile {
				assert.NoError(t, os.WriteFile(path, []byte(tt.content), 0644))
			}

			sf, err := LoadDemoSeedFile(path)
			tt.wantErr(t, err)
			if err == nil {
				assert.NotNil(t, sf)
				assert.Equal(t, tt.wantUsers, len(sf.Users))
			}
		})
	}
}
