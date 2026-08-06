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

package server

import (
	"encoding/json"
	"fmt"
	"os"
)

// DemoSeedFile is the structure of the JSON file passed via --demo-seed-file.
// Only the Users field is consumed by the server — audit scopes and permissions
// are created via REST calls in demo.sh so that the CreateAuditScope RPC runs
// the full pipeline (autoCreateControlsInScope + audit trail events).
type DemoSeedFile struct {
	Users []DemoUser `json:"users"`
}

// LoadDemoSeedFile reads and parses the JSON demo seed file at path.
func LoadDemoSeedFile(path string) (*DemoSeedFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading demo seed file: %w", err)
	}
	var sf DemoSeedFile
	if err = json.Unmarshal(data, &sf); err != nil {
		return nil, fmt.Errorf("parsing demo seed file: %w", err)
	}
	return &sf, nil
}
