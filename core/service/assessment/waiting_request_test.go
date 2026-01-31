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

package assessment

import (
	"context"
	"testing"
	"time"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/timestamppb"

	"confirmate.io/core/api/assessment"
	"confirmate.io/core/api/evidence"
	"confirmate.io/core/api/ontology"
	apiOrch "confirmate.io/core/api/orchestrator"
	"confirmate.io/core/api/orchestrator/orchestratorconnect"
	"confirmate.io/core/persistence"
	"confirmate.io/core/policies"
	"confirmate.io/core/server"
	"confirmate.io/core/server/servertest"
	"confirmate.io/core/service/orchestrator"
	"confirmate.io/core/util/assert"
	"confirmate.io/core/util/prototest"
	"confirmate.io/core/util/testdata"
)

// Test evidence IDs
const (
	testEvidenceID1 = "11111111-1111-1111-1111-111111111111"
	testEvidenceID2 = "22222222-2222-2222-2222-222222222222"
	testEvidenceID3 = "33333333-3333-3333-3333-333333333333"
)

// Test resource IDs
const (
	testResourceID1 = "my-resource"
	testResourceID2 = "my-other-resource"
	testResourceID3 = "my-third-resource"
)

func TestService_AssessEvidenceWaitFor(t *testing.T) {
	type testCase struct {
		name           string
		evidence       *evidence.Evidence
		wantStatus     assessment.AssessmentStatus
		waitForService bool
		wantErr        assert.WantErr
	}

	tests := []testCase{
		{
			name: "first evidence waits for related resource",
			evidence: &evidence.Evidence{
				Id: testEvidenceID1,
				Resource: prototest.NewProtobufResource(t, &ontology.VirtualMachine{
					Id:              testResourceID1,
					Name:            "my resource",
					BlockStorageIds: []string{testResourceID3},
				}),
				TargetOfEvaluationId:           testdata.MockTargetOfEvaluationID1,
				ToolId:                         "my-tool",
				Timestamp:                      timestamppb.Now(),
				ExperimentalRelatedResourceIds: []string{testResourceID3},
			},
			wantStatus: assessment.AssessmentStatus_ASSESSMENT_STATUS_WAITING_FOR_RELATED,
			wantErr:    assert.NoError,
		},
		{
			name: "second evidence also waits for same resource",
			evidence: &evidence.Evidence{
				Id: testEvidenceID2,
				Resource: prototest.NewProtobufResource(t, &ontology.VirtualMachine{
					Id:              testResourceID2,
					Name:            "my other resource",
					BlockStorageIds: []string{testResourceID3},
				}),
				TargetOfEvaluationId:           testdata.MockTargetOfEvaluationID1,
				ToolId:                         "my-tool",
				Timestamp:                      timestamppb.Now(),
				ExperimentalRelatedResourceIds: []string{testResourceID3},
			},
			wantStatus: assessment.AssessmentStatus_ASSESSMENT_STATUS_WAITING_FOR_RELATED,
			wantErr:    assert.NoError,
		},
		{
			name: "third evidence completes assessment with mutual dependency",
			evidence: &evidence.Evidence{
				Id: testEvidenceID3,
				Resource: prototest.NewProtobufResource(t, &ontology.BlockStorage{
					Id:   testResourceID3,
					Name: "my third resource",
				}),
				TargetOfEvaluationId:           testdata.MockTargetOfEvaluationID1,
				ToolId:                         "my-tool",
				Timestamp:                      timestamppb.Now(),
				ExperimentalRelatedResourceIds: []string{testResourceID1},
			},
			wantStatus:     assessment.AssessmentStatus_ASSESSMENT_STATUS_ASSESSED,
			wantErr:        assert.NoError,
			waitForService: true,
		},
	}

	// Setup orchestrator for actual assessment
	orchSvc, err := orchestrator.NewService(
		orchestrator.WithConfig(orchestrator.Config{
			PersistenceConfig: persistence.Config{
				InMemoryDB: true,
			},
		}),
	)
	assert.NoError(t, err)

	_, testSrv := servertest.NewTestConnectServer(t,
		server.WithHandler(orchestratorconnect.NewOrchestratorHandler(orchSvc)),
	)
	defer testSrv.Close()

	// Create assessment service with orchestrator
	svc, err := NewService(
		WithOrchestratorConfig(testSrv.URL, testSrv.Client()),
	)
	assert.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			req := connect.NewRequest(&assessment.AssessEvidenceRequest{
				Evidence: tt.evidence,
			})

			resp, err := svc.(*Service).AssessEvidence(ctx, req)

			tt.wantErr(t, err)

			assert.Equal(t, tt.wantStatus, resp.Msg.Status)

			// Wait for background processing if needed
			if tt.waitForService {
				// Use a timeout to avoid hanging tests
				done := make(chan struct{})
				go func() {
					if s, ok := svc.(*Service); ok {
						s.wg.Wait()
					}
					close(done)
				}()

				select {
				case <-done:
					// Success
				case <-time.After(5 * time.Second):
					t.Fatal("timeout waiting for service to complete")
				}
			}
		})
	}

	// Final verification: no requests should be left over
	if s, ok := svc.(*Service); ok {
		assert.Empty(t, s.requests, "expected all requests to be processed")
	}
}

// TestService_AssessEvidenceWaitFor_Integration is an alternative integration-style test
// that tests the entire flow as a single scenario
func TestService_AssessEvidenceWaitFor_Integration(t *testing.T) {

	orchSvc, err := orchestrator.NewService(
		orchestrator.WithConfig(orchestrator.Config{
			PersistenceConfig: persistence.Config{
				InMemoryDB: true,
			},
			LoadDefaultMetrics:              false,
			CreateDefaultTargetOfEvaluation: true,
		}),
	)
	assert.NoError(t, err)

	_, testSrv := servertest.NewTestConnectServer(t,
		server.WithHandler(orchestratorconnect.NewOrchestratorHandler(orchSvc)),
	)
	aHandler, err := NewService(
		WithOrchestratorConfig(testSrv.URL, testSrv.Client()),
		WithRegoPackageName(policies.DefaultRegoPackage),
	)
	assert.NoError(t, err)
	s := aHandler.(*Service)

	// Create metric
	metric := &assessment.Metric{
		Id:          "bb41142b-ce8c-4c5c-9b42-360f015fd325",
		Name:        "BootLoggingEnabled",
		Category:    "LoggingMonitoring",
		Description: testdata.MockMetricDescription1,
		Version:     testdata.MockMetricVersion1,
		Comments:    testdata.MockMetricComments1,
		Implementation: &assessment.MetricImplementation{
			MetricId: "bb41142b-ce8c-4c5c-9b42-360f015fd325",
			Lang:     assessment.MetricImplementation_LANGUAGE_REGO,
			Code:     ValidRego(),
		},
	}

	_, err = orchSvc.CreateMetric(context.Background(), connect.NewRequest(&apiOrch.CreateMetricRequest{
		Metric: metric,
	},
	))
	assert.NoError(t, err)

	_, err = orchSvc.UpdateMetricConfiguration(
		context.Background(),
		connect.NewRequest(&apiOrch.UpdateMetricConfigurationRequest{
			Configuration: &assessment.MetricConfiguration{
				Operator:             "==",
				TargetValue:          testdata.MockMetricConfigurationTargetValueString,
				IsDefault:            false,
				MetricId:             metric.Id,
				TargetOfEvaluationId: testdata.MockTargetOfEvaluationZerosID,
			},
		}),
	)
	assert.NoError(t, err)

	// Step 1: Add first evidence waiting for related resource
	e1 := &evidence.Evidence{
		Id: testEvidenceID1,
		Resource: prototest.NewProtobufResource(t, &ontology.VirtualMachine{
			Id:              testResourceID1,
			Name:            "my resource",
			BlockStorageIds: []string{testResourceID3},
		}),
		TargetOfEvaluationId:           testdata.MockTargetOfEvaluationZerosID,
		ToolId:                         "my-tool",
		Timestamp:                      timestamppb.Now(),
		ExperimentalRelatedResourceIds: []string{testResourceID3},
	}

	resp1, err := s.AssessEvidence(context.Background(), connect.NewRequest(&assessment.AssessEvidenceRequest{Evidence: e1}))
	assert.NoError(t, err)
	assert.Equal(t, assessment.AssessmentStatus_ASSESSMENT_STATUS_WAITING_FOR_RELATED, resp1.Msg.Status)

	// Step 2: Add second evidence also waiting for same resource
	e2 := &evidence.Evidence{
		Id: testEvidenceID2,
		Resource: prototest.NewProtobufResource(t, &ontology.VirtualMachine{
			Id:              testResourceID2,
			Name:            "my other resource",
			BlockStorageIds: []string{testResourceID3},
		}),
		TargetOfEvaluationId:           testdata.MockTargetOfEvaluationZerosID,
		ToolId:                         "my-tool",
		Timestamp:                      timestamppb.Now(),
		ExperimentalRelatedResourceIds: []string{testResourceID3},
	}

	resp2, err := s.AssessEvidence(context.Background(), connect.NewRequest(&assessment.AssessEvidenceRequest{Evidence: e2}))
	assert.NoError(t, err)
	assert.Equal(t, assessment.AssessmentStatus_ASSESSMENT_STATUS_WAITING_FOR_RELATED, resp2.Msg.Status)

	// Step 3: Add evidence for the resource both are waiting for (with mutual dependency)
	e3 := &evidence.Evidence{
		Id: testEvidenceID3,
		Resource: prototest.NewProtobufResource(t, &ontology.BlockStorage{
			Id:   testResourceID3,
			Name: "my third resource",
		}),
		TargetOfEvaluationId:           testdata.MockTargetOfEvaluationZerosID,
		ToolId:                         "my-tool",
		Timestamp:                      timestamppb.Now(),
		ExperimentalRelatedResourceIds: []string{testResourceID1},
	}

	resp3, err := s.AssessEvidence(context.Background(), connect.NewRequest(&assessment.AssessEvidenceRequest{Evidence: e3}))
	assert.NoError(t, err)
	assert.Equal(t, assessment.AssessmentStatus_ASSESSMENT_STATUS_ASSESSED, resp3.Msg.Status)

	// Wait for background processing with timeout
	waitForServiceWithTimeout(t, s, 5*time.Second)

	// Verify: no requests should be left over
	assert.Empty(t, s.requests, "expected all requests to be processed")
}

// waitForServiceWithTimeout waits for the service to complete background work.
// It accepts *Service so the helper does not depend on the Connect handler type.
func waitForServiceWithTimeout(t *testing.T, s *Service, timeout time.Duration) {
	t.Helper()

	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Success
	case <-time.After(timeout):
		t.Fatalf("timeout after %v waiting for service to complete", timeout)
	}
}
