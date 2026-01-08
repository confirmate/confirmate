package evidence

import (
	"testing"

	"gorm.io/gorm"

	"confirmate.io/core/persistence"
	"confirmate.io/core/persistence/persistencetest"
	"confirmate.io/core/util/assert"
)

// NewTestService creates a new test service not throwing errors
func NewTestService(t *testing.T, doSomething func(service *Service)) (svc *Service) {
	var (
		err error
	)

	svc, err = NewService(
		WithDB(persistencetest.NewInMemoryDB(t, types, nil)))
	if err != nil {
		assert.NoError(t, err)
	}
	if doSomething != nil {
		doSomething(svc)
	}
	return
}

// NewTestServiceWithErrors creates a new test service with a DB configured to return errors as specified by storageWithErrors.
// TODO(lebogg/oxisto): Because we have no interface anymore, we cannot just use a custom Error test DB. Therefore, I
// injected the error callbacks into the DB. This is not ideal, but works.
func NewTestServiceWithErrors(t *testing.T, storageWithErrors *StorageWithError) (svc *Service) {
	var (
		err error
	)

	db := persistencetest.NewInMemoryDB(t, types, nil, func(db *persistence.DB) {
		registerDBErrorCallbacks(db, storageWithErrors)
	})

	svc, err = NewService(
		WithDB(db))
	if err != nil {
		assert.NoError(t, err)
	}
	return
}

// StorageWithError can be used to introduce various errors in a storage operation during unit testing.
type StorageWithError struct {
	// Generic errors to inject per operation kind
	CreateErr error
	SaveErr   error
	UpdateErr error
	GetErr    error
	ListErr   error
	RawErr    error
	CountRes  int64
	CountErr  error
	DeleteErr error

	// Optional scoping by GORM schema (struct) name, e.g. "Evidence" or "Resource".
	// If empty, the error applies to all models for that operation.
	FailOnCreateSchema string
	FailOnUpdateSchema string
	FailOnDeleteSchema string
	FailOnQuerySchema  string
	FailOnRawSchema    string
}

// registerDBErrorCallbacks registers GORM callbacks to inject errors for specific operations.
func registerDBErrorCallbacks(db *persistence.DB, e *StorageWithError) {
	if db == nil || e == nil {
		return
	}
	g := db.DB

	// helper to test schema match
	matchSchema := func(tx *gorm.DB, target string) bool {
		if target == "" {
			return true
		}
		if tx != nil && tx.Statement != nil && tx.Statement.Schema != nil {
			return tx.Statement.Schema.Name == target
		}
		return false
	}

	// Create errors
	if e.CreateErr != nil {
		_ = g.Callback().Create().Before("gorm:create").Register("test_create_error", func(tx *gorm.DB) {
			if matchSchema(tx, e.FailOnCreateSchema) {
				_ = tx.AddError(e.CreateErr)
			}
		})
	}

	// Update/Save errors: prefer explicit SaveErr if set; otherwise use UpdateErr
	if e.SaveErr != nil || e.UpdateErr != nil {
		errToUse := e.UpdateErr
		if e.SaveErr != nil {
			errToUse = e.SaveErr
		}
		_ = g.Callback().Update().Before("gorm:update").Register("test_update_error", func(tx *gorm.DB) {
			if matchSchema(tx, e.FailOnUpdateSchema) {
				_ = tx.AddError(errToUse)
			}
		})
	}

	// Delete errors
	if e.DeleteErr != nil {
		_ = g.Callback().Delete().Before("gorm:delete").Register("test_delete_error", func(tx *gorm.DB) {
			if matchSchema(tx, e.FailOnDeleteSchema) {
				_ = tx.AddError(e.DeleteErr)
			}
		})
	}

	// Read/query path: These are coarse-grained and will affect most SELECTs.
	// Only register if explicitly requested to avoid interfering with unrelated tests.
	if e.GetErr != nil || e.ListErr != nil || e.CountErr != nil || e.RawErr != nil {
		_ = g.Callback().Query().Before("gorm:query").Register("test_query_error", func(tx *gorm.DB) {
			if !matchSchema(tx, e.FailOnQuerySchema) {
				return
			}
			// Priority: RawErr > CountErr > GetErr > ListErr (raw often bypasses this callback but keep for completeness)
			if e.RawErr != nil {
				_ = tx.AddError(e.RawErr)
				return
			}
			if e.CountErr != nil {
				_ = tx.AddError(e.CountErr)
				return
			}
			if e.GetErr != nil {
				_ = tx.AddError(e.GetErr)
				return
			}
			if e.ListErr != nil {
				_ = tx.AddError(e.ListErr)
				return
			}
		})

		// Raw queries have their own callback chain; register separately if RawErr is set
		if e.RawErr != nil {
			_ = g.Callback().Raw().Before("gorm:raw").Register("test_raw_error", func(tx *gorm.DB) {
				if matchSchema(tx, e.FailOnRawSchema) {
					_ = tx.AddError(e.RawErr)
				}
			})
		}
	}
}
