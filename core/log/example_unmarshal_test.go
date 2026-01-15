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

package log_test

import (
	"encoding/json"
	"fmt"

	"confirmate.io/core/log"
)

// Example showing how to use log.Level's UnmarshalText in configuration structs.
// This works for both standard slog levels (DEBUG, INFO, WARN, ERROR) and the custom TRACE level.
func Example_unmarshalLevel() {
	// Configuration struct that uses log.Level (which supports TRACE)
	type Config struct {
		LogLevel log.Level `json:"log_level"`
	}

	// JSON configuration with standard log level
	jsonConfig := `{"log_level": "DEBUG"}`
	var config Config
	if err := json.Unmarshal([]byte(jsonConfig), &config); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Parsed log level: %v (numeric: %d)\n", config.LogLevel, int(config.LogLevel))

	// JSON configuration with custom TRACE level
	jsonConfig2 := `{"log_level": "TRACE"}`
	var config2 Config
	if err := json.Unmarshal([]byte(jsonConfig2), &config2); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Parsed TRACE level: %v (numeric: %d)\n", config2.LogLevel, int(config2.LogLevel))

	// You can also parse custom formats supported by slog.Level
	jsonConfig3 := `{"log_level": "INFO+2"}`
	var config3 Config
	if err := json.Unmarshal([]byte(jsonConfig3), &config3); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Parsed log level: %v (numeric: %d)\n", config3.LogLevel, int(config3.LogLevel))

	// Output:
	// Parsed log level: DEBUG (numeric: -4)
	// Parsed TRACE level: TRACE (numeric: -8)
	// Parsed log level: INFO+2 (numeric: 2)
}

// Example showing direct use of UnmarshalText with log.Level
func Example_unmarshalTextDirect() {
	var level log.Level

	// Parse "INFO" level
	if err := level.UnmarshalText([]byte("INFO")); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Level: %v (numeric: %d)\n", level, int(level))

	// Parse custom TRACE level
	var level2 log.Level
	if err := level2.UnmarshalText([]byte("TRACE")); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Level: %v (numeric: %d)\n", level2, int(level2))

	// Parse with offset: "WARN-1"
	var level3 log.Level
	if err := level3.UnmarshalText([]byte("WARN-1")); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Level: %v (numeric: %d)\n", level3, int(level3))

	// Output:
	// Level: INFO (numeric: 0)
	// Level: TRACE (numeric: -8)
	// Level: INFO+3 (numeric: 3)
}
