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
	"fmt"
	"log/slog"
	"os"
	"reflect"
	"runtime"
	"sync"
	"testing"
	"time"

	"confirmate.io/core/api"
	"confirmate.io/core/api/assessment"
	"confirmate.io/core/api/assessment/assessmentconnect"
	"confirmate.io/core/api/evidence"
	"confirmate.io/core/api/ontology"
	apiOrch "confirmate.io/core/api/orchestrator"
	"confirmate.io/core/persistence"
	"confirmate.io/core/policies"
	"connectrpc.com/connect"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"confirmate.io/core/api/orchestrator/orchestratorconnect"
	"confirmate.io/core/server"
	"confirmate.io/core/server/servertest"
	"confirmate.io/core/service"
	"confirmate.io/core/service/orchestrator"
	"confirmate.io/core/util/assert"
	"confirmate.io/core/util/clitest"
	"confirmate.io/core/util/prototest"
	"confirmate.io/core/util/testdata"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	authPort uint16
)

func TestMain(m *testing.M) {
	clitest.AutoChdir()
	code := m.Run()
	os.Exit(code)
}

// TestNewService tests the NewService function
func TestNewService(t *testing.T) {
	type args struct {
		opts []service.Option[Service]
	}
	tests := []struct {
		name string
		args args
		want assert.Want[*Service]
	}{
		{
			name: "AssessmentServer created with option rego package name",
			args: args{
				opts: []service.Option[Service]{
					WithRegoPackageName("testPkg"),
				},
			},
			want: func(t *testing.T, got *Service, msgAndArgs ...any) bool {
				return assert.Equal(t, "testPkg", got.evalPkg)
			},
		},
		{
			name: "AssessmentServer created with options",
			args: args{
				opts: []service.Option[Service]{
					WithOrchestratorConfig("localhost:9092", nil),
				},
			},
			want: func(t *testing.T, got *Service, msgAndArgs ...any) bool {
				return assert.Equal(t, "localhost:9092", got.orchestratorConfig.targetAddress)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewService(tt.args.opts...)
			assert.NoError(t, err)

			// Create an orchestrator service for testing
			svc, err := orchestrator.NewService(
				orchestrator.WithConfig(orchestrator.Config{
					PersistenceConfig: persistence.Config{
						InMemoryDB: true,
					},
				}),
			)
			assert.NoError(t, err)
			assert.NotNil(t, svc)

			_, testSrv := servertest.NewTestConnectServer(t,
				server.WithHandler(
					orchestratorconnect.NewOrchestratorHandler(svc),
				),
			)
			defer testSrv.Close()

			tt.want(t, got.(*Service))
		})
	}
}

func TestService_AssessEvidence(t *testing.T) {
	type fields struct {
		evidenceResourceMap map[string]*evidence.Evidence
	}
	type args struct {
		req *assessment.AssessEvidenceRequest
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		needsOrch bool // orchestrator is sometimes needed to retrieve metrics
		want      assert.Want[*connect.Response[assessment.AssessEvidenceResponse]]
		wantErr   assert.WantErr
	}{
		{
			name: "Missing evidence",
			want: assert.Nil[*connect.Response[assessment.AssessEvidenceResponse]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				cErr := assert.Is[*connect.Error](t, err)
				return assert.Equal(t, connect.CodeInvalidArgument, cErr.Code())
			},
		},
		{
			name: "Empty evidence",
			args: args{
				req: &assessment.AssessEvidenceRequest{
					Evidence: &evidence.Evidence{},
				},
			},
			want: assert.Nil[*connect.Response[assessment.AssessEvidenceResponse]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				cErr := assert.Is[*connect.Error](t, err)
				return assert.Equal(t, connect.CodeInvalidArgument, cErr.Code())
			},
		},
		{
			name: "Assess evidence without id",
			args: args{
				req: &assessment.AssessEvidenceRequest{
					Evidence: &evidence.Evidence{
						ToolId:    testdata.MockEvidenceToolID1,
						Timestamp: timestamppb.Now(),
						Resource:  prototest.NewProtobufResource(t, &ontology.VirtualMachine{}),
					},
				},
			},
			want: assert.Nil[*connect.Response[assessment.AssessEvidenceResponse]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				cErr := assert.Is[*connect.Error](t, err)
				return assert.Equal(t, connect.CodeInvalidArgument, cErr.Code())
			},
		},
		{
			name: "Assess resource without tool id",
			args: args{
				req: &assessment.AssessEvidenceRequest{
					Evidence: &evidence.Evidence{
						Id:                   testdata.MockEvidenceID1,
						Timestamp:            timestamppb.Now(),
						TargetOfEvaluationId: testdata.MockTargetOfEvaluationID1,
						Resource:             prototest.NewProtobufResource(t, &ontology.VirtualMachine{}),
					},
				},
			},
			want: assert.Nil[*connect.Response[assessment.AssessEvidenceResponse]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				cErr := assert.Is[*connect.Error](t, err)
				return assert.Equal(t, connect.CodeInvalidArgument, cErr.Code())
			},
		},
		{
			name: "Assess resource without timestamp",
			args: args{
				req: &assessment.AssessEvidenceRequest{
					Evidence: &evidence.Evidence{
						Id:                   testdata.MockEvidenceID1,
						ToolId:               testdata.MockEvidenceToolID1,
						TargetOfEvaluationId: testdata.MockTargetOfEvaluationID1,
						Resource: prototest.NewProtobufResource(t, &ontology.VirtualMachine{
							Id:   testdata.MockVirtualMachineID1,
							Name: testdata.MockVirtualMachineName1,
						}),
					},
				},
			},
			want: assert.Nil[*connect.Response[assessment.AssessEvidenceResponse]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				cErr := assert.Is[*connect.Error](t, err)
				return assert.Equal(t, connect.CodeInvalidArgument, cErr.Code())
			},
		},
		{
			name:      "Assess resource happy",
			needsOrch: true, // Needs orchestrator
			fields: fields{
				evidenceResourceMap: make(map[string]*evidence.Evidence),
			},
			args: args{
				req: &assessment.AssessEvidenceRequest{
					Evidence: &evidence.Evidence{
						Id:        testdata.MockEvidenceID1,
						ToolId:    testdata.MockEvidenceToolID1,
						Timestamp: timestamppb.Now(),
						Resource: prototest.NewProtobufResource(t, &ontology.VirtualMachine{
							Id:   testdata.MockVirtualMachineID1,
							Name: testdata.MockVirtualMachineName1,
						}),
						TargetOfEvaluationId: testdata.MockTargetOfEvaluationID1,
					},
				},
			},
			want: func(t *testing.T, got *connect.Response[assessment.AssessEvidenceResponse], args ...any) bool {
				assert.NotNil(t, got.Msg)
				return assert.Equal(t, assessment.AssessmentStatus_ASSESSMENT_STATUS_ASSESSED, got.Msg.Status)
			},
			wantErr: assert.NoError,
		},
		// TODO: integrate when authentication is done
		// {
		// 	name: "Assess resource of wrong cloud service",
		// 	args: args{
		// 		req: &assessment.AssessEvidenceRequest{
		// 			Evidence: &evidence.Evidence{
		// 				Id:        testdata.MockEvidenceID1,
		// 				ToolId:    testdata.MockEvidenceToolID1,
		// 				Timestamp: timestamppb.Now(),
		// 				Resource: prototest.NewProtobufResource(t, &ontology.VirtualMachine{
		// 					Id:   testdata.MockVirtualMachineID1,
		// 					Name: testdata.MockVirtualMachineName1,
		// 				}),
		// 				TargetOfEvaluationId: testdata.MockTargetOfEvaluationID1},
		// 		},
		// 	},
		// 	want: assert.Nil[*connect.Response[assessment.AssessEvidenceResponse]],
		// 	wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
		// 		cErr := assert.Is[*connect.Error](t, err)
		// 		return assert.Equal(t, connect.CodePermissionDenied, cErr.Code())
		// 	},
		// },
		{
			name: "Assess resource without resource id",
			fields: fields{
				evidenceResourceMap: make(map[string]*evidence.Evidence),
			},
			args: args{
				req: &assessment.AssessEvidenceRequest{
					Evidence: &evidence.Evidence{
						Id:                   testdata.MockEvidenceID1,
						ToolId:               testdata.MockEvidenceToolID1,
						Timestamp:            timestamppb.Now(),
						Resource:             prototest.NewProtobufResource(t, &ontology.VirtualMachine{}),
						TargetOfEvaluationId: testdata.MockTargetOfEvaluationID1,
					},
				},
			},
			want: assert.Nil[*connect.Response[assessment.AssessEvidenceResponse]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				cErr := assert.Is[*connect.Error](t, err)
				return assert.Equal(t, connect.CodeInvalidArgument, cErr.Code())
			},
		},
		{
			name:      "Assess resource and wait existing related resources is already there",
			needsOrch: true, // Needs orchestrator
			fields: fields{
				evidenceResourceMap: map[string]*evidence.Evidence{
					"my-other-resource-id": {
						Id: testdata.MockEvidenceID2,
						Resource: prototest.NewProtobufResource(t, &ontology.VirtualMachine{
							Id: testdata.MockVirtualMachineID2,
						}),
					},
				},
			},
			args: args{
				req: &assessment.AssessEvidenceRequest{
					Evidence: &evidence.Evidence{
						Id:                   testdata.MockEvidenceID1,
						ToolId:               testdata.MockEvidenceToolID1,
						Timestamp:            timestamppb.Now(),
						TargetOfEvaluationId: testdata.MockTargetOfEvaluationID1,
						Resource: prototest.NewProtobufResource(t, &ontology.VirtualMachine{
							Id:   testdata.MockVirtualMachineID1,
							Name: testdata.MockVirtualMachineName1,
						}),
						ExperimentalRelatedResourceIds: []string{"my-other-resource-id"},
					},
				},
			},
			want: func(t *testing.T, got *connect.Response[assessment.AssessEvidenceResponse], args ...any) bool {
				assert.NotNil(t, got.Msg)
				return assert.Equal(t, assessment.AssessmentStatus_ASSESSMENT_STATUS_ASSESSED, got.Msg.Status)
			},
			wantErr: assert.NoError,
		},
		{
			name:      "Assess resource and wait, existing related resources are not there",
			needsOrch: true, // Needs orchestrator
			fields: fields{
				evidenceResourceMap: make(map[string]*evidence.Evidence),
			},
			args: args{
				req: &assessment.AssessEvidenceRequest{
					Evidence: &evidence.Evidence{
						Id:                   testdata.MockEvidenceID1,
						ToolId:               testdata.MockEvidenceToolID1,
						Timestamp:            timestamppb.Now(),
						TargetOfEvaluationId: testdata.MockTargetOfEvaluationID1,
						Resource: prototest.NewProtobufResource(t, &ontology.VirtualMachine{
							Id:   testdata.MockVirtualMachineID1,
							Name: testdata.MockVirtualMachineName1,
						}),
						ExperimentalRelatedResourceIds: []string{"my-other-resource-id"},
					},
				},
			},
			want: func(t *testing.T, got *connect.Response[assessment.AssessEvidenceResponse], args ...any) bool {
				assert.NotNil(t, got.Msg)
				return assert.Equal(t, assessment.AssessmentStatus_ASSESSMENT_STATUS_WAITING_FOR_RELATED, got.Msg.Status)
			},
			wantErr: assert.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var s *Service
			var err error

			// Only setup orchestrator if needed
			if tt.needsOrch {
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

				aHandler, err := NewService(
					WithOrchestratorConfig(testSrv.URL, testSrv.Client()),
				)
				s = aHandler.(*Service)
				assert.NoError(t, err)
			} else {
				aHandler, err := NewService()
				s = aHandler.(*Service)
				assert.NoError(t, err)
			}

			// Set evidence resource map if provided
			if tt.fields.evidenceResourceMap != nil {
				s.evidenceResourceMap = tt.fields.evidenceResourceMap
			}

			res, err := s.AssessEvidence(context.Background(), connect.NewRequest(tt.args.req))
			tt.want(t, res)
			tt.wantErr(t, err)
		})
	}
}

func TestService_AssessEvidences(t *testing.T) {
	tests := []struct {
		name         string
		needsOrch    bool
		evidences    []*evidence.Evidence
		wantStatuses []assessment.AssessmentStatus
		wantErr      assert.WantErr
	}{
		{
			name: "Missing toolId",
			evidences: []*evidence.Evidence{{
				Id:                   testdata.MockEvidenceID1,
				Timestamp:            timestamppb.Now(),
				TargetOfEvaluationId: testdata.MockTargetOfEvaluationID1,
				Resource:             prototest.NewProtobufResource(t, &ontology.VirtualMachine{Id: testdata.MockVirtualMachineID1}),
			}},
			wantStatuses: []assessment.AssessmentStatus{assessment.AssessmentStatus_ASSESSMENT_STATUS_FAILED},
			wantErr:      assert.NoError,
		},
		{
			name: "Missing evidenceID",
			evidences: []*evidence.Evidence{{
				Timestamp:            timestamppb.Now(),
				ToolId:               testdata.MockEvidenceToolID1,
				TargetOfEvaluationId: testdata.MockTargetOfEvaluationID1,
				Resource: prototest.NewProtobufResource(t, &ontology.VirtualMachine{
					Id:   testdata.MockVirtualMachineID1,
					Name: testdata.MockVirtualMachineName1,
				}),
			}},
			wantStatuses: []assessment.AssessmentStatus{assessment.AssessmentStatus_ASSESSMENT_STATUS_FAILED},
			wantErr:      assert.NoError,
		},
		{
			name:      "Assess evidences successfully",
			needsOrch: true,
			evidences: []*evidence.Evidence{
				{
					Id:                   testdata.MockEvidenceID1,
					Timestamp:            timestamppb.Now(),
					ToolId:               testdata.MockEvidenceToolID1,
					TargetOfEvaluationId: testdata.MockTargetOfEvaluationZerosID,
					Resource: prototest.NewProtobufResource(t, &ontology.VirtualMachine{
						Id:   testdata.MockVirtualMachineID1,
						Name: testdata.MockVirtualMachineName1,
						BootLogging: &ontology.BootLogging{
							Name:              "loglog",
							LoggingServiceIds: nil,
							Enabled:           true,
						},
					}),
				},
				{
					Id:                   testdata.MockEvidenceID2,
					Timestamp:            timestamppb.Now(),
					ToolId:               testdata.MockEvidenceToolID2,
					TargetOfEvaluationId: testdata.MockTargetOfEvaluationZerosID,
					Resource: prototest.NewProtobufResource(t, &ontology.VirtualMachine{
						Id:   testdata.MockVirtualMachineID2,
						Name: testdata.MockVirtualMachineName2,
						BootLogging: &ontology.BootLogging{
							Name:              "loglog",
							LoggingServiceIds: nil,
							Enabled:           false,
						},
					}),
				},
			},
			wantStatuses: []assessment.AssessmentStatus{
				assessment.AssessmentStatus_ASSESSMENT_STATUS_ASSESSED,
				assessment.AssessmentStatus_ASSESSMENT_STATUS_ASSESSED,
			},
			wantErr: assert.NoError,
		},
	}

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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				err      error
				statuses []assessment.AssessmentStatus
			)

			orchSvc, err := orchestrator.NewService(
				orchestrator.WithConfig(orchestrator.Config{
					PersistenceConfig: persistence.Config{
						InMemoryDB: true,
					},
					LoadDefaultMetrics:              false,
					CreateDefaultTargetOfEvaluation: true,
				}),
			)
			_, testSrv := servertest.NewTestConnectServer(t,
				server.WithHandler(orchestratorconnect.NewOrchestratorHandler(orchSvc)),
			)
			defer testSrv.Close()

			aHandler, err := NewService(
				WithOrchestratorConfig(testSrv.URL, testSrv.Client()),
			)
			require.NoError(t, err)
			s := aHandler.(*Service)

			streamHandle := s.orchestratorStream
			defer func() {
				if streamHandle != nil {
					_ = streamHandle.Close()
				}
			}()

			_, assSrv := servertest.NewTestConnectServer(t,
				server.WithHandler(assessmentconnect.NewAssessmentHandler(s)),
			)
			defer assSrv.Close()

			client := assessmentconnect.NewAssessmentClient(assSrv.Client(), assSrv.URL)
			stream := client.AssessEvidences(context.Background())

			// Create metric in orchestrator
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
						TargetValue:          testdata.MockMetricConfigurationTargetValueTrue,
						IsDefault:            false,
						MetricId:             metric.Id,
						TargetOfEvaluationId: testdata.MockTargetOfEvaluationZerosID,
					},
				}),
			)
			assert.NoError(t, err)

			for _, ev := range tt.evidences {
				sendErr := stream.Send(&assessment.AssessEvidenceRequest{
					Evidence: ev,
				})
				assert.NoError(t, sendErr)
				if sendErr != nil {
					err = sendErr
					break
				}

				res, recvErr := stream.Receive()
				if recvErr != nil {
					err = recvErr
					break
				}
				statuses = append(statuses, res.Status)
			}

			_ = stream.CloseRequest()

			assert.Equal(t, tt.wantStatuses, statuses)
			tt.wantErr(t, err)
		})
	}
}

func TestService_handleEvidence(t *testing.T) {
	type args struct {
		evidence *evidence.Evidence
		resource ontology.IsResource
		metric   *assessment.Metric
		related  map[string]ontology.IsResource
	}
	tests := []struct {
		name    string
		args    args
		want    assert.Want[[]*assessment.AssessmentResult]
		wantErr assert.WantErr
	}{
		{
			name: "nil resource",
			args: args{
				evidence: &evidence.Evidence{
					Id:                   testdata.MockEvidenceID1,
					ToolId:               testdata.MockEvidenceToolID1,
					Timestamp:            timestamppb.Now(),
					TargetOfEvaluationId: testdata.MockTargetOfEvaluationZerosID,
				},
				resource: nil,
			},
			want: assert.Nil[[]*assessment.AssessmentResult],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.ErrorContains(t, err, "invalid embedded resource")
			},
		},
		{
			name: "correct evidence: using metrics which return comparison results",
			args: args{
				evidence: &evidence.Evidence{
					Id:                   testdata.MockEvidenceID1,
					ToolId:               testdata.MockEvidenceToolID1,
					Timestamp:            timestamppb.Now(),
					TargetOfEvaluationId: testdata.MockTargetOfEvaluationZerosID,
					Resource: prototest.NewProtobufResource(t, &ontology.VirtualMachine{
						Id:   testdata.MockVirtualMachineID1,
						Name: testdata.MockVirtualMachineName1,
						BootLogging: &ontology.BootLogging{
							Name:              "loglog",
							LoggingServiceIds: nil,
							Enabled:           true,
						},
					}),
				},
				resource: &ontology.VirtualMachine{
					Id:   testdata.MockVirtualMachineID1,
					Name: testdata.MockVirtualMachineName1,
					BootLogging: &ontology.BootLogging{
						Name:              "loglog",
						LoggingServiceIds: nil,
						Enabled:           true,
					},
				},
				metric: &assessment.Metric{
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
				},
			},
			want: func(t *testing.T, got []*assessment.AssessmentResult, msgAndArgs ...any) bool {
				for _, result := range got {
					err := api.Validate(result)
					assert.NoError(t, err)
				}
				return assert.True(t, got[0].MetricId == "bb41142b-ce8c-4c5c-9b42-360f015fd325" && got[0].Compliant == true)
			},
			wantErr: assert.NoError,
		},
		{
			name: "correct evidence: using metrics which do not return comparison results",
			args: args{
				evidence: &evidence.Evidence{
					Id:                   testdata.MockEvidenceID1,
					ToolId:               testdata.MockEvidenceToolID1,
					Timestamp:            timestamppb.Now(),
					TargetOfEvaluationId: testdata.MockTargetOfEvaluationZerosID,
					Resource: prototest.NewProtobufResource(t, &ontology.VirtualMachine{
						Id:   testdata.MockVirtualMachineID1,
						Name: testdata.MockVirtualMachineName1,
						BootLogging: &ontology.BootLogging{
							LoggingServiceIds: nil,
							Enabled:           true,
						},
					}),
				},
				resource: &ontology.VirtualMachine{
					Id:   testdata.MockVirtualMachineID1,
					Name: testdata.MockVirtualMachineName1,
					BootLogging: &ontology.BootLogging{
						LoggingServiceIds: nil,
						Enabled:           true,
					},
				},
				metric: &assessment.Metric{
					Id:          "4fbcbf09-35c3-4d7b-b9a9-97c7ba36f0de",
					Name:        "ApprovedCommitAuthorEnforced",
					Category:    "DevelopmentLifeCycle",
					Description: testdata.MockMetricDescription1,
					Version:     testdata.MockMetricVersion1,
					Comments:    testdata.MockMetricComments1,
					Implementation: &assessment.MetricImplementation{
						MetricId: "4fbcbf09-35c3-4d7b-b9a9-97c7ba36f0de",
						Lang:     assessment.MetricImplementation_LANGUAGE_REGO,
						Code:     ValidRego(),
					},
				},
			},
			want: assert.Nil[[]*assessment.AssessmentResult],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.Contains(t, err.Error(), fmt.Sprintf("no results"))
			},
		},
		{
			name: "broken Any message",
			args: args{
				evidence: &evidence.Evidence{
					Id:                   testdata.MockEvidenceID1,
					ToolId:               testdata.MockEvidenceToolID1,
					Timestamp:            timestamppb.Now(),
					TargetOfEvaluationId: testdata.MockTargetOfEvaluationZerosID,
					Resource:             &ontology.Resource{},
				},
			},
			want: assert.Nil[[]*assessment.AssessmentResult],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.Contains(t, err.Error(), ontology.ErrNotOntologyResource.Error())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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

			// Create metric and configuration if metric is provided
			if tt.args.metric != nil {
				_, err = orchSvc.CreateMetric(context.Background(), connect.NewRequest(&apiOrch.CreateMetricRequest{
					Metric: tt.args.metric,
				},
				))
				assert.NoError(t, err)

				_, err = orchSvc.UpdateMetricConfiguration(
					context.Background(),
					connect.NewRequest(&apiOrch.UpdateMetricConfigurationRequest{
						Configuration: &assessment.MetricConfiguration{
							Operator:             "==",
							TargetValue:          testdata.MockMetricConfigurationTargetValueTrue,
							IsDefault:            false,
							MetricId:             tt.args.metric.Id,
							TargetOfEvaluationId: testdata.MockTargetOfEvaluationZerosID,
						},
					}),
				)
				assert.NoError(t, err)
			}

			aHandler, err := NewService(
				WithOrchestratorConfig(testSrv.URL, testSrv.Client()),
				WithRegoPackageName(policies.DefaultRegoPackage),
			)
			s := aHandler.(*Service)
			assert.NoError(t, err)

			res, err := s.handleEvidence(context.Background(), tt.args.evidence, tt.args.resource, tt.args.related)
			tt.want(t, res)
			tt.wantErr(t, err)
		})
	}
}

// TestService_AssessEvidence_DetectMisconfiguredEvidenceEvenWhenAlreadyCached tests the following workflow: First an
// evidence with a VM resource is assessed. The resource contains all required fields s.t. the metric cache is filled
// with all applicable metrics. In a second step we assess another evidence. It is also of type "VirtualMachine" but all
// other fields are not set (e.g. MalwareProtection). Thus, metric will be applied and therefore no error occurs in
// AssessEvidence-handleEvidence (assessment.go) which loops over all evaluations
// Todo: Add it to table test above (would probably need some function injection in test cases like we do with storage)
func TestService_AssessEvidence_DetectMisconfiguredEvidenceEvenWhenAlreadyCached(t *testing.T) {

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

	// First assess evidence with a valid VM resource s.t. the cache is created for the combination of resource type and
	// tool id (="VirtualMachine-{testdata.MockEvidenceToolID}")
	e := &evidence.Evidence{
		Id:                   testdata.MockEvidenceID1,
		ToolId:               testdata.MockEvidenceToolID1,
		Timestamp:            timestamppb.Now(),
		TargetOfEvaluationId: testdata.MockTargetOfEvaluationZerosID,
		Resource: prototest.NewProtobufResource(t, &ontology.VirtualMachine{
			Id:   testdata.MockVirtualMachineID1,
			Name: testdata.MockVirtualMachineName1,
		}),
		ExperimentalRelatedResourceIds: []string{"my-other-resource-id"},
	}
	e.Resource = prototest.NewProtobufResource(t, &ontology.VirtualMachine{
		Id:   testdata.MockVirtualMachineID1,
		Name: testdata.MockVirtualMachineName1,
	})
	_, err = s.AssessEvidence(context.Background(), connect.NewRequest(&assessment.AssessEvidenceRequest{Evidence: e}))
	assert.NoError(t, err)

	// Now assess a new evidence which has not a valid format other than the resource type and tool id is set correctly
	// Prepare resource. Make sure both evidences have the same type (for caching key)
	a := prototest.NewProtobufResource(t, &ontology.VirtualMachine{
		Id:   uuid.NewString(),
		Name: "Some other name",
	})

	assert.NoError(t, err)

	_, err = s.AssessEvidence(context.Background(), connect.NewRequest(
		&assessment.AssessEvidenceRequest{
			Evidence: &evidence.Evidence{
				Id:                   uuid.NewString(),
				Timestamp:            timestamppb.Now(),
				TargetOfEvaluationId: testdata.MockTargetOfEvaluationZerosID,
				// Make sure both evidences have the same tool id (for caching key)
				ToolId:   e.ToolId,
				Resource: a,
			},
		}))
	assert.NoError(t, err)
}

func TestService_AssessmentResultHooks(t *testing.T) {
	var (
		hookCallCounter = 0
		wg              sync.WaitGroup
		hookCounts      = 2
	)

	wg.Add(hookCounts)

	firstHookFunction := func(ctx context.Context, assessmentResult *assessment.AssessmentResult, err error) {
		hookCallCounter++
		slog.Info("Hello from inside the firstHookFunction")
		wg.Done()
	}

	secondHookFunction := func(ctx context.Context, assessmentResult *assessment.AssessmentResult, err error) {
		hookCallCounter++
		slog.Info("Hello from inside the secondHookFunction")
		wg.Done()
	}

	type args struct {
		req         *assessment.AssessEvidenceRequest
		resultHooks []assessment.ResultHookFunc
	}
	tests := []struct {
		name    string
		args    args
		want    assert.Want[*connect.Response[assessment.AssessEvidenceResponse]]
		wantErr assert.WantErr
	}{
		{
			name: "Store evidence to the map",
			args: args{
				req: &assessment.AssessEvidenceRequest{
					Evidence: &evidence.Evidence{
						Id:                   testdata.MockEvidenceID1,
						ToolId:               testdata.MockEvidenceToolID1,
						Timestamp:            timestamppb.Now(),
						TargetOfEvaluationId: testdata.MockTargetOfEvaluationZerosID,
						Resource: prototest.NewProtobufResource(t, &ontology.VirtualMachine{
							Id:   testdata.MockVirtualMachineID1,
							Name: testdata.MockVirtualMachineName1,
							BootLogging: &ontology.BootLogging{
								Name:              "BootLogging",
								LoggingServiceIds: []string{"SomeResourceId2"},
								Enabled:           true,
								RetentionPeriod:   durationpb.New(time.Hour * 24 * 36),
							},
							OsLogging: &ontology.OSLogging{
								Name:              "OSLogging",
								LoggingServiceIds: []string{"SomeResourceId2"},
								Enabled:           true,
								RetentionPeriod:   durationpb.New(time.Hour * 24 * 36),
							},
							MalwareProtection: &ontology.MalwareProtection{
								Enabled:              true,
								NumberOfThreatsFound: 5,
								DurationSinceActive:  durationpb.New(time.Hour * 24 * 20),
								ApplicationLogging: &ontology.ApplicationLogging{
									Name:              "AppLogging",
									Enabled:           true,
									LoggingServiceIds: []string{"SomeAnalyticsService?"},
								},
							},
						}),
					}},
				resultHooks: []assessment.ResultHookFunc{firstHookFunction, secondHookFunction},
			},
			want: func(t *testing.T, got *connect.Response[assessment.AssessEvidenceResponse], args ...any) bool {
				assert.NotNil(t, got.Msg)
				return assert.Equal(t, assessment.AssessmentStatus_ASSESSMENT_STATUS_ASSESSED, got.Msg.Status)
			},
			wantErr: assert.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hookCallCounter = 0
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

			for i, hookFunction := range tt.args.resultHooks {
				s.RegisterAssessmentResultHook(hookFunction)

				// Check if hook is registered
				funcName1 := runtime.FuncForPC(reflect.ValueOf(s.resultHooks[i]).Pointer()).Name()
				funcName2 := runtime.FuncForPC(reflect.ValueOf(hookFunction).Pointer()).Name()
				assert.Equal(t, funcName1, funcName2)
			}

			// To test the hooks we have to call a function that calls the hook function
			res, err := s.AssessEvidence(context.Background(), connect.NewRequest(tt.args.req))

			// wait for all hooks
			wg.Wait()

			tt.want(t, res)
			tt.wantErr(t, err)
		})
	}
}

// TestService_Metrics tests the Metrics() method
func TestService_Metrics(t *testing.T) {
	tests := []struct {
		name    string
		want    assert.Want[[]*assessment.Metric]
		wantErr assert.WantErr
	}{
		{
			name: "Retrieve metrics from orchestrator",
			want: func(t *testing.T, got []*assessment.Metric, msgAndArgs ...any) bool {
				return assert.NotNil(t, got)
			},
			wantErr: assert.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				err    error
				metric *assessment.Metric
			)

			metric = &assessment.Metric{
				Id:          testdata.MockMetricID1,
				Name:        testdata.MockMetricName1,
				Category:    testdata.MockMetricCategory1,
				Description: testdata.MockMetricDescription1,
				Version:     testdata.MockMetricVersion1,
				Comments:    testdata.MockMetricComments1,
			}

			// Create an orchestrator service for testing
			orchSvc, err := orchestrator.NewService(
				orchestrator.WithConfig(orchestrator.Config{
					PersistenceConfig: persistence.Config{
						InMemoryDB: true,
					},
					CreateDefaultTargetOfEvaluation: true,
					LoadDefaultCatalogs:             false,
					LoadDefaultMetrics:              false,
				}),
			)
			assert.NoError(t, err)
			assert.NotNil(t, orchSvc)

			srv, testSrv := servertest.NewTestConnectServer(t,
				server.WithHandler(orchestratorconnect.NewOrchestratorHandler(orchSvc)),
			)
			defer testSrv.Close()

			assert.NotNil(t, srv)
			assert.NotNil(t, testSrv)

			// Create metric
			_, err = orchSvc.CreateMetric(context.Background(), connect.NewRequest(&apiOrch.CreateMetricRequest{
				Metric: metric,
			},
			))
			assert.NoError(t, err)

			// Create assessment service
			assessmentHandler, err := NewService(WithOrchestratorConfig(testSrv.URL, testSrv.Client()))
			assert.NoError(t, err)

			// Test
			assessmentSvc := assessmentHandler.(*Service)
			metrics, err := assessmentSvc.Metrics()
			tt.want(t, metrics)
			tt.wantErr(t, err)
		})
	}
}

// TestService_MetricImplementation tests the MetricImplementation() method
func TestService_MetricImplementation(t *testing.T) {
	tests := []struct {
		name    string
		lang    assessment.MetricImplementation_Language
		want    assert.Want[*assessment.MetricImplementation]
		wantErr assert.WantErr
	}{
		{
			name: "Successfully retrieve Rego implementation",
			lang: assessment.MetricImplementation_LANGUAGE_REGO,
			want: func(t *testing.T, got *assessment.MetricImplementation, msgAndArgs ...any) bool {
				return assert.NotNil(t, got)
			},
			wantErr: assert.NoError,
		},
		{
			name: "Unsupported language",
			lang: assessment.MetricImplementation_LANGUAGE_UNSPECIFIED,
			want: assert.Nil[*assessment.MetricImplementation],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.ErrorContains(t, err, "unsupported language")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				err    error
				metric *assessment.Metric
			)

			metricImpl := &assessment.MetricImplementation{
				MetricId: testdata.MockMetricID1,
				Lang:     assessment.MetricImplementation_LANGUAGE_REGO,
				Code:     "test implementation",
			}

			metric = &assessment.Metric{
				Id:             testdata.MockMetricID1,
				Name:           testdata.MockMetricName1,
				Category:       testdata.MockMetricCategory1,
				Description:    testdata.MockMetricDescription1,
				Version:        testdata.MockMetricVersion1,
				Comments:       testdata.MockMetricComments1,
				Implementation: metricImpl,
			}

			// Create an orchestrator service for testing
			svc, err := orchestrator.NewService(
				orchestrator.WithConfig(orchestrator.Config{
					PersistenceConfig: persistence.Config{
						InMemoryDB: true,
					},
					CreateDefaultTargetOfEvaluation: false,
					LoadDefaultCatalogs:             false,
					LoadDefaultMetrics:              false,
				}),
			)
			assert.NoError(t, err)
			assert.NotNil(t, svc)

			srv, testSrv := servertest.NewTestConnectServer(t,
				server.WithHandler(orchestratorconnect.NewOrchestratorHandler(svc)),
			)
			defer testSrv.Close()

			assert.NotNil(t, srv)
			assert.NotNil(t, testSrv)

			// Create metric and implementation
			_, err = svc.CreateMetric(context.Background(), connect.NewRequest(&apiOrch.CreateMetricRequest{
				Metric: metric,
			},
			))
			assert.NoError(t, err)

			// Create assessment service
			assessmentHandler, err := NewService(WithOrchestratorConfig(testSrv.URL, testSrv.Client()))
			assert.NoError(t, err)

			// Test
			assessmentSvc := assessmentHandler.(*Service)
			impl, err := assessmentSvc.MetricImplementation(tt.lang, metric)
			tt.want(t, impl)
			tt.wantErr(t, err)
		})
	}
}

// TestService_MetricConfiguration tests the MetricConfiguration() method including caching
func TestService_MetricConfiguration(t *testing.T) {
	type fields struct {
		db persistence.DB
	}
	tests := []struct {
		name           string
		toeID          string
		metric         *assessment.Metric
		preCacheConfig *cachedConfiguration
		want           assert.Want[*assessment.MetricConfiguration]
		wantCached     bool
		wantErr        assert.WantErr
	}{
		{
			name:  "Successfully retrieve and cache configuration",
			toeID: testdata.MockTargetOfEvaluationID1,
			metric: &assessment.Metric{
				Id:          testdata.MockMetricID1,
				Name:        testdata.MockMetricName1,
				Description: testdata.MockMetricDescription1,
				Category:    testdata.MockMetricCategory1,
				Version:     testdata.MockMetricVersion1,
				Comments:    testdata.MockMetricComments1,
			},
			want: func(t *testing.T, got *assessment.MetricConfiguration, msgAndArgs ...any) bool {
				return assert.NotNil(t, got)
			},
			wantCached: true,
			wantErr:    assert.NoError,
		},
		{
			name: "Successfully retrieve and cache configuration",
			metric: &assessment.Metric{
				Id:          testdata.MockMetricID1,
				Name:        testdata.MockMetricName1,
				Description: testdata.MockMetricDescription1,
				Category:    testdata.MockMetricCategory1,
				Version:     testdata.MockMetricVersion1,
				Comments:    testdata.MockMetricComments1,
			},
			preCacheConfig: &cachedConfiguration{
				cachedAt: time.Now(),
				MetricConfiguration: &assessment.MetricConfiguration{
					Operator: "==",
				},
			},
			want: func(t *testing.T, got *assessment.MetricConfiguration, msgAndArgs ...any) bool {
				return assert.Equal(t, "==", got.Operator)
			},
			wantCached: true,
			wantErr:    assert.NoError,
		},
		{
			name:  "Successfully retrieve and cache configuration",
			toeID: testdata.MockTargetOfEvaluationID1,
			metric: &assessment.Metric{
				Id:          testdata.MockMetricID1,
				Name:        testdata.MockMetricName1,
				Description: testdata.MockMetricDescription1,
				Category:    testdata.MockMetricCategory1,
				Version:     testdata.MockMetricVersion1,
				Comments:    testdata.MockMetricComments1,
			},
			preCacheConfig: &cachedConfiguration{
				cachedAt: time.Now().Add(-2 * EvictionTime), // Expired
				MetricConfiguration: &assessment.MetricConfiguration{
					Operator: "old-value",
				},
			},
			want: func(t *testing.T, got *assessment.MetricConfiguration, msgAndArgs ...any) bool {
				return assert.NotNil(t, got)
			},
			wantCached: true,
			wantErr:    assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				assSvc *Service
			)

			orchSvc, err := orchestrator.NewService(
				orchestrator.WithConfig(orchestrator.Config{
					PersistenceConfig: persistence.Config{
						InMemoryDB: true,
					},
					LoadDefaultCatalogs:             false,
					LoadDefaultMetrics:              false,
					CreateDefaultTargetOfEvaluation: false,
				}),
			)
			assert.NoError(t, err)
			assert.NotNil(t, orchSvc)

			srv, testSrv := servertest.NewTestConnectServer(t,
				server.WithHandler(orchestratorconnect.NewOrchestratorHandler(orchSvc)),
			)
			defer testSrv.Close()

			assert.NotNil(t, srv)
			assert.NotNil(t, testSrv)

			// Create metric and implementation
			_, err = orchSvc.CreateMetric(context.Background(), connect.NewRequest(&apiOrch.CreateMetricRequest{
				Metric: tt.metric,
			},
			))
			assert.NoError(t, err)

			res, err := orchSvc.CreateTargetOfEvaluation(
				context.Background(),
				connect.NewRequest(&apiOrch.CreateTargetOfEvaluationRequest{
					TargetOfEvaluation: &apiOrch.TargetOfEvaluation{
						Name:        "Test TOE",
						Description: "test description",
					},
				}),
			)
			assert.NoError(t, err)

			_, err = orchSvc.UpdateMetricConfiguration(
				context.Background(),
				connect.NewRequest(&apiOrch.UpdateMetricConfigurationRequest{
					Configuration: &assessment.MetricConfiguration{
						Operator:             "==",
						TargetValue:          testdata.MockMetricConfigurationTargetValueString,
						IsDefault:            false,
						MetricId:             testdata.MockMetricID1,
						TargetOfEvaluationId: res.Msg.Id,
					},
				}),
			)
			assert.NoError(t, err)

			// Create assessment service
			handler, err := NewService(
				WithOrchestratorConfig(testSrv.URL, testSrv.Client()),
			)
			assert.NoError(t, err)
			assSvc = handler.(*Service)
			assSvc.cachedConfigurations = make(map[string]cachedConfiguration)

			// Pre-populate cache if needed
			if tt.preCacheConfig != nil {
				key := fmt.Sprintf("%s-%s", res.Msg.Id, tt.metric.Id)
				assSvc.cachedConfigurations[key] = *tt.preCacheConfig
			}

			// Execute test
			config, err := assSvc.MetricConfiguration(res.Msg.Id, tt.metric)

			tt.want(t, config)
			tt.wantErr(t, err)

			// Verify caching
			if tt.wantCached && err == nil {
				key := fmt.Sprintf("%s-%s", res.Msg.Id, tt.metric.Id)
				_, exists := assSvc.cachedConfigurations[key]
				assert.True(t, exists, "Configuration should be cached")
			}
		})
	}
}

// TestService_initOrchestratorStream tests the orchestrator stream initialization
func TestService_initOrchestratorStream(t *testing.T) {
	tests := []struct {
		name    string
		wantErr assert.WantErr
	}{
		{
			name:    "Successfully initialize stream",
			wantErr: assert.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup orchestrator
			orchSvc, err := orchestrator.NewService(
				orchestrator.WithConfig(orchestrator.Config{
					PersistenceConfig: persistence.Config{
						InMemoryDB: true,
					},
				}),
			)
			assert.NoError(t, err)
			assert.NotNil(t, orchSvc)

			_, testSrv := servertest.NewTestConnectServer(t,
				server.WithHandler(
					orchestratorconnect.NewOrchestratorHandler(orchSvc),
				),
			)
			defer testSrv.Close()

			// Create service
			assSvc := &Service{
				orchestratorConfig: orchestratorConfig{
					targetAddress: testSrv.URL,
					client:        testSrv.Client(),
				},
			}

			assSvc.orchestratorClient = orchestratorconnect.NewOrchestratorClient(
				assSvc.orchestratorConfig.client,
				assSvc.orchestratorConfig.targetAddress,
			)

			// Execute test
			err = assSvc.initOrchestratorStream()
			tt.wantErr(t, err)

			if err == nil {
				assert.NotNil(t, assSvc.orchestratorStream)
			}
		})
	}
}

// TestService_informHooks tests the informHooks method
func TestService_informHooks(t *testing.T) {
	tests := []struct {
		name     string
		setupSvc func(t *testing.T) (*Service, *int)
		result   *assessment.AssessmentResult
		err      error
		verify   func(t *testing.T, counter *int)
	}{
		{
			name: "Inform single hook with result",
			setupSvc: func(t *testing.T) (*Service, *int) {
				counter := 0
				svc := &Service{}
				svc.RegisterAssessmentResultHook(func(ctx context.Context, result *assessment.AssessmentResult, err error) {
					counter++
				})
				return svc, &counter
			},
			result: &assessment.AssessmentResult{
				Id:        "test-result",
				Compliant: true,
			},
			err: nil,
			verify: func(t *testing.T, counter *int) {
				time.Sleep(100 * time.Millisecond)
				assert.Equal(t, 1, *counter)
			},
		},
		{
			name: "Inform multiple hooks",
			setupSvc: func(t *testing.T) (*Service, *int) {
				counter := 0
				svc := &Service{}
				svc.RegisterAssessmentResultHook(func(ctx context.Context, result *assessment.AssessmentResult, err error) {
					counter++
				})
				svc.RegisterAssessmentResultHook(func(ctx context.Context, result *assessment.AssessmentResult, err error) {
					counter++
				})
				return svc, &counter
			},
			result: &assessment.AssessmentResult{
				Id: "test-result",
			},
			verify: func(t *testing.T, counter *int) {
				time.Sleep(100 * time.Millisecond)
				assert.Equal(t, 2, *counter)
			},
		},
		{
			name: "Inform hooks with error",
			setupSvc: func(t *testing.T) (*Service, *int) {
				counter := 0
				svc := &Service{}
				svc.RegisterAssessmentResultHook(func(ctx context.Context, result *assessment.AssessmentResult, err error) {
					if err != nil {
						counter++
					}
				})
				return svc, &counter
			},
			result: nil,
			err:    fmt.Errorf("test error"),
			verify: func(t *testing.T, counter *int) {
				time.Sleep(100 * time.Millisecond)
				assert.Equal(t, 1, *counter)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, counter := tt.setupSvc(t)
			svc.informHooks(context.Background(), tt.result, tt.err)
			tt.verify(t, counter)
		})
	}
}

// TestService_RegisterAssessmentResultHook tests hook registration
func TestService_RegisterAssessmentResultHook(t *testing.T) {
	svc := &Service{}

	hook1 := func(ctx context.Context, result *assessment.AssessmentResult, err error) {}
	hook2 := func(ctx context.Context, result *assessment.AssessmentResult, err error) {}

	svc.RegisterAssessmentResultHook(hook1)
	assert.Equal(t, 1, len(svc.resultHooks))

	svc.RegisterAssessmentResultHook(hook2)
	assert.Equal(t, 2, len(svc.resultHooks))
}

// TestService_createOrchestratorStreamFactory tests the stream factory creation
func TestService_createOrchestratorStreamFactory(t *testing.T) {
	t.Run("Create and use stream factory", func(t *testing.T) {
		// Setup orchestrator
		orchSvc, err := orchestrator.NewService(
			orchestrator.WithConfig(orchestrator.Config{
				PersistenceConfig: persistence.Config{
					InMemoryDB: true,
				},
			}),
		)
		assert.NoError(t, err)
		assert.NotNil(t, orchSvc)

		_, testSrv := servertest.NewTestConnectServer(t,
			server.WithHandler(
				orchestratorconnect.NewOrchestratorHandler(orchSvc),
			),
		)
		defer testSrv.Close()

		// Create service
		handler, err := NewService(
			WithOrchestratorConfig(testSrv.URL, testSrv.Client()),
		)
		assert.NoError(t, err)

		svc := handler.(*Service)

		// Create factory
		factory := svc.createOrchestratorStreamFactory()
		assert.NotNil(t, factory)

		// Test that factory can create streams
		stream := factory(context.Background())
		assert.NotNil(t, stream)
	})
}

func ValidRego() string {
	return `package cch.metrics.boot_logging_enabled

	import data.cch.compare
	import rego.v1
	import input.bootLogging as logging

	default applicable = false

	default compliant = false

	applicable if {
		logging
	}

	compliant if {
		compare(data.operator, data.target_value, logging.enabled)
	}`
}
