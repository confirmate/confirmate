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

package util

import (
	"os"
	"path/filepath"
	"strings"
)

// ExpandPath expands a file path, replacing ~ with the user's home directory if present.
func ExpandPath(path string) (expanded string) {
	var home string
	var clean string
	var err error

	if path == "" {
		return path
	}
	if strings.HasPrefix(path, "~") {
		home, err = os.UserHomeDir()
		if err != nil {
			return path
		}

		clean = strings.TrimPrefix(path, "~")
		clean = strings.TrimPrefix(clean, string(os.PathSeparator))
		expanded = filepath.Join(home, clean)
		return expanded
	}

	return path
}
