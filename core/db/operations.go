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

// package db provides database operations for Confirmate Core. It includes CRUD operations
// and advanced query options using Gorm.
package db

import (
	"errors"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ================================================================================================
// Write Operations
// ================================================================================================

// Create attempts to insert the provided record into the database.
//
// If a constraint violation occurs, it returns [ErrUniqueConstraintFailed] or
// [ErrConstraintFailed], depending on the error message.
func (s *Storage) Create(r any) (err error) {
	err = s.DB.Create(r).Error

	if err != nil && (strings.Contains(err.Error(), "constraint failed: UNIQUE constraint failed") ||
		strings.Contains(err.Error(), "duplicate key value violates unique constraint")) {
		return ErrUniqueConstraintFailed
	}

	if err != nil && strings.Contains(err.Error(), "constraint failed") {
		return ErrConstraintFailed
	}

	return
}

// Save attempts to save the given record to the database, applying optional conditions for
// filtering. If a constraint violation occurs, it returns ErrConstraintFailed.
func (s *Storage) Save(r any, conds ...any) (err error) {
	db := applyWhere(s.DB, conds...).Save(r)
	err = db.Error

	if err != nil && strings.Contains(err.Error(), "constraint failed") {
		return ErrConstraintFailed
	}

	return err
}

// Update applies the provided changes to the database record, optionally applying conditions for
// filtering.
//
// Returns [ErrConstraintFailed] on a constraint violation or [ErrRecordNotFound] if no matching
// record is found.
func (s *Storage) Update(r any, conds ...any) (err error) {
	db := s.DB.Session(&gorm.Session{FullSaveAssociations: true}).Model(r)
	db = applyWhere(db, conds...).Updates(r)
	if err = db.Error; err != nil { // db error
		if strings.Contains(err.Error(), "constraint failed") {
			return ErrConstraintFailed
		} else {
			return err
		}
	}

	// No record with given ID found
	if db.RowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
}

// Delete attempts to delete the record with the given ID from the database.
//
// Returns [ErrRecordNotFound] if no matching record is found.
func (s *Storage) Delete(r any, conds ...any) (err error) {
	// Remove record r with a given ID
	db := s.DB.Delete(r, conds...)
	if err = db.Error; err != nil { // db error
		return err
	}

	// No record with given ID found
	if db.RowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
}

// ================================================================================================
// Read Operations
// ================================================================================================

// Get attempts to retrieve a record from the database.
//
// If no record is found, it returns [ErrRecordNotFound].
func (s *Storage) Get(r any, conds ...any) (err error) {
	// Preload all associations of r if necessary
	db, conds := applyPreload(s.DB, conds...)
	err = db.First(r, conds...).Error

	// if the record is not found, use the error message defined in our package
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = ErrRecordNotFound
	}
	return
}

// List retrieves a list of records from the database.
func (s *Storage) List(r any, orderBy string, asc bool, offset int, limit int, conds ...any) error {
	var (
		db = s.DB
		// Set the default direction to "ascending"
		orderDirection = "asc"
	)

	if limit != -1 {
		db = s.DB.Limit(limit)
	}

	// Set the direction to "descending"
	if !asc {
		orderDirection = "desc"
	}
	orderStmt := orderBy + " " + orderDirection
	// No explicit ordering
	if orderBy == "" {
		orderStmt = ""
	}

	// Preload all associations of r if necessary
	db, conds = applyPreload(db.Offset(offset), conds...)

	return db.Order(orderStmt).Find(r, conds...).Error
}

// Count retrieves the count of records in the database that match the provided conditions.
func (s *Storage) Count(r any, conds ...any) (count int64, err error) {
	db := applyWhere(s.DB.Model(r), conds...)

	err = db.Count(&count).Error
	return
}

// ================================================================================================
// Advanced und customizable operations
// ================================================================================================

// Raw executes a raw SQL query and scans the result into the provided destination. Returns an error
// if the query fails.
func (s *Storage) Raw(r any, query string, args ...any) error {
	return s.DB.Raw(query, args...).Scan(r).Error
}

// ================================================================================================
// Internal Helper Functions
// ================================================================================================

// applyWhere applies the conditional arguments to db.Where. We now basically distinguish between
// three cases:
//   - an empty conditions list means no db.Where function is called
//   - one condition specified means that it is takes as the query parameter.
//     This will query for the specified primary key
//   - otherwise, the first condition will be taken as the query parameter and all others will be
//     taken as additional args.
func applyWhere(db *gorm.DB, conds ...any) *gorm.DB {
	if len(conds) == 0 {
		return db
	} else if len(conds) == 1 {
		return db.Where(conds[0])
	} else {
		return db.Where(conds[0], conds[1:]...)
	}
}

// applyPreload checks for any preload options and prepends them to the DB query. If no extra option
// is specified, [clause.Associations] is used as the default preload.
func applyPreload(db *gorm.DB, conds ...any) (*gorm.DB, []any) {
	if len(conds) > 0 {
		if preload, ok := conds[0].(*preload); ok {
			if preload.query != "" {
				return db.Preload(preload.query, preload.args...), conds[1:]
			} else {
				return db, conds[1:]
			}
		}
	}

	return db.Preload(clause.Associations), conds
}

// ================================================================================================
// Query Options
// ================================================================================================

// QueryOption is a condition that can be passed to the CRUD functions for customizing the query.
type QueryOption interface{}

type preload struct {
	query string
	args  []any
}

// WithPreload allows the customization of Gorm's preload feature with the specified query and
// arguments.
func WithPreload(query string, args ...any) QueryOption {
	return &preload{query: query, args: args}
}

// WithoutPreload disables any kind of preloading of Gorm. This is necessary if custom join tables
// are used, otherwise Gorm will throw errors.
func WithoutPreload() QueryOption {
	return &preload{query: ""}
}
