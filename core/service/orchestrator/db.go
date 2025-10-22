// Copyright 2025 Fraunhofer AISEC:
// This code is licensed under the terms of the Apache License, Version 2.0.
// See the LICENSE file in this project for details.

package orchestrator

import (
	"confirmate.io/core/api/assessment"
	"confirmate.io/core/api/orchestrator"
	"confirmate.io/core/db"
)

// TODO(all): Decide if we want to keep just these types in Orchestrator DB or add others DB for each service. See https://github.com/confirmate/confirmate/issues/74
var types = []any{
	&orchestrator.TargetOfEvaluation{},
	&orchestrator.Certificate{},
	&orchestrator.State{},
	&orchestrator.Catalog{},
	&orchestrator.Category{},
	&orchestrator.Control{},
	&orchestrator.AuditScope{},
	// TODO(all): The question is if this should be stored here or in the Assessment DB.
	&assessment.MetricConfiguration{},
}

// jointTable defines the MetricConfiguration as a custom join table for gorm
var jointTable = db.CustomJointTable{
	Model:      orchestrator.TargetOfEvaluation{},
	Field:      "ConfiguredMetrics",
	JointTable: assessment.MetricConfiguration{},
}
