package evidence

import (
	"testing"

	"confirmate.io/core/persistence/persistencetest"
)

// NewTestService creates a new test service not throwing errors
func NewTestService(t *testing.T) (svc *Service) {
	var (
		err error
	)

	svc, err = NewService(
		WithDB(persistencetest.NewInMemoryDB(t, types, nil)))
	if err != nil {
		panic(err)
	}
	return
}
