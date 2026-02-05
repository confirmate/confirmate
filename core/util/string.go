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

package util

import (
	"unicode"
)

// CamelCaseToSnakeCase converts a `camelCase` string to `snake_case`
func CamelCaseToSnakeCase(input string) string {
	if input == "" {
		return ""
	}

	snakeCase := make([]rune, 0, len(input))
	runeArray := []rune(input)

	for i := range runeArray {
		if i > 0 && marksNewWord(i, runeArray) {
			snakeCase = append(snakeCase, '_', unicode.ToLower(runeArray[i]))
		} else {
			snakeCase = append(snakeCase, unicode.ToLower(runeArray[i]))
		}
	}

	return string(snakeCase)
}

// marksNewWord checks if the current character starts a new word excluding the first word
func marksNewWord(i int, input []rune) bool {

	if i >= len(input) {
		return false
	}

	// If previous or following rune/character is lowercase or a number then it is a new word
	if i < len(input)-1 && unicode.IsUpper(input[i]) && unicode.IsLower(input[i+1]) {
		return true
	} else if i > 0 && unicode.IsLower(input[i-1]) && unicode.IsUpper(input[i]) {
		return true
	} else if i < len(input)-1 && unicode.IsDigit(input[i]) && unicode.IsLower(input[i+1]) {
		return true
	} else if i > 0 && unicode.IsUpper(input[i-1]) && unicode.IsDigit(input[i]) {
		return true
	}

	return false
}
