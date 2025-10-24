// Copyright 2025 Fraunhofer AISEC:
// This code is licensed under the terms of the Apache License, Version 2.0.
// See the LICENSE file in this project for details.

package db

import (
	"testing"
	"time"

	"confirmate.io/core/api/orchestrator"
	"confirmate.io/core/util/testutil/assert"

	"confirmate.io/core/api/assessment"
	_ "github.com/proullon/ramsql/driver"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	MockMetricID1          = "Mock Metric 1"
	MockMetricDescription1 = "This is a mock metric"
	MockMetricCategory1    = "Mock Category 1"
	MockMetricVersion1     = "1.0"
	MockMetricComments1    = "Mock metric comments 1"

	MockTargetOfEvaluationID1          = "Mock TOE 1"
	MockTargetOfEvaluationName1        = "Mock TOE Name 1"
	MockTargetOfEvaluationDescription1 = "This is a mock TOE description 1"
)

var mockToe = orchestrator.TargetOfEvaluation{
	Id:                MockTargetOfEvaluationID1,
	Name:              MockTargetOfEvaluationName1,
	Description:       MockTargetOfEvaluationDescription1,
	ConfiguredMetrics: []*assessment.Metric{},
}

func Test_storage_Create(t *testing.T) {
	var (
		err    error
		s      *Storage
		metric *assessment.Metric
	)

	metric = &assessment.Metric{
		Id:          MockMetricID1,
		Category:    MockMetricCategory1,
		Description: MockMetricDescription1,
		Version:     MockMetricVersion1,
		Comments:    MockMetricComments1,
	}

	// Create storage
	s, err = NewStorage(
		WithInMemory(),
		WithAutoMigration(&assessment.Metric{}),
	)
	assert.NoError(t, err)

	err = s.Create(metric)
	assert.NoError(t, err)

	err = s.Create(metric)
	assert.Error(t, err)
}

func Test_storage_Get(t *testing.T) {
	var (
		err    error
		s      *Storage
		target *orchestrator.TargetOfEvaluation
	)

	target = &mockToe
	// assert.NoError(t, api.Validate(target))

	// Create storage
	s, err = NewStorage(
		WithInMemory(),
		WithAutoMigration(
			&orchestrator.TargetOfEvaluation{},
			&assessment.Metric{},
		),
	)
	assert.NoError(t, err)

	// Return error since no record in the DB yet
	err = s.Get(&orchestrator.TargetOfEvaluation{})
	assert.ErrorIs(t, err, ErrRecordNotFound)

	// Create target of evaluation
	err = s.Create(target)
	assert.NoError(t, err)

	// Get target of evaluation via passing entire record
	gotTarget := &orchestrator.TargetOfEvaluation{}
	err = s.Get(&gotTarget)
	assert.NoError(t, err)
	assert.Equal(t, target, gotTarget)

	// Get target of evaluation via name
	gotTarget2 := &orchestrator.TargetOfEvaluation{}
	err = s.Get(&gotTarget2, "name = ?", target.Name)
	assert.NoError(t, err)
	assert.Equal(t, target, gotTarget2)

	// Get target of evaluation via description
	gotTarget3 := &orchestrator.TargetOfEvaluation{}
	err = s.Get(&gotTarget3, "description = ?", target.Description)
	assert.NoError(t, err)
	// assert.NoError(t, api.Validate(gotTarget3))
	assert.Equal(t, target, gotTarget3)

	var metric = &assessment.Metric{
		Id:          MockMetricID1,
		Category:    MockMetricCategory1,
		Description: MockMetricDescription1,
		Version:     MockMetricVersion1,
		Comments:    MockMetricComments1,
	}
	// Check if metric has all necessary fields
	// assert.NoError(t, api.Validate(metric))

	// Create metric
	err = s.Create(metric)
	assert.NoError(t, err)

	// Get metric via Id
	gotMetric := &assessment.Metric{}
	err = s.Get(gotMetric, "id = ?", MockMetricID1)
	assert.NoError(t, err)
	//assert.NoError(t, api.Validate(gotMetric))
	assert.Equal(t, metric, gotMetric)

	var impl = &assessment.MetricImplementation{
		MetricId:  MockMetricID1,
		Code:      "TestCode",
		UpdatedAt: timestamppb.New(time.Date(2000, 1, 1, 1, 1, 1, 1, time.UTC)),
	}
	// Check if impl has all necessary fields
	// assert.NoError(t, api.Validate(impl))

	// Create metric implementation
	err = s.Create(impl)
	assert.NoError(t, err)

	// Get metric implementation via Id
	gotImpl := &assessment.MetricImplementation{}
	err = s.Get(gotImpl, "metric_id = ?", MockMetricID1)
	assert.NoError(t, err)
	// assert.NoError(t, api.Validate(gotImpl))
	assert.Equal(t, impl, gotImpl)
}
