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

package policies

import (
	"log/slog"
	"strings"
	"sync"

	"confirmate.io/core/api/assessment"
	"confirmate.io/core/api/evidence"
	"confirmate.io/core/api/ontology"
	"confirmate.io/core/api/orchestrator"
)

var (
	logger *slog.Logger
)

// metricsCache holds all cached metrics for different combinations of Tools with resource types
type metricsCache struct {
	sync.RWMutex
	// Metrics cached in a map. Key is composed of tool id and resource types concatenation
	m map[string][]*assessment.Metric
}

// PolicyEval is an interface for the policy evaluation engine
type PolicyEval interface {
	// Eval evaluates a given evidence against a metric coming from the metrics source. In order to avoid unnecessary
	// unwrapping, the callee of this function needs to supply the unwrapped ontology resource, since they most likely
	// unwrapped the resource already, e.g. to check for validation.
	Eval(evidence *evidence.Evidence, r ontology.IsResource, related map[string]ontology.IsResource, src MetricsSource) (data []*CombinedResult, err error)
	HandleMetricEvent(event *orchestrator.ChangeEvent) (err error)
}

type CombinedResult struct {
	Applicable bool
	Compliant  bool
	// TODO(oxisto): They are now part of the individual comparison results
	TargetValue interface{}
	// TODO(oxisto): They are now part of the individual comparison results
	Operator   string
	MetricID   string
	MetricName string
	Config     *assessment.MetricConfiguration

	// ComparisonResult is an optional feature to get more infos about the comparisons
	ComparisonResult []*assessment.ComparisonResult

	// Message contains an optional string that the metric can supply to provide a human readable representation of the result
	Message string
}

// MetricsSource is used to retrieve a list of metrics and to retrieve a metric
// configuration as well as implementation for a particular metric (and target of evaluation)
type MetricsSource interface {
	Metrics() ([]*assessment.Metric, error)
	MetricConfiguration(targetID string, metric *assessment.Metric) (*assessment.MetricConfiguration, error)
	MetricImplementation(lang assessment.MetricImplementation_Language, metric *assessment.Metric) (*assessment.MetricImplementation, error)
}

// ControlsSource is used to retrieve a list of controls
type ControlsSource interface {
	Controls() ([]*orchestrator.Control, error)
}

// createKey creates a key by concatenating toolID and all types
func createKey(evidence *evidence.Evidence, types []string) (key string) {
	// Merge toolID and types to one slice and concatenate all its elements
	key = strings.Join(append(types, evidence.ToolId), "-")
	key = strings.ReplaceAll(key, " ", "")
	return
}
