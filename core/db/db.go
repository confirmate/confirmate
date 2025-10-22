// Copyright 2025 Fraunhofer AISEC:
// This code is licensed under the terms of the Apache License, Version 2.0.
// See the LICENSE file in this project for details.

package db

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/proullon/ramsql/driver"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Storage struct {
	// TODO(lebogg): Eventually, make this unexported
	DB *gorm.DB

	// types contain all types that we need to auto-migrate into database tables
	types []any
}

type StorageOption func(*Storage)

// WithAdditionalAutoMigration is an option to add additional types to GORM's auto-migration.
func WithAdditionalAutoMigration(types ...any) StorageOption {
	return func(s *Storage) {
		s.types = append(s.types, types...)
	}
}

func NewStorage(opts ...StorageOption) (*Storage, error) {
	ramdb, err := sql.Open("ramsql", "confirmate_inmemory")

	g, err := gorm.Open(postgres.New(postgres.Config{
		Conn: ramdb,
	}),
		&gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("could not open in-memory sqlite database: %w", err)
	}

	return &Storage{DB: g}, nil
}

func (s *Storage) Create(r any) (err error) {
	err = s.DB.Create(r).Error

	if err != nil && (strings.Contains(err.Error(), "constraint failed: UNIQUE constraint failed") ||
		strings.Contains(err.Error(), "duplicate key value violates unique constraint")) {
		return fmt.Errorf("could not create record: unique constraint failed")
	}

	if err != nil && strings.Contains(err.Error(), "constraint failed") {
		return fmt.Errorf("could not create record: constraint failed")
	}

	return
}

// List retrieves records from the database with optional ordering, offset, limit, and conditions.
func (s *Storage) List(r any, orderBy string, asc bool, offset int, limit int, conds ...any) error {
	var query = s.DB
	// Set default direction to "ascending"
	var orderDirection = "asc"

	if limit != -1 {
		query = s.DB.Limit(limit)
	}

	// Set direction to "descending"
	if !asc {
		orderDirection = "desc"
	}
	orderStmt := orderBy + " " + orderDirection
	// No explicit ordering
	if orderBy == "" {
		orderStmt = ""
	}

	return query.Order(orderStmt).Find(r, conds...).Error
}
