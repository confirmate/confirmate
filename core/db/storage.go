// Copyright 2025 Fraunhofer AISEC:
// This code is licensed under the terms of the Apache License, Version 2.0.
// See the LICENSE file in this project for details.

package db

import (
	"database/sql"
	"fmt"

	_ "github.com/proullon/ramsql/driver"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// We set the default for maximum number for connections to 1 to avoid issues with concurrent access to the database.
const defaultMaxConn = 1

type Storage struct {
	*gorm.DB

	// types contain all types that we need to auto-migrate into database tables
	types []any

	inMemoryDB *sql.DB

	// customJointTables holds configuration for custom join table setups, including model, field, and the join table reference.
	customJointTables []CustomJointTable

	// maxConn is the maximum number of connections. 0 means unlimited.
	maxConn int
}

type StorageOption func(*Storage)

// WithAutoMigration is an option to add types to GORM's auto-migration.
func WithAutoMigration(types ...any) StorageOption {
	return func(s *Storage) {
		// We append because there can be default types already defined. Currently, we don't have any.
		s.types = append(s.types, types...)
	}
}

// WithInMemory is an option to configure Storage to use an in memory DB
func WithInMemory() StorageOption {
	return func(s *Storage) {
		s.inMemoryDB, _ = sql.Open("ramsql", "confirmate_inmemory")
	}
}

// WithSetupJoinTable is an option to add types to GORM's auto-migration.
func WithSetupJoinTable(jointTables CustomJointTable) StorageOption {
	return func(s *Storage) {
		s.customJointTables = append(s.customJointTables, jointTables)
	}
}

type CustomJointTable struct {
	Model      any    // The main struct (e.g., &TargetOfEvaluation{})
	Field      string // The Field name in the struct (e.g., "ConfiguredMetrics")
	JointTable any    // The custom join table struct (e.g., &MetricConfiguration{})
}

// WithMaxOpenConns is an option to configure the maximum number of open connections
func WithMaxOpenConns(max int) StorageOption {
	return func(s *Storage) {
		s.maxConn = max
	}
}

func NewStorage(opts ...StorageOption) (s *Storage, err error) {
	s = &Storage{
		maxConn: defaultMaxConn,
	}

	// Add options and/or override default ones
	for _, o := range opts {
		o(s)
	}

	// Open an in-memory database
	ramdb, err := sql.Open("ramsql", "confirmate_inmemory")
	g, err := gorm.Open(postgres.New(postgres.Config{
		Conn: ramdb,
	}),
		&gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("could not open in-memory sqlite database: %w", err)
	}

	s.DB = g

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

	// Setup custom joint tables if any are provided
	for _, jt := range s.customJointTables {
		if err = s.DB.SetupJoinTable(jt.Model, jt.Field, jt.JointTable); err != nil {
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
