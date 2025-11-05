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

package persistence

import (
	"testing"
	"time"

	"confirmate.io/core/api/orchestrator"
	"confirmate.io/core/util/testutil/assert"
	"google.golang.org/protobuf/types/known/timestamppb"

	"confirmate.io/core/api/assessment"
	_ "github.com/proullon/ramsql/driver"
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

func Test_DB_Create(t *testing.T) {
	var (
		err    error
		s      *DB
		metric *assessment.Metric
	)

	metric = &assessment.Metric{
		Id:          MockMetricID1,
		Category:    MockMetricCategory1,
		Description: MockMetricDescription1,
		Version:     MockMetricVersion1,
		Comments:    MockMetricComments1,
	}

	// Create DB
	s, err = NewDB(
		WithInMemory(),
		WithAutoMigration(&assessment.Metric{}, &assessment.MetricImplementation{}),
	)
	assert.NoError(t, err)

	err = s.Create(metric)
	assert.NoError(t, err)

	err = s.Create(metric)
	assert.Error(t, err)
}

func Test_DB_Get(t *testing.T) {
	var (
		err    error
		s      *DB
		target *orchestrator.TargetOfEvaluation
	)

	target = &mockToe
	// assert.NoError(t, api.Validate(target))

	// Create DB
	s, err = NewDB(
		WithInMemory(),
		WithAutoMigration(
			&orchestrator.TargetOfEvaluation{},
			&assessment.Metric{},
			&assessment.MetricImplementation{},
		),
	)
	assert.NoError(t, err)

	// Return error since no record in the DB yet
	err = s.Get(&orchestrator.TargetOfEvaluation{})
	assert.ErrorIs(t, err, ErrRecordNotFound)
	_ = target

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
