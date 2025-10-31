// package dbtest provides utilities for testing database operations in Confirmate Core.
package dbtest

import (
	"testing"

	"confirmate.io/core/db"
	"confirmate.io/core/util/testutil/assert"
)

// NewInMemoryStorage creates a new in-memory database storage for testing purposes.
//
// It applies auto-migration for the provided types and sets up the specified join tables.
// If there is an error during the creation of the storage, the test will panic immediately.
func NewInMemoryStorage(t *testing.T, types []any, joinTable db.CustomJoinTable, init ...func(*db.Storage)) *db.Storage {
	opts := []db.StorageOption{
		db.WithInMemory(),
		db.WithAutoMigration(types...),
		db.WithSetupJoinTable(joinTable),
	}
	db, err := db.NewStorage(
		opts...,
	)

	for _, fn := range init {
		fn(db)
	}

	assert.NoError(t, err, "could not create in-memory storage")
	if err != nil {
		t.FailNow()
	}
	return db
}
