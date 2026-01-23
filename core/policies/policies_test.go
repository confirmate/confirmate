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
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gopkg.in/yaml.v3"

	"confirmate.io/core/api/assessment"
	"confirmate.io/core/api/evidence"
	"confirmate.io/core/api/ontology"
	"confirmate.io/core/api/orchestrator"
	"confirmate.io/core/util/assert"
)

// Mock resource IDs for testing
const (
	mockObjStorage1EvidenceID = "1"
	mockObjStorage1ResourceID = "/mockresources/storages/object1"
	mockObjStorage2EvidenceID = "2"
	mockObjStorage2ResourceID = "/mockresources/storages/object2"
	mockVM1EvidenceID         = "3"
	mockVM1ResourceID         = "/mockresources/compute/vm1"
	mockVM2EvidenceID         = "4"
	mockVM2ResourceID         = "/mockresources/compute/vm2"
	mockBlockStorage1ID       = "/mockresources/storage/storage1"
	mockBlockStorage2ID       = "/mockresources/storage/storage2"
)

// TestMain provides setup and teardown for all tests in this package
func TestMain(m *testing.M) {
	// TODO: enable cli test helpers again
	// clitest.AutoChdir()

	os.Exit(m.Run())
}

// mockMetricsSource implements the MetricsSource interface for testing.
// It loads metrics from the filesystem.
type mockMetricsSource struct {
	t *testing.T
}

// Ensure mockMetricsSource implements MetricsSource interface
var _ MetricsSource = (*mockMetricsSource)(nil)

// Metrics returns all metrics loaded from the metrics directory
func (m *mockMetricsSource) Metrics() ([]*assessment.Metric, error) {
	metricsPath := "security-metrics/metrics"
	metrics := make([]*assessment.Metric, 0)

	err := filepath.Walk(metricsPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("error accessing path %s: %w", path, err)
		}

		// Skip directories and non-YAML files
		if info.IsDir() || (!strings.HasSuffix(info.Name(), ".yaml") && !strings.HasSuffix(info.Name(), ".yml")) {
			return nil
		}

		// Read the YAML file
		b, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("error reading file %s: %w", path, err)
		}

		var metric assessment.Metric

		dec := yaml.NewDecoder(bytes.NewReader(b))
		if err := dec.Decode(&metric); err != nil {
			return fmt.Errorf("error unmarshalling metric %s: %w", path, err)
		}

		// Set the category automatically, since it is not included in the YAML definition
		metric.Category = filepath.Base(filepath.Dir(filepath.Dir(path)))

		metrics = append(metrics, &metric)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error walking through metrics directory: %w", err)
	}

	return metrics, nil
}

// MetricConfiguration returns the default configuration for a given metric
func (m *mockMetricsSource) MetricConfiguration(targetID string, metric *assessment.Metric) (*assessment.MetricConfiguration, error) {
	// Fetch the metric configuration directly from our file
	bundle := fmt.Sprintf("security-metrics/metrics/%s/%s/data.json", metric.Category, metric.Id)

	b, err := os.ReadFile(bundle)
	assert.NoError(m.t, err)

	var config assessment.MetricConfiguration
	err = protojson.Unmarshal(b, &config)
	assert.NoError(m.t, err)

	config.IsDefault = true
	config.MetricId = metric.Id
	config.TargetOfEvaluationId = targetID

	return &config, nil
}

// MetricImplementation returns the Rego implementation for a given metric
func (m *mockMetricsSource) MetricImplementation(_ assessment.MetricImplementation_Language, metric *assessment.Metric) (*assessment.MetricImplementation, error) {
	// Fetch the metric implementation directly from our file
	bundle := fmt.Sprintf("security-metrics/metrics/%s/%s/metric.rego", metric.Category, metric.Id)

	b, err := os.ReadFile(bundle)
	assert.NoError(m.t, err)

	impl := &assessment.MetricImplementation{
		MetricId: metric.Id,
		Lang:     assessment.MetricImplementation_LANGUAGE_REGO,
		Code:     string(b),
	}

	return impl, nil
}

// updatedMockMetricsSource extends mockMetricsSource with updated configuration values
type updatedMockMetricsSource struct {
	mockMetricsSource
}

// Ensure updatedMockMetricsSource implements MetricsSource interface
var _ MetricsSource = (*updatedMockMetricsSource)(nil)

// MetricConfiguration returns an updated (non-default) configuration
func (u *updatedMockMetricsSource) MetricConfiguration(targetID string, metric *assessment.Metric) (*assessment.MetricConfiguration, error) {
	return &assessment.MetricConfiguration{
		Operator:             "==",
		TargetValue:          structpb.NewBoolValue(false),
		IsDefault:            false,
		UpdatedAt:            timestamppb.New(time.Date(2022, 12, 1, 0, 0, 0, 0, time.Local)),
		MetricId:             metric.Id,
		TargetOfEvaluationId: targetID,
	}, nil
}

// mockPolicyEval implements the PolicyEval interface for testing
type mockPolicyEval struct {
	t       *testing.T
	results []*CombinedResult
	err     error
}

// Ensure mockPolicyEval implements PolicyEval interface
var _ PolicyEval = (*mockPolicyEval)(nil)

// Eval returns pre-configured results
func (m *mockPolicyEval) Eval(evidence *evidence.Evidence, r ontology.IsResource, related map[string]ontology.IsResource, src MetricsSource) ([]*CombinedResult, error) {
	return m.results, m.err
}

// HandleMetricEvent handles metric change events
func (m *mockPolicyEval) HandleMetricEvent(event *orchestrator.ChangeEvent) error {
	return m.err
}

// mockControlsSource implements the ControlsSource interface for testing
type mockControlsSource struct {
	t        *testing.T
	controls []*orchestrator.Control
	err      error
}

// Ensure mockControlsSource implements ControlsSource interface
var _ ControlsSource = (*mockControlsSource)(nil)

// Controls returns pre-configured controls
func (m *mockControlsSource) Controls() ([]*orchestrator.Control, error) {
	return m.controls, m.err
}

// Test functions

func TestCreateKey(t *testing.T) {
	tests := []struct {
		name     string
		evidence *evidence.Evidence
		types    []string
		want     string
	}{
		{
			name: "simple key",
			evidence: &evidence.Evidence{
				ToolId: "tool1",
			},
			types: []string{"TypeA", "TypeB"},
			want:  "TypeA-TypeB-tool1",
		},
		{
			name: "key with spaces in types",
			evidence: &evidence.Evidence{
				ToolId: "tool2",
			},
			types: []string{"Type A", "Type B"},
			want:  "TypeA-TypeB-tool2",
		},
		{
			name: "empty types",
			evidence: &evidence.Evidence{
				ToolId: "tool3",
			},
			types: []string{},
			want:  "tool3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := createKey(tt.evidence, tt.types)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMetricsCache(t *testing.T) {
	cache := &metricsCache{
		m: make(map[string][]*assessment.Metric),
	}

	// Test concurrent writes
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			key := fmt.Sprintf("key-%d", id)
			cache.Lock()
			cache.m[key] = []*assessment.Metric{
				{Id: fmt.Sprintf("metric-%d", id)},
			}
			cache.Unlock()
		}(i)
	}
	wg.Wait()

	// Test concurrent reads
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			key := fmt.Sprintf("key-%d", id)
			cache.RLock()
			metrics := cache.m[key]
			cache.RUnlock()
			assert.NotNil(t, metrics)
		}(i)
	}
	wg.Wait()
}

func TestMockMetricsSource_Metrics(t *testing.T) {
	mock := &mockMetricsSource{t: t}

	metrics, err := mock.Metrics()
	assert.NoError(t, err)
	assert.NotEmpty(t, metrics)
}

func TestMockMetricsSource_MetricConfiguration(t *testing.T) {
	mock := &mockMetricsSource{t: t}

	// Get metrics first
	metrics, err := mock.Metrics()
	assert.NoError(t, err)
	assert.NotEmpty(t, metrics)

	// Get configuration for first metric
	config, err := mock.MetricConfiguration("test-target", metrics[0])
	assert.NoError(t, err)
	assert.NotNil(t, config)
	assert.True(t, config.IsDefault)
}

func TestUpdatedMockMetricsSource_MetricConfiguration(t *testing.T) {
	mock := &updatedMockMetricsSource{
		mockMetricsSource: mockMetricsSource{t: t},
	}

	config, err := mock.MetricConfiguration("test-target", &assessment.Metric{Id: "test-metric"})
	assert.NoError(t, err)
	assert.NotNil(t, config)
	assert.False(t, config.IsDefault)
	assert.Equal(t, "==", config.Operator)
}

func TestMockPolicyEval_Eval(t *testing.T) {
	expectedResults := []*CombinedResult{
		{
			Applicable: true,
			Compliant:  true,
			MetricID:   "test-metric",
		},
	}

	mock := &mockPolicyEval{
		t:       t,
		results: expectedResults,
	}

	results, err := mock.Eval(nil, nil, nil, nil)
	assert.NoError(t, err)
	assert.Equal(t, expectedResults, results)
}

func TestMockControlsSource_Controls(t *testing.T) {
	expectedControls := []*orchestrator.Control{
		{Id: "control-1"},
		{Id: "control-2"},
	}

	mock := &mockControlsSource{
		t:        t,
		controls: expectedControls,
	}

	controls, err := mock.Controls()
	assert.NoError(t, err)
	assert.Equal(t, expectedControls, controls)
}
