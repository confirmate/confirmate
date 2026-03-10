// Copyright 2016-2026 Fraunhofer AISEC
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

package collection_test

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"confirmate.io/core/api/ontology"
	"confirmate.io/core/service/collection"
	"confirmate.io/core/service/collection/collectiontest"
	"confirmate.io/core/util/assert"
)

func TestRunOnce_CollectsIndividualErrors(t *testing.T) {
	var (
		errBoom error
		svc     *collection.Service
		err     error
		res     collection.CollectionResult
	)

	errBoom = errors.New("boom")

	svc, err = collection.NewService(collection.Config{
		Collectors: []collection.Collector{
			collectiontest.NewFunctionCollector(nil),
			collectiontest.NewFunctionCollector(func() ([]ontology.IsResource, error) {
				return nil, errBoom
			}),
		},
	})
	assert.NoError(t, err)

	res = svc.RunOnce()

	assert.Equal(t, 2, len(res.Collectors))

	var (
		foundError bool
	)

	for _, collectorResult := range res.Collectors {
		if collectorResult.Err != nil {
			foundError = true
			assert.ErrorIs(t, collectorResult.Err, errBoom)
		}
	}

	assert.True(t, foundError)
	assert.False(t, res.FinishedAt.Before(res.StartedAt))
}

func TestStart_RunsPeriodicallyAndDoesNotStopOnCollectorError(t *testing.T) {
	var (
		ctx               context.Context
		cancel            context.CancelFunc
		svc               *collection.Service
		err               error
		resultCh          <-chan collection.CollectionResult
		collectedRuns     int
		successCalls      atomic.Int32
		mu                sync.Mutex
		failingCalledOnce bool
	)

	ctx, cancel = context.WithCancel(context.Background())
	defer cancel()

	svc, err = collection.NewService(collection.Config{
		Interval: 20 * time.Millisecond,
		Collectors: []collection.Collector{
			collectiontest.NewFunctionCollector(func() ([]ontology.IsResource, error) {
				successCalls.Add(1)
				return nil, nil
			}),
			collectiontest.NewFunctionCollector(func() ([]ontology.IsResource, error) {
				mu.Lock()
				failingCalledOnce = true
				mu.Unlock()
				return nil, errors.New("expected failure")
			}),
		},
	})
	assert.NoError(t, err)

	resultCh = svc.Start(ctx)

	for collectedRuns < 2 {
		select {
		case <-time.After(500 * time.Millisecond):
			t.Fatal("timed out waiting for collection runs")
		case result, ok := <-resultCh:
			if !ok {
				t.Fatal("result channel closed too early")
			}

			assert.Equal(t, 2, len(result.Collectors))
			collectedRuns++
		}
	}

	cancel()

	assert.True(t, successCalls.Load() >= 2)
	mu.Lock()
	assert.True(t, failingCalledOnce)
	mu.Unlock()
}
