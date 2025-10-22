// Copyright 2025 Fraunhofer AISEC:
// This code is licensed under the terms of the Apache License, Version 2.0.
// See the LICENSE file in this project for details.

package db

import "errors"

var (
	ErrRecordNotFound         = errors.New("record not in the database")
	ErrConstraintFailed       = errors.New("constraint failed")
	ErrUniqueConstraintFailed = errors.New("unique constraint failed")
	ErrUnsupportedType        = errors.New("unsupported type")
	ErrDatabase               = errors.New("database error")
	ErrEntryAlreadyExists     = errors.New("entry already exists")
)
