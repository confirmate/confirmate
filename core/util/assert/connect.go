// Copyright 2016-2025 Fraunhofer AISEC
//
// SPDX-License-Identifier: Apache-2.0
//
//	                               /$$$$$$  /$$                                     /$$
//	                             /$$__  $$|__/                                    | $$
//	 /$$$$$$$  /$$$$$$  /$$$$$$$ | $$  \__/ /$$  /$$$$$$  /$$$$$$/$$$$   /$$$$$$  /$$$$$$    /$$$$$$
//	/$$_____/ /$$__  $$| $$__  $$| $$$$    | $$ /$$__  $$| $$_  $$_  $$ |____  $$|_  $$_/   /$$__  $$
//
// | $$      | $$  \ $$| $$  \ $$| $$_/    | $$| $$  \__/| $$ \ $$ \ $$  /$$$$$$$  | $$    | $$$$$$$$
// | $$      | $$  | $$| $$  | $$| $$      | $$| $$      | $$ | $$ | $$ /$$__  $$  | $$ /$$| $$_____/
// |  $$$$$$$|  $$$$$$/| $$  | $$| $$      | $$| $$      | $$ | $$ | $$|  $$$$$$$  |  $$$$/|  $$$$$$$
// \_______/ \______/ |__/  |__/|__/      |__/|__/      |__/ |__/ |__/ \_______/   \___/   \_______/
//
// This file is part of Confirmate Core.

package assert

import (
	"errors"
	"testing"

	"buf.build/go/protovalidate"
	"connectrpc.com/connect"
)

// IsConnectError asserts that err is a *[connect.Error] and has the specified code.
// Returns true if the assertion passes.
func IsConnectError(t TestingT, err error, code connect.Code) bool {
	tt, ok := t.(*testing.T)
	if ok {
		tt.Helper()
	}

	cErr, ok := err.(*connect.Error)
	if !ok {
		return Fail(t, "Error is not a connect.Error", "Expected: *connect.Error\nActual: %T", err)
	}

	return Equal(t, code, cErr.Code())
}

// IsValidationError asserts that err contains a [protovalidate.ValidationError] with a violation for the specified field.
//
// This is useful for testing that validation errors mention the specific field that failed.
// The field name is matched against the field paths in the validation violations.
// Returns true if the assertion passes.
func IsValidationError(t TestingT, err error, field string) bool {
	tt, ok := t.(*testing.T)
	if ok {
		tt.Helper()
	}

	// Try to unwrap to get the protovalidate.ValidationError
	var validationErr *protovalidate.ValidationError
	if !errors.As(err, &validationErr) {
		return Fail(t, "Error does not contain protovalidate.ValidationError", "Error: %v", err)
	}

	// Check if any violation mentions the expected field
	var availableFields []string
	for _, violation := range validationErr.Violations {
		fieldPath := protovalidate.FieldPathString(violation.Proto.GetField())
		if fieldPath != "" {
			availableFields = append(availableFields, fieldPath)
		}
		if fieldPath == field {
			return true
		}
	}

	return Fail(t, "Validation error does not include expected field", "Expected field: %s\nAvailable fields: %v", field, availableFields)
}
