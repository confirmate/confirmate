// Copyright 2025 Fraunhofer AISEC:
// This code is licensed under the terms of the Apache License, Version 2.0.
// See the LICENSE file in this project for details.

package db

import (
	"testing"

	"confirmate.io/core/util/testutil/assert"

	"confirmate.io/core/api/assessment"
	_ "github.com/proullon/ramsql/driver"
)

const (
	MockMetricID1          = "Mock Metric 1"
	MockMetricDescription1 = "This is a mock metric"
	MockMetricCategory1    = "Mock Category 1"
	MockMetricVersion1     = "1.0"
	MockMetricComments1    = "Mock metric comments 1"
)

func Test_storage_Create(t *testing.T) {
	var (
		err    error
		s      *Storage
		metric *assessment.Metric
		//target *orchestrator.TargetOfEvaluation
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

	// target = &orchestrator.TargetOfEvaluation{Id: testdata.MockTargetOfEvaluationID1, Name: testdata.MockTargetOfEvaluationName1}
	// err = s.Create(target)
	// assert.NoError(t, err)
	err = s.Create(metric)
	assert.NoError(t, err)

	err = s.Create(metric)
	assert.Error(t, err)
}
