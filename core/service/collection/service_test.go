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

	"confirmate.io/core/api/evidence"
	"confirmate.io/core/api/ontology"
	"confirmate.io/core/service/collection"
	"confirmate.io/core/service/collection/collectiontest"
	"confirmate.io/core/util/assert"

	"connectrpc.com/connect"
	"github.com/google/uuid"
)

type mockEvidenceStoreClient struct {
	storeEvidence func(context.Context, *connect.Request[evidence.StoreEvidenceRequest]) (*connect.Response[evidence.StoreEvidenceResponse], error)
}

func (c mockEvidenceStoreClient) StoreEvidence(ctx context.Context, req *connect.Request[evidence.StoreEvidenceRequest]) (*connect.Response[evidence.StoreEvidenceResponse], error) {
	if c.storeEvidence != nil {
		return c.storeEvidence(ctx, req)
	}

	return connect.NewResponse(&evidence.StoreEvidenceResponse{}), nil
}

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
			collectiontest.NewNamedFunctionCollector("collector-ok", nil),
			collectiontest.NewNamedFunctionCollector("collector-fail", func() ([]ontology.IsResource, error) {
				return nil, errBoom
			}),
		},
	})
	assert.NoError(t, err)

	res = svc.RunOnce()

	assert.Equal(t, 2, len(res.Collectors))
	assert.Equal(t, "collector-ok", res.Collectors[0].CollectorName)
	assert.Equal(t, "collector-fail", res.Collectors[1].CollectorName)
	assert.NotEqual(t, "", res.Collectors[0].CollectorID)
	assert.NotEqual(t, "", res.Collectors[1].CollectorID)
	assert.NotEqual(t, res.Collectors[0].CollectorID, res.Collectors[1].CollectorID)
	assert.Nil(t, res.Collectors[0].Err)
	assert.ErrorIs(t, res.Collectors[1].Err, errBoom)

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

func TestRunOnce_ForwardsCollectedResourcesToEvidenceStore(t *testing.T) {
	var (
		targetOfEvaluationID string
		storedEvidence       []*evidence.Evidence
		mu                   sync.Mutex
		svc                  *collection.Service
		err                  error
		res                  collection.CollectionResult
	)

	targetOfEvaluationID = uuid.NewString()

	svc, err = collection.NewService(collection.Config{
		Collectors: []collection.Collector{
			collectiontest.NewNamedFunctionCollector("collector-ok", func() ([]ontology.IsResource, error) {
				return []ontology.IsResource{
					&ontology.VirtualMachine{Id: "vm-1"},
				}, nil
			}),
		},
		TargetOfEvaluationID: targetOfEvaluationID,
		EvidenceStoreClient: mockEvidenceStoreClient{
			storeEvidence: func(_ context.Context, req *connect.Request[evidence.StoreEvidenceRequest]) (*connect.Response[evidence.StoreEvidenceResponse], error) {
				mu.Lock()
				storedEvidence = append(storedEvidence, req.Msg.GetEvidence())
				mu.Unlock()

				return connect.NewResponse(&evidence.StoreEvidenceResponse{}), nil
			},
		},
	})
	assert.NoError(t, err)

	res = svc.RunOnce()

	assert.Equal(t, 1, len(res.Collectors))
	assert.Nil(t, res.Collectors[0].Err)

	mu.Lock()
	assert.Equal(t, 1, len(storedEvidence))
	assert.Equal(t, targetOfEvaluationID, storedEvidence[0].TargetOfEvaluationId)
	assert.Equal(t, "vm-1", storedEvidence[0].Resource.GetVirtualMachine().GetId())
	assert.NotEqual(t, "", storedEvidence[0].ToolId)
	assert.NotEqual(t, "", storedEvidence[0].Id)
	mu.Unlock()
}

func TestNewService_ReturnsError_WhenEvidenceForwardingEnabledWithoutTargetOfEvaluationID(t *testing.T) {
	var (
		svc *collection.Service
		err error
	)

	svc, err = collection.NewService(collection.Config{
		Collectors: []collection.Collector{
			collectiontest.NewFunctionCollector(nil),
		},
		EvidenceStoreClient: mockEvidenceStoreClient{},
	})

	assert.Nil(t, svc)
	assert.ErrorContains(t, err, "target of evaluation id must be set")
}
