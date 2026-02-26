package evaluation

import (
	"confirmate.io/core/api/evaluation"
)

// types contains all Evaluation types that we need to auto-migrate into database tables
var types = []any{
	&evaluation.EvaluationResult{},
}
