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

// package persistencetest provides utilities for testing database operations in Confirmate Core.
package persistencetest

import (
	"testing"

	"confirmate.io/core/persistence"
	"confirmate.io/core/util/assert"
)

// NewInMemoryDB creates a new in-memory database for testing purposes.
//
// It applies auto-migration for the provided types and sets up the specified join tables.
// If there is an error during the creation of the DB, the test will panic immediately.
func NewInMemoryDB(t *testing.T, types []any, joinTable persistence.CustomJoinTable, init ...func(*persistence.DB)) *persistence.DB {
	opts := []persistence.DBOption{
		persistence.WithInMemory(),
		persistence.WithAutoMigration(types...),
		persistence.WithSetupJoinTable(joinTable),
	}
	db, err := persistence.NewDB(
		opts...,
	)

	for _, fn := range init {
		fn(db)
	}

	if !assert.NoError(t, err, "could not create in-memory db") {
		panic(err)
	}

	return db
}
