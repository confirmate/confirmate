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
	"io"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"confirmate.io/core/api/evidence"
	"confirmate.io/core/api/evidence/evidenceconnect"
	"confirmate.io/core/api/ontology"
	"confirmate.io/core/server"
	"confirmate.io/core/server/servertest"
	"confirmate.io/core/service/collection"
	"confirmate.io/core/service/collection/collectiontest"
	"confirmate.io/core/util/assert"

	"connectrpc.com/connect"
	"github.com/google/uuid"
)

type mockEvidenceStoreHandler struct {
	evidenceconnect.UnimplementedEvidenceStoreHandler

	mu       sync.Mutex
	requests []*evidence.StoreEvidenceRequest

	responseFunc func(*evidence.StoreEvidenceRequest) (*evidence.StoreEvidencesResponse, error)
}

func (h *mockEvidenceStoreHandler) StoreEvidences(_ context.Context, stream *connect.BidiStream[evidence.StoreEvidenceRequest, evidence.StoreEvidencesResponse]) (err error) {
	var (
		req *evidence.StoreEvidenceRequest
		res *evidence.StoreEvidencesResponse
	)

	for {
		req, err = stream.Receive()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}

		h.mu.Lock()
		h.requests = append(h.requests, req)
		h.mu.Unlock()

		if h.responseFunc != nil {
			res, err = h.responseFunc(req)
			if err != nil {
				return err
			}
		} else {
			res = &evidence.StoreEvidencesResponse{Status: evidence.EvidenceStatus_EVIDENCE_STATUS_OK}
		}

		err = stream.Send(res)
		if err != nil {
			return err
		}
	}
}

func (h *mockEvidenceStoreHandler) Requests() (requests []*evidence.StoreEvidenceRequest) {
	h.mu.Lock()
	defer h.mu.Unlock()

	requests = append([]*evidence.StoreEvidenceRequest(nil), h.requests...)
	return requests
}

func TestRunOnce_CollectsIndividualErrors(t *testing.T) {
	var (
		errBoom error
		svc     *collection.Service
		err     error
		res     collection.CollectionResult
	)

	errBoom = errors.New("boom")

	svc, err = collection.NewService(
		collection.WithConfig(collection.Config{
			Collectors: []collection.Collector{
				collectiontest.NewFunctionCollector("collector-ok", nil),
				collectiontest.NewFunctionCollector("collector-fail", func() ([]ontology.IsResource, error) {
					return nil, errBoom
				}),
			},
		}),
	)
	assert.NoError(t, err)
	defer func() {
		assert.NoError(t, svc.Close())
	}()

	res = svc.RunOnce()

	assert.Equal(t, 2, len(res.CollectorResults))
	assert.Equal(t, "collector-ok", res.CollectorResults[0].CollectorName)
	assert.Equal(t, "collector-fail", res.CollectorResults[1].CollectorName)
	assert.NotEqual(t, "", res.CollectorResults[0].CollectorID)
	assert.NotEqual(t, "", res.CollectorResults[1].CollectorID)
	assert.NotEqual(t, res.CollectorResults[0].CollectorID, res.CollectorResults[1].CollectorID)
	assert.Nil(t, res.CollectorResults[0].Err)
	assert.ErrorIs(t, res.CollectorResults[1].Err, errBoom)

	var (
		foundError bool
	)

	for _, collectorResult := range res.CollectorResults {
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

	svc, err = collection.NewService(
		collection.WithConfig(collection.Config{
			Interval: 20 * time.Millisecond,
			Collectors: []collection.Collector{
				collectiontest.NewFunctionCollector("good-collector", func() ([]ontology.IsResource, error) {
					successCalls.Add(1)
					return nil, nil
				}),
				collectiontest.NewFunctionCollector("fail-collector", func() ([]ontology.IsResource, error) {
					mu.Lock()
					failingCalledOnce = true
					mu.Unlock()
					return nil, errors.New("expected failure")
				}),
			},
		}),
	)
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

			assert.Equal(t, 2, len(result.CollectorResults))
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
		handler              *mockEvidenceStoreHandler
		testHTTPServer       *httptest.Server
		svc                  *collection.Service
		err                  error
		res                  collection.CollectionResult
		storedRequests       []*evidence.StoreEvidenceRequest
	)

	handler = &mockEvidenceStoreHandler{}
	_, testHTTPServer = servertest.NewTestConnectServer(t,
		server.WithHandler(evidenceconnect.NewEvidenceStoreHandler(handler)),
	)
	defer testHTTPServer.Close()

	targetOfEvaluationID = uuid.NewString()

	svc, err = collection.NewService(
		collection.WithConfig(collection.Config{
			Collectors: []collection.Collector{
				collectiontest.NewFunctionCollector("collector-ok", func() ([]ontology.IsResource, error) {
					return []ontology.IsResource{
						&ontology.VirtualMachine{Id: "vm-1"},
					}, nil
				}),
			},
			TargetOfEvaluationID:    targetOfEvaluationID,
			EvidenceStoreAddress:    testHTTPServer.URL,
			EvidenceStoreHTTPClient: testHTTPServer.Client(),
		}),
	)
	assert.NoError(t, err)
	defer func() {
		assert.NoError(t, svc.Close())
	}()

	res = svc.RunOnce()

	assert.Equal(t, 1, len(res.CollectorResults))
	assert.Nil(t, res.CollectorResults[0].Err)

	storedRequests = handler.Requests()
	assert.Equal(t, 1, len(storedRequests))
	assert.Equal(t, targetOfEvaluationID, storedRequests[0].GetEvidence().GetTargetOfEvaluationId())
	assert.Equal(t, "vm-1", storedRequests[0].GetEvidence().GetResource().GetVirtualMachine().GetId())
	assert.NotEqual(t, "", storedRequests[0].GetEvidence().GetToolId())
	assert.NotEqual(t, "", storedRequests[0].GetEvidence().GetId())
}

func TestRunOnce_ReturnsError_WhenEvidenceStoreReturnsErrorStatus(t *testing.T) {
	var (
		targetOfEvaluationID string
		handler              *mockEvidenceStoreHandler
		testHTTPServer       *httptest.Server
		svc                  *collection.Service
		err                  error
		res                  collection.CollectionResult
	)

	handler = &mockEvidenceStoreHandler{
		responseFunc: func(_ *evidence.StoreEvidenceRequest) (*evidence.StoreEvidencesResponse, error) {
			return &evidence.StoreEvidencesResponse{
				Status:        evidence.EvidenceStatus_EVIDENCE_STATUS_ERROR,
				StatusMessage: "invalid evidence",
			}, nil
		},
	}
	_, testHTTPServer = servertest.NewTestConnectServer(t,
		server.WithHandler(evidenceconnect.NewEvidenceStoreHandler(handler)),
	)
	defer testHTTPServer.Close()

	targetOfEvaluationID = uuid.NewString()

	svc, err = collection.NewService(
		collection.WithConfig(collection.Config{
			Collectors: []collection.Collector{
				collectiontest.NewFunctionCollector("collector-ok", func() ([]ontology.IsResource, error) {
					return []ontology.IsResource{
						&ontology.VirtualMachine{Id: "vm-1"},
					}, nil
				}),
			},
			TargetOfEvaluationID:    targetOfEvaluationID,
			EvidenceStoreAddress:    testHTTPServer.URL,
			EvidenceStoreHTTPClient: testHTTPServer.Client(),
		}),
	)
	assert.NoError(t, err)
	defer func() {
		assert.NoError(t, svc.Close())
	}()

	res = svc.RunOnce()

	assert.Equal(t, 1, len(res.CollectorResults))
	assert.ErrorContains(t, res.CollectorResults[0].Err, "evidence-store rejected evidence")
}

func TestNewService_ReturnsError_WhenEvidenceForwardingEnabledWithoutTargetOfEvaluationID(t *testing.T) {
	var (
		svc *collection.Service
		err error
	)

	svc, err = collection.NewService(
		collection.WithConfig(collection.Config{
			Collectors: []collection.Collector{
				collectiontest.NewFunctionCollector("collector", nil),
			},
			EvidenceStoreAddress: "http://localhost:8080",
		}),
	)

	assert.Nil(t, svc)
	assert.ErrorContains(t, err, "target of evaluation id must be set")
}
