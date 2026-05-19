// Copyright 2016-2025 Fraunhofer AISEC
//
// SPDX-License-Identifier: Apache-2.0
//
//                                 /$$$$$$  /$$                                     /$$
//                               /$$__  $$|__/                                    | $$
//   /$$$$$$$  /$$$$$$  /$$$$$$$ | $$  \__/ /$$  /$$$$$$  /$$$$$$/$$$$   /$$$$$$  /$$$$$$    /$$$$$$
//  /$$_____/ /$$__  $$| $$__  $$| $$$$    | $$ /$$__  $$| $$_  $$_  $$ |____  $$|_  $$_/   /$$__  $$
// | $$      | $$  \ $$| $$  \ $$| $$_/    | $$| $$  \__/| $$ \ $$ \ $$  /$$$$$$$  | $$    | $$$$$$$$
// | $$      | $$  | $$| $$  | $$| $$      | $$| $$      | $$ | $$ | $$ /$$__  $$  | $$ /$$| $$_____/
// |  $$$$$$$|  $$$$$$/| $$  | $$| $$      | $$| $$      | $$ | $$ | $$|  $$$$$$$  |  $$$$/|  $$$$$$$
// \_______/ \______/ |__/  |__/|__/      |__/|__/      |__/ |__/ |__/ \_______/   \___/   \_______/
//
// This file is part of Confirmate Core.

package orchestrator

import (
	"confirmate.io/core/api/assessment"
	"confirmate.io/core/api/evaluation"
	"confirmate.io/core/api/orchestrator"
	"confirmate.io/core/persistence"
)

// types contains all Orchestrator types that we need to auto-migrate into database tables.
// Order matters: types must appear before any other type that holds a FK or many2many
// constraint referencing their table.
var types = []any{
	// No outgoing FK dependencies — must come first.
	&orchestrator.User{},
	&orchestrator.UserPermission{},
	&assessment.Metric{},
	&assessment.MetricImplementation{},
	&orchestrator.TargetOfEvaluation{},
	&orchestrator.Certificate{},
	&orchestrator.State{},
	&orchestrator.Catalog{},
	// Control depends on Metric (control_metrics join table).
	// Category depends on Control (category_controls join table).
	&orchestrator.Control{},
	&orchestrator.Category{},
	&orchestrator.AuditScope{},
	&orchestrator.AssessmentTool{},
	&assessment.MetricConfiguration{},
	&assessment.AssessmentResult{},
	&evaluation.EvaluationResult{},
	// ControlInScope depends on AuditScope and Control.
	// AuditTrailEvent depends on AuditScope.
	&orchestrator.ControlInScope{},
	&orchestrator.AuditTrailEvent{},
}

// joinTables defines the [MetricConfiguration] as a custom join table between
// [orchestrator.TargetOfEvaluation] and [assessment.Metric].
var joinTables = []persistence.CustomJoinTable{
	{
		Model:     orchestrator.TargetOfEvaluation{},
		Field:     "ConfiguredMetrics",
		JoinTable: assessment.MetricConfiguration{},
	},
}
