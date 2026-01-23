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
	"confirmate.io/core/policies"
	"confirmate.io/core/stream"
	"connectrpc.com/connect"
	"github.com/google/uuid"

	"confirmate.io/core/api/orchestrator/orchestratorconnect"
	"confirmate.io/core/server"
	"confirmate.io/core/server/servertest"
	"confirmate.io/core/service"
	orchestratorsvc "confirmate.io/core/service/orchestrator"
	"confirmate.io/core/util/assert"
	"confirmate.io/core/util/prototest"
	"confirmate.io/core/util/testdata"
	"google.golang.org/protobuf/types/known/durationpb"
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

			svc, err := orchestratorsvc.NewService()
			assert.NoError(t, err)

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
		name    string
		fields  fields
		args    args
		want    assert.Want[*connect.Response[assessment.AssessEvidenceResponse]]
		wantErr assert.WantErr
	}{
		{
			name: "Missing evidence",
			want: nil,
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
			want: nil,
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
			want: nil,
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
			want: nil,
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
			want: nil,
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				cErr := assert.Is[*connect.Error](t, err)
				return assert.Equal(t, connect.CodeInvalidArgument, cErr.Code())
			},
		},
		{
			name: "Assess resource happy",
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
		{
			name: "Assess resource of wrong cloud service",
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
						TargetOfEvaluationId: testdata.MockTargetOfEvaluationID1},
				},
			},
			want: nil,
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				cErr := assert.Is[*connect.Error](t, err)
				return assert.Equal(t, connect.CodePermissionDenied, cErr.Code())
			},
		},
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
			want: nil,
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				cErr := assert.Is[*connect.Error](t, err)
				return assert.Equal(t, connect.CodeInvalidArgument, cErr.Code())
			},
		},
		{
			name: "Assess resource and wait existing related resources is already there",
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
			name: "Assess resource and wait, existing related resources are not there",
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
				return assert.Equal(t, assessment.AssessmentStatus_ASSESSMENT_STATUS_ASSESSED, got.Msg.Status)
			},
			wantErr: assert.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, err := NewService()
			assert.NoError(t, err)
			res, err := s.AssessEvidence(context.Background(), connect.NewRequest(tt.args.req))
			tt.want(t, res)
			tt.wantErr(t, err)
		})
	}
}

func TestService_AssessEvidences(t *testing.T) {
	type args struct {
		stream *connect.BidiStreamForClient[assessment.AssessEvidenceRequest, assessment.AssessEvidencesResponse]
		req    *assessment.AssessEvidenceRequest
	}
	type fields struct {
		svc *Service
	}
	tests := []struct {
		name    string
		args    args
		fields  fields
		want    assert.Want[*assessment.AssessEvidencesResponse]
		wantErr assert.WantErr
	}{
		{
			name: "Missing toolId",
			args: args{
				req: &assessment.AssessEvidenceRequest{
					Evidence: &evidence.Evidence{
						Id:                   testdata.MockEvidenceID1,
						Timestamp:            timestamppb.Now(),
						TargetOfEvaluationId: testdata.MockTargetOfEvaluationID1,
						Resource:             prototest.NewProtobufResource(t, &ontology.VirtualMachine{Id: testdata.MockVirtualMachineID1}),
					},
				},
			},
			want: func(t *testing.T, got *assessment.AssessEvidencesResponse, args ...any) bool {
				assert.Equal(t, assessment.AssessmentStatus_ASSESSMENT_STATUS_FAILED, got.Status)
				return assert.Contains(t, got.StatusMessage, "evidence.tool_id: value length must be at least 1 characters")
			},
			wantErr: assert.NoError,
		},
		{
			name: "Missing evidenceID",
			args: args{
				req: &assessment.AssessEvidenceRequest{
					Evidence: &evidence.Evidence{
						Timestamp:            timestamppb.Now(),
						ToolId:               testdata.MockEvidenceToolID1,
						TargetOfEvaluationId: testdata.MockTargetOfEvaluationID1,
						Resource:             prototest.NewProtobufResource(t, &ontology.VirtualMachine{Id: testdata.MockVirtualMachineID1}),
					},
				},
			},
			wantErr: assert.NoError,
			want: func(t *testing.T, got *assessment.AssessEvidencesResponse, args ...any) bool {
				assert.Equal(t, assessment.AssessmentStatus_ASSESSMENT_STATUS_FAILED, got.Status)
				return assert.Contains(t, got.StatusMessage, "evidence.id: value is empty, which is not a valid UUID")
			},
		},
		{
			name: "Assess evidences",
			args: args{
				req: &assessment.AssessEvidenceRequest{
					Evidence: &evidence.Evidence{
						Id:                   testdata.MockEvidenceID1,
						Timestamp:            timestamppb.Now(),
						ToolId:               testdata.MockEvidenceToolID1,
						TargetOfEvaluationId: testdata.MockTargetOfEvaluationID1,
						Resource: prototest.NewProtobufResource(t, &ontology.VirtualMachine{
							Id:   testdata.MockVirtualMachineID1,
							Name: testdata.MockVirtualMachineName1,
						}),
					},
				},
			},
			want: func(t *testing.T, got *assessment.AssessEvidencesResponse, args ...any) bool {
				assert.Equal(t, assessment.AssessmentStatus_ASSESSMENT_STATUS_ASSESSED, got.Status)
				return assert.Empty(t, got.Status)
			},
			wantErr: assert.NoError,
		},
		{
			name: "Error in stream to client - Send()-err",
			args: args{
				req: &assessment.AssessEvidenceRequest{
					Evidence: &evidence.Evidence{
						Timestamp:            timestamppb.Now(),
						ToolId:               testdata.MockEvidenceToolID1,
						TargetOfEvaluationId: testdata.MockTargetOfEvaluationID1,
						Resource:             prototest.NewProtobufResource(t, &ontology.VirtualMachine{Id: testdata.MockVirtualMachineID1}),
					},
				},
			},
			want: assert.Nil[*assessment.AssessEvidencesResponse],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.ErrorContains(t, err, "rpc error: code = Unknown desc = cannot send response to the client")
			},
		},
		{
			name: "Error in stream to server - Recv()-err",
			args: args{
				req: &assessment.AssessEvidenceRequest{
					Evidence: &evidence.Evidence{
						Timestamp:            timestamppb.Now(),
						ToolId:               testdata.MockEvidenceToolID1,
						TargetOfEvaluationId: testdata.MockTargetOfEvaluationID1,
						Resource:             prototest.NewProtobufResource(t, &ontology.VirtualMachine{Id: testdata.MockVirtualMachineID1}),
					},
				},
			},
			want: assert.Nil[*assessment.AssessEvidencesResponse],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.ErrorContains(t, err, "rpc error: code = Unknown desc = cannot receive stream request")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create service
			svc, err := NewService()
			assert.NoError(t, err)

			// Create an initial test server
			_, testSrv1 := servertest.NewTestConnectServer(t,
				server.WithHandler(
					assessmentconnect.NewAssessmentHandler(svc),
				),
			)
			serverURL := testSrv1.URL

			// Retrieve port, so we can restart the server later on the same port
			// port := testSrv1.Listener.Addr().(*net.TCPAddr).Port

			httpClient := testSrv1.Client()

			client := assessmentconnect.NewAssessmentClient(httpClient, serverURL)
			factory := func(ctx context.Context) *connect.BidiStreamForClient[assessment.AssessEvidenceRequest, assessment.AssessEvidencesResponse] {
				return client.AssessEvidenceStream(ctx)
			}
			ctx := context.Background()
			rs, err := stream.NewRestartableBidiStream(ctx, factory, stream.DefaultRestartConfig())
			assert.NoError(t, err)

			err = rs.Send(tt.args.req)
			// err := tt.fields.svc.AssessEvidenceStream(context.Background(), tt.args.stream)
			tt.wantErr(t, err)

			res, err := rs.Receive()
			// TODO: only handle the last response?
			tt.want(t, res)
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
			name: "correct evidence: using metrics which return comparison results",
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
			wantErr: assert.NoError,
		},
		{
			name: "correct evidence: using metrics which do not return comparison results",
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
			wantErr: assert.NoError,
		},
		{
			name: "broken Any message",
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
			s := &Service{}

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
	s := &Service{
		cachedConfigurations: make(map[string]cachedConfiguration),
		evidenceResourceMap:  make(map[string]*evidence.Evidence),
		pe:                   policies.NewRegoEval(policies.WithPackageName(policies.DefaultRegoPackage)),
	}
	// First assess evidence with a valid VM resource s.t. the cache is created for the combination of resource type and
	// tool id (="VirtualMachine-{testdata.MockEvidenceToolID}")
	e := &evidence.Evidence{
		Id:                   testdata.MockEvidenceID1,
		ToolId:               testdata.MockEvidenceToolID1,
		Timestamp:            timestamppb.Now(),
		TargetOfEvaluationId: testdata.MockTargetOfEvaluationID1,
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
	_, err := s.AssessEvidence(context.Background(), connect.NewRequest(&assessment.AssessEvidenceRequest{Evidence: e}))
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
				TargetOfEvaluationId: testdata.MockTargetOfEvaluationID1,
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
		hookCounts      = 20
	)

	wg.Add(hookCounts)

	firstHookFunction := func(ctx context.Context, assessmentResult *assessment.AssessmentResult, err error) {
		hookCallCounter++
		logger.Info("Hello from inside the firstHookFunction")
		wg.Done()
	}

	secondHookFunction := func(ctx context.Context, assessmentResult *assessment.AssessmentResult, err error) {
		hookCallCounter++
		logger.Info("Hello from inside the secondHookFunction")
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
						TargetOfEvaluationId: testdata.MockTargetOfEvaluationID1,
						Resource: prototest.NewProtobufResource(t, &ontology.VirtualMachine{
							Id:   testdata.MockVirtualMachineID1,
							Name: testdata.MockVirtualMachineName1,
							BootLogging: &ontology.BootLogging{
								LoggingServiceIds: []string{"SomeResourceId2"},
								Enabled:           true,
								RetentionPeriod:   durationpb.New(time.Hour * 24 * 36),
							},
							OsLogging: &ontology.OSLogging{
								LoggingServiceIds: []string{"SomeResourceId2"},
								Enabled:           true,
								RetentionPeriod:   durationpb.New(time.Hour * 24 * 36),
							},
							MalwareProtection: &ontology.MalwareProtection{
								Enabled:              true,
								NumberOfThreatsFound: 5,
								DurationSinceActive:  durationpb.New(time.Hour * 24 * 20),
								ApplicationLogging: &ontology.ApplicationLogging{
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
			s := &Service{}

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
