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
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// defaultMaxConn is the default for maximum number for connections (default: 1) to avoid issues
// with concurrent access to the database.
const defaultMaxConn = 1

// DB is out main database interface that allows to interact with the persistence layer. It is
// closely aligned to the [gorm] operations.
type DB interface {
	// Create attempts to insert the provided record into the database.
	//
	// If a constraint violation occurs, it must return [ErrUniqueConstraintFailed] or
	// [ErrConstraintFailed], depending on the error message.
	Create(r any) (err error)

	// Save attempts to save the given record to the database, applying optional conditions for
	// filtering.
	//
	// If a constraint violation occurs, it must return [ErrConstraintFailed].
	Save(r any, conds ...any) (err error)

	// Update applies the provided changes to the database record, optionally applying conditions
	// for filtering.
	//
	// Must return [ErrConstraintFailed] on a constraint violation or [ErrRecordNotFound] if no
	// matching record is found.
	Update(r any, conds ...any) (err error)

	// Delete attempts to delete the record with the given ID from the database.
	//
	// Must return [ErrRecordNotFound] if no matching record is found.
	Delete(r any, conds ...any) (err error)

	// Get attempts to retrieve a record from the database.
	//
	// If no record is found, it returns [ErrRecordNotFound].
	Get(r any, conds ...any) (err error)

	// List retrieves a list of records from the database.
	List(r any, orderBy string, asc bool, offset int, limit int, conds ...any) (err error)

	// Count retrieves the count of records in the database that match the provided conditions.
	Count(r any, conds ...any) (count int64, err error)

	// Raw executes a raw SQL query and scans the result into the provided destination. Returns an error
	// if the query fails.
	Raw(r any, query string, args ...any) (err error)
}

// db is our main database struct that wraps GORM's DB instance and provides additional
// configuration options.
type db struct {
	*gorm.DB

	// types contain all types that we need to auto-migrate into database tables
	types []any

	// customJoinTables holds configuration for custom join table setups, including model, field, and the join table reference.
	customJoinTables []CustomJoinTable

	// maxConn is the maximum number of connections. 0 means unlimited.
	maxConn int
}

// DBOption defines a function type for configuring the [DB] instance.
type DBOption func(*db)

// WithAutoMigration is an option to add types to GORM's auto-migration.
func WithAutoMigration(types ...any) DBOption {
	return func(s *db) {
		// We append because there can be default types already defined. Currently, we don't have any.
		s.types = append(s.types, types...)
	}
}

// WithInMemory is an option to configure [DB] to use an in-memory DB. This creates a new
// in-memory database each time it is called.
//
// So if you need to have access to the same in-memory DB, you need to share the [DB] instance.
func WithInMemory() DBOption {
	return func(s *db) {
		s.DB, _ = newInMemoryDB()
	}
}

// WithSetupJoinTable is an option to add types to GORM's auto-migration.
func WithSetupJoinTable(joinTables ...CustomJoinTable) DBOption {
	return func(s *db) {
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
	return func(s *db) {
		s.maxConn = max
	}
}

// NewDB creates a new [DB] instance with the provided options.
func NewDB(opts ...DBOption) (s DB, err error) {
	var db = &db{
		maxConn: defaultMaxConn,
	}

	// Add options and/or override default ones
	for _, o := range opts {
		o(db)
	}

	// Open an in-memory database
	if db.DB == nil {
		db.DB, err = newInMemoryDB()
		if err != nil {
			return nil, fmt.Errorf("could not create in-memory db: %w", err)
		}
	}

	// Set max open connections
	if db.maxConn > 0 {
		sqlDB, err := db.DB.DB()
		if err != nil {
			return nil, fmt.Errorf("could not retrieve sql.DB: %v", err)
		}

		sqlDB.SetMaxOpenConns(db.maxConn)
	}

	// Register custom serializers
	schema.RegisterSerializer("durationpb", &DurationSerializer{})
	schema.RegisterSerializer("timestamppb", &TimestampSerializer{})
	schema.RegisterSerializer("valuepb", &ValueSerializer{})
	schema.RegisterSerializer("anypb", &AnySerializer{})

	// Setup custom join tables if any are provided
	for _, jt := range db.customJoinTables {
		if err = db.DB.SetupJoinTable(jt.Model, jt.Field, jt.JoinTable); err != nil {
			err = fmt.Errorf("error during join-table: %w", err)
			return
		}
	}

	// After successful DB initialization, migrate the schema
	if err = db.DB.AutoMigrate(db.types...); err != nil {
		err = fmt.Errorf("error during auto-migration: %w", err)
		return
	}

	s = db

	return
}
