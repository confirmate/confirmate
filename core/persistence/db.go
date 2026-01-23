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

// DB is our main database interface that allows to interact with the persistence layer. It is
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

// gormDB is our main database struct that wraps GORM's DB instance and provides additional
// configuration options.
type gormDB struct {
	*gorm.DB

	cfg  Config
	gcfg gorm.Config
	pcfg postgres.Config
}

// DBOption defines a function type for configuring the [DB] instance.
type DBOption func(*gormDB)

// WithConfig is an option to add types to GORM's auto-migration.
func WithConfig(cfg Config) DBOption {
	return func(s *gormDB) {
		s.cfg = cfg
	}
}

// CustomJoinTable holds the configuration for setting up a custom join table in GORM.
type CustomJoinTable struct {
	Model     any    // The main struct (e.g., &TargetOfEvaluation{})
	Field     string // The Field name in the struct (e.g., "ConfiguredMetrics")
	JoinTable any    // The custom join table struct (e.g., &MetricConfiguration{})
}

// NewDB creates a new [DB] instance with the provided options.
func NewDB(opts ...DBOption) (s DB, err error) {
	var (
		db *gormDB
	)

	db = &gormDB{
		cfg: DefaultConfig,
	}

	// Add options and/or override default ones
	for _, o := range opts {
		o(db)
	}

	// Build the postrges config out of our persistence config
	if db.cfg.InMemoryDB {
		db.pcfg.Conn, err = sql.Open("ramsql", fmt.Sprintf("confirmate_inmemory_%d", rand.Uint64()))
		if err != nil {
			return nil, fmt.Errorf("could not open in-memory database: %w", err)
		}

		// Also limit max connection to 1 for in-memory DB
		db.cfg.MaxConn = 1
	} else {
		db.pcfg.DSN = db.cfg.buildDSN()
	}

	// Set up GORM DB connection
	db.DB, err = gorm.Open(postgres.New(db.pcfg), &db.gcfg)
	if err != nil {
		return nil, fmt.Errorf("could not create gorm connection: %w", err)
	}

	// Set max open connections
	if db.cfg.MaxConn > 0 {
		sqlDB, err := db.DB.DB()
		if err != nil {
			return nil, fmt.Errorf("could not retrieve sql.DB: %v", err)
		}

		sqlDB.SetMaxOpenConns(db.cfg.MaxConn)
	}

	// Register custom serializers
	schema.RegisterSerializer("durationpb", &DurationSerializer{})
	schema.RegisterSerializer("timestamppb", &TimestampSerializer{})
	schema.RegisterSerializer("valuepb", &ValueSerializer{})
	schema.RegisterSerializer("anypb", &AnySerializer{})

	// Setup custom join tables if any are provided
	for _, jt := range db.cfg.CustomJoinTables {
		if err = db.DB.SetupJoinTable(jt.Model, jt.Field, jt.JoinTable); err != nil {
			err = fmt.Errorf("error during join-table: %w", err)
			return
		}
	}

	// After successful DB initialization, migrate the schema
	if err = db.DB.AutoMigrate(db.cfg.Types...); err != nil {
		err = fmt.Errorf("error during auto-migration: %w", err)
		return
	}

	// Run optional init function after migrations
	if db.cfg.InitFunc != nil {
		if err = db.cfg.InitFunc(db); err != nil {
			err = fmt.Errorf("error during init function: %w", err)
			return
		}
	}

	s = db

	return
}
