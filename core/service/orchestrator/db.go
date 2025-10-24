// Copyright 2025 Fraunhofer AISEC:
// This code is licensed under the terms of the Apache License, Version 2.0.
// See the LICENSE file in this project for details.

package orchestrator

import (
	"confirmate.io/core/api/assessment"
	"confirmate.io/core/api/orchestrator"
	"confirmate.io/core/db"
)

// TODO: Add other services' types
var types = []any{
	&orchestrator.TargetOfEvaluation{},
	&orchestrator.Certificate{},
	&orchestrator.State{},
	&orchestrator.Catalog{},
	&orchestrator.Category{},
	&orchestrator.Control{},
	&orchestrator.AuditScope{},
	&assessment.MetricConfiguration{},
	&assessment.Metric{},
}

// jointTable defines the MetricConfiguration as a custom join table for gorm
var jointTable = db.CustomJointTable{
	Model:      orchestrator.TargetOfEvaluation{},
	Field:      "ConfiguredMetrics",
	JointTable: assessment.MetricConfiguration{},
}
