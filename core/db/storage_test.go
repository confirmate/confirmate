// Copyright 2025 Fraunhofer AISEC:
// This code is licensed under the terms of the Apache License, Version 2.0.
// See the LICENSE file in this project for details.

package db

import (
	"testing"

	"confirmate.io/core/api"

	"confirmate.io/util/testutil/assert"

	"confirmate.io/util/testdata"

	"confirmate.io/core/api/assessment"
	_ "github.com/proullon/ramsql/driver"
)

func Test_storage_Create(t *testing.T) {
	var (
		err    error
		s      *Storage
		metric *assessment.Metric
		//target *orchestrator.TargetOfEvaluation
	)

	metric = &assessment.Metric{
		Id:          testdata.MockMetricID1,
		Category:    testdata.MockMetricCategory1,
		Description: testdata.MockMetricDescription1,
		Version:     testdata.MockMetricVersion1,
		Comments:    testdata.MockMetricComments1,
	}
	// Check if metric has all necessary fields
	assert.NoError(t, api.Validate(metric))

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
