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
	"errors"
	"testing"

	"confirmate.io/core/persistence"
	"confirmate.io/core/util/assert"
)

type testRecord struct {
	ID   string `gorm:"primaryKey"`
	Name string
}

func TestErrorDB_Injectors(t *testing.T) {
	const (
		testID    = "test-id"
		testQuery = "SELECT 1"
		testOrder = "id"
	)

	types := []any{&testRecord{}}
	injected := errors.New("injected error")

	tests := []struct {
		name    string
		mk      func(*testing.T) persistence.DB
		run     func(t *testing.T, db persistence.DB) error
		wantErr assert.WantErr
	}{
		{
			name: "CreateErrorDB",
			mk: func(t *testing.T) persistence.DB {
				return CreateErrorDB(t, injected, types, nil)
			},
			run: func(t *testing.T, db persistence.DB) error {
				return db.Create(&testRecord{ID: testID})
			},
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.ErrorIs(t, err, injected)
			},
		},
		{
			name: "SaveErrorDB",
			mk: func(t *testing.T) persistence.DB {
				return SaveErrorDB(t, injected, types, nil)
			},
			run: func(t *testing.T, db persistence.DB) error {
				return db.Save(&testRecord{ID: testID})
			},
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.ErrorIs(t, err, injected)
			},
		},
		{
			name: "UpdateErrorDB",
			mk: func(t *testing.T) persistence.DB {
				return UpdateErrorDB(t, injected, types, nil)
			},
			run: func(t *testing.T, db persistence.DB) error {
				return db.Update(&testRecord{ID: testID})
			},
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.ErrorIs(t, err, injected)
			},
		},
		{
			name: "DeleteErrorDB",
			mk: func(t *testing.T) persistence.DB {
				return DeleteErrorDB(t, injected, types, nil)
			},
			run: func(t *testing.T, db persistence.DB) error {
				return db.Delete(&testRecord{ID: testID})
			},
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.ErrorIs(t, err, injected)
			},
		},
		{
			name: "GetErrorDB",
			mk: func(t *testing.T) persistence.DB {
				return GetErrorDB(t, injected, types, nil)
			},
			run: func(t *testing.T, db persistence.DB) error {
				var rec testRecord
				return db.Get(&rec, testID)
			},
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.ErrorIs(t, err, injected)
			},
		},
		{
			name: "ListErrorDB",
			mk: func(t *testing.T) persistence.DB {
				return ListErrorDB(t, injected, types, nil)
			},
			run: func(t *testing.T, db persistence.DB) error {
				var recs []testRecord
				return db.List(&recs, testOrder, true, 0, 10)
			},
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.ErrorIs(t, err, injected)
			},
		},
		{
			name: "CountErrorDB",
			mk: func(t *testing.T) persistence.DB {
				return CountErrorDB(t, injected, types, nil)
			},
			run: func(t *testing.T, db persistence.DB) error {
				_, err := db.Count(&testRecord{})
				return err
			},
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.ErrorIs(t, err, injected)
			},
		},
		{
			name: "RawErrorDB",
			mk: func(t *testing.T) persistence.DB {
				return RawErrorDB(t, injected, types, nil)
			},
			run: func(t *testing.T, db persistence.DB) error {
				var out []testRecord
				return db.Raw(&out, testQuery)
			},
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.ErrorIs(t, err, injected)
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			db := tc.mk(t)
			err := tc.run(t, db)
			tc.wantErr(t, err)
		})
	}
}

func TestErrorDB_Passthrough(t *testing.T) {
	const (
		testID   = "test-id"
		testName = "alpha"
	)

	types := []any{&testRecord{}}

	db := MultiErrorDB(t,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		types,
		nil,
	)

	createRec := &testRecord{ID: testID, Name: testName}
	assert.NoError(t, db.Create(createRec))

	var got testRecord
	assert.NoError(t, db.Get(&got, "id = ?", testID))
	assert.Equal(t, testID, got.ID)
	assert.Equal(t, testName, got.Name)

	updateRec := &testRecord{ID: testID, Name: "beta"}
	assert.NoError(t, db.Update(updateRec))

	var gotUpdated testRecord
	assert.NoError(t, db.Get(&gotUpdated, "id = ?", testID))
	assert.Equal(t, "beta", gotUpdated.Name)

	saveRec := &testRecord{ID: testID, Name: "gamma"}
	assert.NoError(t, db.Save(saveRec))

	var gotSaved testRecord
	assert.NoError(t, db.Get(&gotSaved, "id = ?", testID))
	assert.Equal(t, "gamma", gotSaved.Name)

	var listed []testRecord
	assert.NoError(t, db.List(&listed, "id", true, 0, 10))
	assert.Equal(t, 1, len(listed))

	count, err := db.Count(&testRecord{})
	assert.NoError(t, err)
	assert.Equal(t, int64(1), count)

	var raw []testRecord
	assert.NoError(t, db.Raw(&raw, "SELECT * FROM test_records WHERE id = ?", testID))
	assert.Equal(t, 1, len(raw))

	assert.NoError(t, db.Delete(&testRecord{}, "id = ?", testID))

	count, err = db.Count(&testRecord{})
	assert.NoError(t, err)
	assert.Equal(t, int64(0), count)
}
