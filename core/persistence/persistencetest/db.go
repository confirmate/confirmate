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

package persistencetest

import (
	"testing"

	"confirmate.io/core/persistence"
)

// errorDB is a test-only implementation of the DB interface which allows injecting
// errors for specific operations.
type errorDB struct {
	persistence.DB
	createErr error
	saveErr   error
	updateErr error
	deleteErr error
	getErr    error
	listErr   error
	countErr  error
	rawErr    error
}

// CreateErrorDB returns an ErrorDB that fails on Create with the provided error.
func CreateErrorDB(t *testing.T, err error, types []any, joinTable []persistence.CustomJoinTable, init ...func(persistence.DB)) persistence.DB {
	return &errorDB{
		createErr: err,
		DB:        NewInMemoryDB(t, types, joinTable, init...),
	}
}

// SaveErrorDB returns an ErrorDB that fails on Save with the provided error.
func SaveErrorDB(t *testing.T, err error, types []any, joinTable []persistence.CustomJoinTable, init ...func(persistence.DB)) persistence.DB {
	return &errorDB{
		saveErr: err,
		DB:      NewInMemoryDB(t, types, joinTable, init...),
	}
}

// UpdateErrorDB returns an ErrorDB that fails on Update with the provided error.
func UpdateErrorDB(t *testing.T, err error, types []any, joinTable []persistence.CustomJoinTable, init ...func(persistence.DB)) persistence.DB {
	return &errorDB{
		updateErr: err,
		DB:        NewInMemoryDB(t, types, joinTable, init...),
	}
}

// DeleteErrorDB returns an ErrorDB that fails on Delete with the provided error.
func DeleteErrorDB(t *testing.T, err error, types []any, joinTable []persistence.CustomJoinTable, init ...func(persistence.DB)) persistence.DB {
	return &errorDB{
		deleteErr: err,
		DB:        NewInMemoryDB(t, types, joinTable, init...),
	}
}

// GetErrorDB returns an ErrorDB that fails on Get with the provided error.
func GetErrorDB(t *testing.T, err error, types []any, joinTable []persistence.CustomJoinTable, init ...func(persistence.DB)) persistence.DB {
	return &errorDB{
		getErr: err,
		DB:     NewInMemoryDB(t, types, joinTable, init...),
	}
}

// ListErrorDB returns an ErrorDB that fails on List with the provided error.
func ListErrorDB(t *testing.T, err error, types []any, joinTable []persistence.CustomJoinTable, init ...func(persistence.DB)) persistence.DB {
	return &errorDB{
		listErr: err,
		DB:      NewInMemoryDB(t, types, joinTable, init...),
	}
}

// CountErrorDB returns an ErrorDB that fails on Count with the provided error.
func CountErrorDB(t *testing.T, err error, types []any, joinTable []persistence.CustomJoinTable, init ...func(persistence.DB)) persistence.DB {
	return &errorDB{
		countErr: err,
		DB:       NewInMemoryDB(t, types, joinTable, init...),
	}
}

// RawErrorDB returns an ErrorDB that fails on Raw with the provided error.
func RawErrorDB(t *testing.T, err error, types []any, joinTable []persistence.CustomJoinTable, init ...func(persistence.DB)) persistence.DB {
	return &errorDB{
		rawErr: err,
		DB:     NewInMemoryDB(t, types, joinTable, init...),
	}
}

func (t *errorDB) Create(r any) error {
	if t.createErr != nil {
		return t.createErr
	}

	return t.DB.Create(r)
}

func (t *errorDB) Save(r any, conds ...any) error {
	if t.saveErr != nil {
		return t.saveErr
	}

	return t.DB.Save(r, conds...)
}

func (t *errorDB) Update(r any, conds ...any) error {
	if t.updateErr != nil {
		return t.updateErr
	}

	return t.DB.Update(r, conds...)
}

func (t *errorDB) Delete(r any, conds ...any) error {
	if t.deleteErr != nil {
		return t.deleteErr
	}

	return t.DB.Delete(r, conds...)
}

func (t *errorDB) Get(r any, conds ...any) error {
	if t.getErr != nil {
		return t.getErr
	}

	return t.DB.Get(r, conds...)
}

func (t *errorDB) List(r any, orderBy string, asc bool, offset int, limit int, conds ...any) error {
	if t.listErr != nil {
		return t.listErr
	}

	return t.DB.List(r, orderBy, asc, offset, limit, conds...)
}

func (t *errorDB) Count(r any, conds ...any) (int64, error) {
	if t.countErr != nil {
		return 0, t.countErr
	}

	return t.DB.Count(r, conds...)
}

func (t *errorDB) Raw(r any, query string, args ...any) error {
	if t.rawErr != nil {
		return t.rawErr
	}

	return t.DB.Raw(r, query, args...)
}
