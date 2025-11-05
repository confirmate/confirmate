// package dbtest provides utilities for testing database operations in Confirmate Core.
package dbtest

import (
	"testing"

	"confirmate.io/core/persistence"
	"confirmate.io/core/util/testutil/assert"
)

// NewInMemoryDB creates a new in-memory database for testing purposes.
//
// It applies auto-migration for the provided types and sets up the specified join tables.
// If there is an error during the creation of the DB, the test will panic immediately.
func NewInMemoryDB(t *testing.T, types []any, joinTable persistence.CustomJoinTable, init ...func(*persistence.DB)) *persistence.DB {
	opts := []persistence.DBOption{
		persistence.WithInMemory(),
		persistence.WithAutoMigration(types...),
		persistence.WithSetupJoinTable(joinTable),
	}
	db, err := persistence.NewDB(
		opts...,
	)

	for _, fn := range init {
		fn(db)
	}

	assert.NoError(t, err, "could not create in-memory db")
	if err != nil {
		t.FailNow()
	}
	return db
}
