package evidence

import (
	"confirmate.io/core/api/evidence"
	"confirmate.io/core/persistence/models"
)

var types = []any{
	&evidence.Evidence{},
	&evidence.Resource{},
	&models.EvidencePermission{},
}
