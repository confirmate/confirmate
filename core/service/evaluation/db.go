package evaluation

import "confirmate.io/core/api/evaluation"

// types contains all Orchestrator types that we need to auto-migrate into database tables
var types = []any{
	&evaluation.EvaluationResult{},
}
