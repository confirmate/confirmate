package evidencetest

import (
	"confirmate.io/core/api/evidence"
	"confirmate.io/core/persistence"
	"github.com/google/uuid"
)

var (
	MockEvidence1 = &evidence.Evidence{
		Id: uuid.NewString(),
	}
)
var InitDBWithEvidence = func(db *persistence.DB) {
	err := db.Create(MockEvidence1)
	if err != nil {
		panic(err)
	}
}
