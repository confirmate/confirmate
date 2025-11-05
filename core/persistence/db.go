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
	"gorm.io/gorm/schema"
)

// defaultMaxConn is the default for maximum number for connections (default: 1) to avoid issues
// with concurrent access to the database.
const defaultMaxConn = 1

// DB is our main database struct that wraps GORM's DB instance and provides additional
// configuration options.
type DB struct {
	*gorm.DB

	// types contain all types that we need to auto-migrate into database tables
	types []any

	// customJoinTables holds configuration for custom join table setups, including model, field, and the join table reference.
	customJoinTables []CustomJoinTable

	// maxConn is the maximum number of connections. 0 means unlimited.
	maxConn int
}

// DBOption defines a function type for configuring the [DB] instance.
type DBOption func(*DB)

// WithAutoMigration is an option to add types to GORM's auto-migration.
func WithAutoMigration(types ...any) DBOption {
	return func(s *DB) {
		// We append because there can be default types already defined. Currently, we don't have any.
		s.types = append(s.types, types...)
	}
}

// WithInMemory is an option to configure [DB] to use an in-memory DB. This creates a new
// in-memory database each time it is called.
//
// So if you need to have access to the same in-memory DB, you need to share the [DB] instance.
func WithInMemory() DBOption {
	return func(s *DB) {
		s.DB, _ = newInMemoryDB()
	}
}

// WithSetupJoinTable is an option to add types to GORM's auto-migration.
func WithSetupJoinTable(joinTables ...CustomJoinTable) DBOption {
	return func(s *DB) {
		s.customJoinTables = append(s.customJoinTables, joinTables...)
	}
}

// CustomJoinTable holds the configuration for setting up a custom join table in GORM.
type CustomJoinTable struct {
	Model     any    // The main struct (e.g., &TargetOfEvaluation{})
	Field     string // The Field name in the struct (e.g., "ConfiguredMetrics")
	JoinTable any    // The custom join table struct (e.g., &MetricConfiguration{})
}

// WithMaxOpenConns is an option to configure the maximum number of open connections
func WithMaxOpenConns(max int) DBOption {
	return func(s *DB) {
		s.maxConn = max
	}
}

// NewDB creates a new [DB] instance with the provided options.
func NewDB(opts ...DBOption) (s *DB, err error) {
	s = &DB{
		maxConn: defaultMaxConn,
	}

	// Add options and/or override default ones
	for _, o := range opts {
		o(s)
	}

	// Open an in-memory database
	if s.DB == nil {
		s.DB, err = newInMemoryDB()
		if err != nil {
			return nil, fmt.Errorf("could not create in-memory db: %w", err)
		}
	}

	// Set max open connections
	if s.maxConn > 0 {
		db, err := s.DB.DB()
		if err != nil {
			return nil, fmt.Errorf("could not retrieve sql.DB: %v", err)
		}

		db.SetMaxOpenConns(s.maxConn)
	}

	// Register custom serializers
	schema.RegisterSerializer("durationpb", &DurationSerializer{})
	schema.RegisterSerializer("timestamppb", &TimestampSerializer{})
	schema.RegisterSerializer("valuepb", &ValueSerializer{})
	schema.RegisterSerializer("anypb", &AnySerializer{})

	// Setup custom join tables if any are provided
	for _, jt := range s.customJoinTables {
		if err = s.DB.SetupJoinTable(jt.Model, jt.Field, jt.JoinTable); err != nil {
			err = fmt.Errorf("error during join-table: %w", err)
			return
		}
	}

	// After successful DB initialization, migrate the schema
	if err = s.DB.AutoMigrate(s.types...); err != nil {
		err = fmt.Errorf("error during auto-migration: %w", err)
		return
	}

	return
}

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
