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
