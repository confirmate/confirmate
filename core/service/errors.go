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

package service

import (
	"errors"
	"fmt"

	"confirmate.io/core/persistence"

	"connectrpc.com/connect"
)

// HandleDatabaseError translates database errors into appropriate connect errors.
// If the error is [persistence.ErrRecordNotFound], it returns a [connect.CodeNotFound] error
// with the provided notFoundMsg (or a default message if not provided).
// For other errors, it returns a [connect.CodeInternal] error.
// If err is nil, it returns nil.
func HandleDatabaseError(err error, notFoundErr ...error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, persistence.ErrRecordNotFound) {
		if len(notFoundErr) == 0 {
			notFoundErr = append(notFoundErr, ErrNotFound("entity"))
		}
		return connect.NewError(connect.CodeNotFound, notFoundErr[0])
	}

	return connect.NewError(connect.CodeInternal, fmt.Errorf("database error: %w", err))
}

// ErrNotFound returns a [connect.CodeNotFound] error with the given entity name.
func ErrNotFound(entity string) error {
	return connect.NewError(connect.CodeNotFound, fmt.Errorf("%s not found", entity))
}
