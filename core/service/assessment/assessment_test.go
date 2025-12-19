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

package assessment

import (
	"context"
	"testing"

	"confirmate.io/core/api"
	"confirmate.io/core/api/assessment"
	"confirmate.io/core/api/evidence"
	"confirmate.io/core/api/ontology"
	"confirmate.io/core/policies"

	"confirmate.io/core/api/orchestrator/orchestratorconnect"
	"confirmate.io/core/server"
	"confirmate.io/core/server/servertest"
	"confirmate.io/core/service"
	orchestratorsvc "confirmate.io/core/service/orchestrator"
	"confirmate.io/core/util/assert"
	"confirmate.io/core/util/prototest"
	"confirmate.io/core/util/testdata"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	authPort uint16
)

func TestMain(m *testing.M) {
}

// TestNewService is a simply test for NewService
func TestNewService(t *testing.T) {
	type args struct {
		opts []service.Option[*Service]
	}
	tests := []struct {
		name string
		args args
		want assert.Want[*Service]
	}{
		{
			name: "AssessmentServer created with option rego package name",
			args: args{
				opts: []service.Option[*Service]{
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
				opts: []service.Option[*Service]{
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
			got := NewService(tt.args.opts...)

			svc, err := orchestratorsvc.NewService()
			assert.NoError(t, err)

			_, testSrv := servertest.NewTestConnectServer(t,
				server.WithHandler(
					orchestratorconnect.NewOrchestratorHandler(svc),
				),
			)
			defer testSrv.Close()

			tt.want(t, got)
		})
	}
}

// TestAssessEvidence tests AssessEvidence
func TestService_AssessEvidence(t *testing.T) {
	type fields struct {
		// authz               service.AuthorizationStrategy
		// orchestrator        *api.orchestratorconnect[orchestrator.OrchestratorClient]
		evidenceResourceMap map[string]*evidence.Evidence
	}
	type args struct {
		in0      context.Context
		evidence *evidence.Evidence
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantResp *assessment.AssessEvidenceResponse
		wantErr  assert.ErrorAssertionFunc
	}{
		{
			name: "Missing evidence",
			args: args{
				in0: context.TODO(),
			},
			wantResp: nil,
			wantErr: func(tt assert.TestingT, err error, i ...interface{}) bool {
				return assert.ErrorContains(t, err, "evidence: value is required")
			},
		},
		{
			name: "Empty evidence",
			args: args{
				in0:      context.TODO(),
				evidence: &evidence.Evidence{},
			},
			wantResp: nil,
			wantErr: func(tt assert.TestingT, err error, i ...interface{}) bool {
				return assert.ErrorContains(t, err, "evidence.id: value is empty, which is not a valid UUID")
			},
		},
		{
			name:   "Assess evidence without id",
			fields: fields{
				// evidenceStore: api.NewRPCConnection(testdata.MockGRPCTarget, evidence.NewEvidenceStoreClient, grpc.WithContextDialer(bufConnDialer)),
				// orchestrator: api.NewRPCConnection(testdata.MockGRPCTarget, orchestrator.NewOrchestratorClient, grpc.WithContextDialer(bufConnDialer)),
			},
			args: args{
				in0: context.TODO(),
				evidence: &evidence.Evidence{
					ToolId:    testdata.MockEvidenceToolID1,
					Timestamp: timestamppb.Now(),
					Resource:  prototest.NewProtobufResource(t, &ontology.VirtualMachine{}),
				},
			},
			wantResp: nil,
			wantErr: func(tt assert.TestingT, err error, i ...interface{}) bool {
				return assert.ErrorContains(t, err, "evidence.id: value is empty, which is not a valid UUID")
			},
		},
		{
			name:   "Assess resource without tool id",
			fields: fields{
				// evidenceStore: api.NewRPCConnection(testdata.MockGRPCTarget, evidence.NewEvidenceStoreClient, grpc.WithContextDialer(bufConnDialer)),
				// orchestrator:  api.NewRPCConnection(testdata.MockGRPCTarget, orchestrator.NewOrchestratorClient, grpc.WithContextDialer(bufConnDialer)),
			},
			args: args{
				in0: context.TODO(),
				evidence: &evidence.Evidence{
					Id:                   testdata.MockEvidenceID1,
					Timestamp:            timestamppb.Now(),
					TargetOfEvaluationId: testdata.MockTargetOfEvaluationID1,
					Resource:             prototest.NewProtobufResource(t, &ontology.VirtualMachine{}),
				},
			},
			wantResp: nil,
			wantErr: func(tt assert.TestingT, err error, i ...interface{}) bool {
				return assert.ErrorContains(t, err, "evidence.tool_id: value length must be at least 1 characters")
			},
		},
		{
			name:   "Assess resource without timestamp",
			fields: fields{
				// evidenceStore: api.NewRPCConnection(testdata.MockGRPCTarget, evidence.NewEvidenceStoreClient, grpc.WithContextDialer(bufConnDialer)),
				// orchestrator:  api.NewRPCConnection(testdata.MockGRPCTarget, orchestrator.NewOrchestratorClient, grpc.WithContextDialer(bufConnDialer)),
			},
			args: args{
				in0: context.TODO(),
				evidence: &evidence.Evidence{
					Id:                   testdata.MockEvidenceID1,
					ToolId:               testdata.MockEvidenceToolID1,
					TargetOfEvaluationId: testdata.MockTargetOfEvaluationID1,
					Resource: prototest.NewProtobufResource(t, &ontology.VirtualMachine{
						Id:   testdata.MockVirtualMachineID1,
						Name: testdata.MockVirtualMachineName1,
					}),
				},
			},
			wantResp: nil,
			wantErr: func(tt assert.TestingT, err error, i ...interface{}) bool {
				return assert.ErrorContains(t, err, "evidence.timestamp: value is required")
			},
		},
		{
			name: "Assess resource happy",
			fields: fields{
				// evidenceStore:       api.NewRPCConnection(testdata.MockGRPCTarget, evidence.NewEvidenceStoreClient, grpc.WithContextDialer(bufConnDialer)),
				// orchestrator:        api.NewRPCConnection(testdata.MockGRPCTarget, orchestrator.NewOrchestratorClient, grpc.WithContextDialer(bufConnDialer)),
				// authz:               servicetest.NewAuthorizationStrategy(true),
				evidenceResourceMap: make(map[string]*evidence.Evidence),
			},
			args: args{
				in0: context.TODO(),
				evidence: &evidence.Evidence{
					Id:        testdata.MockEvidenceID1,
					ToolId:    testdata.MockEvidenceToolID1,
					Timestamp: timestamppb.Now(),
					Resource: prototest.NewProtobufResource(t, &ontology.VirtualMachine{
						Id:   testdata.MockVirtualMachineID1,
						Name: testdata.MockVirtualMachineName1,
					}),
					TargetOfEvaluationId: testdata.MockTargetOfEvaluationID1},
			},
			wantResp: &assessment.AssessEvidenceResponse{
				Status: assessment.AssessmentStatus_ASSESSMENT_STATUS_ASSESSED,
			},
			wantErr: func(tt assert.TestingT, err error, i ...interface{}) bool {
				return assert.NoError(t, err)
			},
		},
		{
			name:   "Assess resource of wrong could service",
			fields: fields{
				// authz: servicetest.NewAuthorizationStrategy(false, testdata.MockTargetOfEvaluationID2),
			},
			args: args{
				in0: context.TODO(),
				evidence: &evidence.Evidence{
					Id:        testdata.MockEvidenceID1,
					ToolId:    testdata.MockEvidenceToolID1,
					Timestamp: timestamppb.Now(),
					Resource: prototest.NewProtobufResource(t, &ontology.VirtualMachine{
						Id:   testdata.MockVirtualMachineID1,
						Name: testdata.MockVirtualMachineName1,
					}),
					TargetOfEvaluationId: testdata.MockTargetOfEvaluationID1},
			},
			wantResp: nil,
			wantErr: func(tt assert.TestingT, err error, i ...interface{}) bool {
				return assert.ErrorContains(t, err, service.ErrPermissionDenied.Error())
			},
		},
		{
			name: "Assess resource without resource id",
			fields: fields{
				// evidenceStore:       api.NewRPCConnection(testdata.MockGRPCTarget, evidence.NewEvidenceStoreClient, grpc.WithContextDialer(bufConnDialer)),
				// orchestrator:        api.NewRPCConnection(testdata.MockGRPCTarget, orchestrator.NewOrchestratorClient, grpc.WithContextDialer(bufConnDialer)),
				// authz:               servicetest.NewAuthorizationStrategy(true),
				evidenceResourceMap: make(map[string]*evidence.Evidence),
			},
			args: args{
				in0: context.TODO(),
				evidence: &evidence.Evidence{
					Id:                   testdata.MockEvidenceID1,
					ToolId:               testdata.MockEvidenceToolID1,
					Timestamp:            timestamppb.Now(),
					Resource:             prototest.NewProtobufResource(t, &ontology.VirtualMachine{}),
					TargetOfEvaluationId: testdata.MockTargetOfEvaluationID1,
				},
			},
			wantResp: nil,
			wantErr: func(tt assert.TestingT, err error, i ...interface{}) bool {
				return assert.ErrorContains(t, err, "id: value is required")
			},
		},
		// {
		// 	name: "No RPC connections",
		// 	fields: fields{
		// 		// evidenceStore:       api.NewRPCConnection(testdata.MockGRPCTarget, evidence.NewEvidenceStoreClient, grpc.WithContextDialer(connectionRefusedDialer)),
		// 		// orchestrator:        api.NewRPCConnection(testdata.MockGRPCTarget, orchestrator.NewOrchestratorClient, grpc.WithContextDialer(connectionRefusedDialer)),
		// 		// authz:               servicetest.NewAuthorizationStrategy(true),
		// 		evidenceResourceMap: make(map[string]*evidence.Evidence),
		// 	},
		// 	args: args{
		// 		in0: context.TODO(),
		// 		evidence: &evidence.Evidence{
		// 			Id:                   testdata.MockEvidenceID1,
		// 			ToolId:               testdata.MockEvidenceToolID1,
		// 			Timestamp:            timestamppb.Now(),
		// 			TargetOfEvaluationId: testdata.MockTargetOfEvaluationID1,
		// 			Resource: prototest.NewProtobufResource(t, &ontology.VirtualMachine{
		// 				Id:   testdata.MockVirtualMachineID1,
		// 				Name: testdata.MockVirtualMachineName1,
		// 			}),
		// 		},
		// 	},
		// 	wantResp: nil,
		// 	wantErr: func(tt assert.TestingT, err error, i ...interface{}) bool {
		// 		return assert.ErrorContains(t, err, "connection refused")
		// 	},
		// },
		{
			name: "Assess resource and wait existing related resources is already there",
			fields: fields{
				// evidenceStore: api.NewRPCConnection(testdata.MockGRPCTarget, evidence.NewEvidenceStoreClient, grpc.WithContextDialer(bufConnDialer)),
				// orchestrator:  api.NewRPCConnection(testdata.MockGRPCTarget, orchestrator.NewOrchestratorClient, grpc.WithContextDialer(bufConnDialer)),
				// authz:         servicetest.NewAuthorizationStrategy(true),
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
				in0: context.TODO(),
				evidence: &evidence.Evidence{
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
			wantResp: &assessment.AssessEvidenceResponse{
				Status: assessment.AssessmentStatus_ASSESSMENT_STATUS_ASSESSED,
			},
			wantErr: func(tt assert.TestingT, err error, i ...interface{}) bool {
				return assert.NoError(t, err)
			},
		},
		{
			name: "Assess resource and wait existing related resources is not there",
			fields: fields{
				// evidenceStore:       api.NewRPCConnection(testdata.MockGRPCTarget, evidence.NewEvidenceStoreClient, grpc.WithContextDialer(bufConnDialer)),
				// orchestrator:        api.NewRPCConnection(testdata.MockGRPCTarget, orchestrator.NewOrchestratorClient, grpc.WithContextDialer(bufConnDialer)),
				// authz:               servicetest.NewAuthorizationStrategy(true),
				evidenceResourceMap: make(map[string]*evidence.Evidence),
			},
			args: args{
				in0: context.TODO(),
				evidence: &evidence.Evidence{
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
			wantResp: &assessment.AssessEvidenceResponse{
				Status: assessment.AssessmentStatus_ASSESSMENT_STATUS_WAITING_FOR_RELATED,
			},
			wantErr: func(tt assert.TestingT, err error, i ...interface{}) bool {
				return assert.NoError(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// s := &Service{
			// 	orchestrator:         tt.fields.orchestrator,
			// 	orchestratorStreams:  api.NewStreamsOf(api.WithLogger[orchestrator.Orchestrator_StoreAssessmentResultsClient, *orchestrator.StoreAssessmentResultRequest](log)),
			// 	evidenceResourceMap:  tt.fields.evidenceResourceMap,
			// 	requests:             make(map[string]waitingRequest),
			// 	pe:                   policies.NewRegoEval(policies.WithPackageName(policies.DefaultRegoPackage)),
			// 	authz:                tt.fields.authz,
			// }
			svc, err := orchestratorsvc.NewService()
			assert.NoError(t, err)

			_, testSrv := servertest.NewTestConnectServer(t,
				server.WithHandler(
					orchestratorconnect.NewOrchestratorHandler(svc),
				),
			)
			defer testSrv.Close()

			s := NewService()
			gotResp, err := s.AssessEvidence(tt.args.in0, &assessment.AssessEvidenceRequest{Evidence: tt.args.evidence})

			tt.wantErr(t, err)
			// Check response
			assert.Equal(t, tt.wantResp, gotResp)
		})
	}
}

func TestService_handleEvidence(t *testing.T) {
	type fields struct {
		authz service.AuthorizationStrategy
		// evidenceStore *api.RPCConnection[evidence.EvidenceStoreClient]
		// orchestrator  *api.RPCConnection[orchestrator.OrchestratorClient]
	}
	type args struct {
		evidence *evidence.Evidence
		resource ontology.IsResource
		related  map[string]ontology.IsResource
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    assert.Want[[]*assessment.AssessmentResult]
		wantErr assert.WantErr
	}{
		{
			name:   "correct evidence: using metrics which return comparison results",
			fields: fields{
				// evidenceStore: api.NewRPCConnection(testdata.MockGRPCTarget, evidence.NewEvidenceStoreClient, grpc.WithContextDialer(bufConnDialer)),
				// orchestrator:  api.NewRPCConnection(testdata.MockGRPCTarget, orchestrator.NewOrchestratorClient, grpc.WithContextDialer(bufConnDialer)),
			},
			args: args{
				evidence: &evidence.Evidence{
					Id:                   testdata.MockEvidenceID1,
					ToolId:               testdata.MockEvidenceToolID1,
					Timestamp:            timestamppb.Now(),
					TargetOfEvaluationId: testdata.MockTargetOfEvaluationID1,
					Resource: prototest.NewProtobufResource(t, &ontology.Application{
						Id:   "Application",
						Name: "Application",
						Functionalities: []*ontology.Functionality{
							{
								Type: &ontology.Functionality_CryptographicHash{
									CryptographicHash: &ontology.CryptographicHash{
										Algorithm: "md5",
										UsesSalt:  false,
									},
								},
							},
						},
					}),
				},
				resource: &ontology.Application{
					Id:   "Application",
					Name: "Application",
					Functionalities: []*ontology.Functionality{
						{
							Type: &ontology.Functionality_CryptographicHash{
								CryptographicHash: &ontology.CryptographicHash{
									Algorithm: "md5",
									UsesSalt:  false,
								},
							},
						},
					},
				},
			},
			want: func(t *testing.T, got []*assessment.AssessmentResult, msgAndArgs ...any) bool {
				for _, result := range got {
					err := api.Validate(result)
					assert.NoError(t, err)
				}
				return assert.Equal(t, 3, len(got))
			},
			wantErr: assert.Nil[error],
		},
		{
			name:   "correct evidence: using metrics which do not return comparison results",
			fields: fields{
				// evidenceStore: api.NewRPCConnection(testdata.MockGRPCTarget, evidence.NewEvidenceStoreClient, grpc.WithContextDialer(bufConnDialer)),
				// orchestrator:  api.NewRPCConnection(testdata.MockGRPCTarget, orchestrator.NewOrchestratorClient, grpc.WithContextDialer(bufConnDialer)),
			},
			args: args{
				evidence: &evidence.Evidence{
					Id:                   testdata.MockEvidenceID1,
					ToolId:               testdata.MockEvidenceToolID1,
					Timestamp:            timestamppb.Now(),
					TargetOfEvaluationId: testdata.MockTargetOfEvaluationID1,
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
			},
			want: func(t *testing.T, got []*assessment.AssessmentResult, msgAndArgs ...any) bool {
				for _, result := range got {
					err := api.Validate(result)
					assert.NoError(t, err)
				}
				return assert.Equal(t, 9, len(got))
			},
			wantErr: assert.Nil[error],
		},
		{
			name:   "broken Any message",
			fields: fields{
				// evidenceStore: api.NewRPCConnection(testdata.MockGRPCTarget, evidence.NewEvidenceStoreClient, grpc.WithContextDialer(bufConnDialer)),
				// orchestrator:  api.NewRPCConnection(testdata.MockGRPCTarget, orchestrator.NewOrchestratorClient, grpc.WithContextDialer(bufConnDialer)),
			},
			args: args{
				evidence: &evidence.Evidence{
					Id:                   testdata.MockEvidenceID1,
					ToolId:               testdata.MockEvidenceToolID1,
					Timestamp:            timestamppb.Now(),
					TargetOfEvaluationId: testdata.MockTargetOfEvaluationID1,
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
			svc, err := orchestratorsvc.NewService()
			assert.NoError(t, err)

			_, testSrv := servertest.NewTestConnectServer(t,
				server.WithHandler(
					orchestratorconnect.NewOrchestratorHandler(svc),
				),
			)
			defer testSrv.Close()

			s := &Service{
				// orchestrator:         tt.fields.orchestrator,
				// orchestratorStreams:  api.NewStreamsOf(api.WithLogger[orchestrator.Orchestrator_StoreAssessmentResultsClient, *orchestrator.StoreAssessmentResultRequest](log)),
				// cachedConfigurations: make(map[string]cachedConfiguration),
				pe:    policies.NewRegoEval(policies.WithPackageName(policies.DefaultRegoPackage)),
				authz: tt.fields.authz,
			}

			results, err := s.handleEvidence(context.Background(), tt.args.evidence, tt.args.resource, tt.args.related)

			tt.wantErr(t, err)
			tt.want(t, results)
		})
	}
}
