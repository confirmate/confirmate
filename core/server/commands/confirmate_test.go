// Copyright 2016-2026 Fraunhofer AISEC
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

package commands

import (
	"os"
	"path/filepath"
	"testing"

	"confirmate.io/core/util/assert"
	"github.com/urfave/cli/v3"
)

func boolFlagDefault(flags []cli.Flag, name string) (value bool, found bool) {
	var (
		flag     cli.Flag
		boolFlag *cli.BoolFlag
		ok       bool
	)

	for _, flag = range flags {
		boolFlag, ok = flag.(*cli.BoolFlag)
		if ok && boolFlag.Name == name {
			return boolFlag.Value, true
		}
	}

	return false, false
}

func TestOAuthServerFlags_Defaults(t *testing.T) {
	tests := []struct {
		name     string
		flagName string
		want     bool
	}{
		{
			name:     "oauth2 embedded is enabled by default",
			flagName: "oauth2-embedded",
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				got *cli.BoolFlag
			)

			for _, f := range oauthServerFlags {
				flag, ok := f.(*cli.BoolFlag)
				if ok && flag.Name == tt.flagName {
					got = flag
					break
				}
			}

			if !assert.NotNil(t, got) {
				return
			}

			assert.Equal(t, tt.want, got.Value)
		})
	}
}

func TestCommandDBInMemoryDefaults(t *testing.T) {
	tests := []struct {
		name      string
		flags     []cli.Flag
		flagName  string
		wantValue bool
	}{
		{
			name:      "confirmate command defaults db-in-memory to true",
			flags:     ConfirmateCommand.Flags,
			flagName:  "db-in-memory",
			wantValue: true,
		},
		{
			name:      "orchestrator command keeps db-in-memory default false",
			flags:     OrchestratorCommand.Flags,
			flagName:  "db-in-memory",
			wantValue: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				gotValue bool
				found    bool
			)

			gotValue, found = boolFlagDefault(tt.flags, tt.flagName)
			if !assert.True(t, found) {
				return
			}

			assert.Equal(t, tt.wantValue, gotValue)
		})
	}
}

func TestResolveDefaultPolicyPath(t *testing.T) {
	t.Run("resolves to core fallback when original path is missing", func(t *testing.T) {
		var (
			tmp      string
			oldWd    string
			err      error
			input    string
			expected string
			got      string
		)

		tmp = t.TempDir()
		oldWd, err = os.Getwd()
		if err != nil {
			t.Fatalf("getwd: %v", err)
		}

		t.Cleanup(func() {
			_ = os.Chdir(oldWd)
		})

		err = os.Chdir(tmp)
		if err != nil {
			t.Fatalf("chdir: %v", err)
		}

		expected = filepath.Join("core", "policies", "security-metrics", "metrics")
		err = os.MkdirAll(expected, 0o755)
		if err != nil {
			t.Fatalf("mkdir fallback path: %v", err)
		}

		input = filepath.Join("policies", "security-metrics", "metrics")
		got = resolveDefaultPolicyPath(input)

		assert.Equal(t, expected, got)
	})
}
