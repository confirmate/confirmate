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
func NewInMemoryStorage(t *testing.T, joinTables db.CustomJoinTable, types ...interface{}) *db.Storage {
	db, err := db.NewStorage(
		db.WithInMemory(),
		db.WithAutoMigration(types...),
		db.WithSetupJoinTable(joinTables),
	)
	assert.NoError(t, err, "could not create in-memory storage")
	if err != nil {
		t.FailNow()
	}
	return db
}
