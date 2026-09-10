// Copyright 2016-2025 Fraunhofer AISEC
//
// SPDX-License-Identifier: Apache-2.0
//
//	                               /$$$$$$  /$$                                     /$$
//	                             /$$__  $$|__/                                    | $$
//	 /$$$$$$$  /$$$$$$  /$$$$$$$ | $$  \__/ /$$  /$$$$$$  /$$$$$$/$$$$   /$$$$$$  /$$$$$$    /$$$$$$
//	/$$_____/ /$$__  $$| $$__  $$| $$$$    | $$ /$$__  $$| $$_  $$_  $$ |____  $$|_  $$_/   /$$__  $$
//
// | $$      | $$  \ $$| $$  \ $$| $$_/    | $$| $$  \__/| $$ \ $$ \ $$  /$$$$$$$  | $$    | $$$$$$$$
// | $$      | $$  | $$| $$  | $$| $$      | $$| $$      | $$ | $$ | $$ /$$__  $$  | $$ /$$| $$_____/
// |  $$$$$$$|  $$$$$$/| $$  | $$| $$      | $$| $$      | $$ | $$ | $$|  $$$$$$$  |  $$$$/|  $$$$$$$
// \_______/ \______/ |__/  |__/|__/      |__/|__/      |__/ |__/ |__/ \_______/   \___/   \_______/
//
// This file is part of Confirmate Core.

package assert

import "testing"

// Getter is an interface for database operations that can retrieve objects by ID.
type Getter interface {
	Get(dest any, conds ...any) error
	List(dest any, orderBy string, asc bool, pageSize int, page int, conds ...any) error
}

// InDBGet retrieves an object from the database by ID and returns it.
// If the object cannot be retrieved, the test fails and returns the zero value.
func InDBGet[T any](t *testing.T, db Getter, id string) *T {
	t.Helper()

	var obj T
	err := db.Get(&obj, "id = ?", id)
	if !NoError(t, err) {
		return nil
	}

	return &obj
}

// InDBList retrieves a list of objects from the database by ID and returns it.
// If the objects cannot be retrieved, the test fails and returns the zero value.
func InDBList[T any](t *testing.T, db Getter, page_size int) []T {
	t.Helper()

	var obj []T
	err := db.List(&obj, "", true, 0, page_size)
	if !NoError(t, err) {
		return nil
	}

	return obj
}
