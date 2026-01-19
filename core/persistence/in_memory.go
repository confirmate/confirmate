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

package persistence

import (
	"database/sql"
	"fmt"
	"math/rand/v2"

	_ "github.com/proullon/ramsql/driver"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// newInMemoryDB creates a new in-memory Ramsql database connection.
//
// This creates a unique in-memory database instance each time it is called.
func newInMemoryDB() (g *gorm.DB, err error) {
	var (
		db *sql.DB
	)

	db, err = sql.Open("ramsql", fmt.Sprintf("confirmate_inmemory_%d", rand.Uint64()))
	if err != nil {
		return nil, fmt.Errorf("could not open in-memory database: %w", err)
	}

	g, err = gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}),
		&gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("could not create gorm connection: %w", err)
	}

	return g, nil
}
