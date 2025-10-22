// Copyright 2025 Fraunhofer AISEC:
// This code is licensed under the terms of the Apache License, Version 2.0.
// See the LICENSE file in this project for details.

package orchestrator

import (
	"confirmate.io/core/api/orchestrator"
)

// TODO(all): Decide if we want to keep these types in Orchestrator DB or separate DB for each service. See https://github.com/confirmate/confirmate/issues/74
var types = []any{
	&orchestrator.TargetOfEvaluation{},
	&orchestrator.Certificate{},
	&orchestrator.State{},
	&orchestrator.Catalog{},
	&orchestrator.Category{},
	&orchestrator.Control{},
	&orchestrator.AuditScope{},
}
