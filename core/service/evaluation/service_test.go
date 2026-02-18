package evaluation

import (
	"context"
	"net/http"
	"testing"

	"confirmate.io/core/api/evaluation"
	"confirmate.io/core/api/evaluation/evaluationconnect"
	"confirmate.io/core/api/orchestrator"
	"confirmate.io/core/api/orchestrator/orchestratorconnect"
	"confirmate.io/core/persistence"
	"confirmate.io/core/persistence/persistencetest"
	"confirmate.io/core/service"
	"confirmate.io/core/service/evaluation/evaluationtest"
	"confirmate.io/core/util"
	"confirmate.io/core/util/assert"
	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestNewService(t *testing.T) {
	type args struct {
		opts []service.Option[Service]
	}
	tests := []struct {
		name string
		// Named input parameters for target function.
		args    args
		want    assert.Want[evaluationconnect.EvaluationHandler]
		wantErr assert.WantErr
	}{
		{
			name: "db error - creating db with invalid config",
			args: args{
				opts: []service.Option[Service]{
					WithConfig(Config{
						PersistenceConfig: persistence.Config{
							Host:             "localhost",
							Port:             5432,
							DBName:           "confirmate",
							User:             "confirmate",
							Password:         "confirmate",
							SSLMode:          "disable",
							MaxConn:          10,
							InMemoryDB:       false,
							Types:            []any{},
							CustomJoinTables: []persistence.CustomJoinTable{},
						}}),
				},
			},
			want: assert.Nil[evaluationconnect.EvaluationHandler],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.ErrorContains(t, err, "could not create db:")
			},
		},
		{
			name: "happy path with WithConfig option",
			args: args{
				opts: []service.Option[Service]{
					WithConfig(Config{
						PersistenceConfig:   persistence.DefaultConfig,
						OrchestratorClient:  http.DefaultClient,
						OrchestratorAddress: "http://testhost:8080",
					}),
				},
			},
			want: func(t *testing.T, got evaluationconnect.EvaluationHandler, msgAndArgs ...any) bool {
				assert.NotNil(t, got)
				svc, ok := got.(*Service)
				if !ok {
					t.Fatalf("expected *Service, got %T", got)
				}
				assert.NotEmpty(t, svc.db)
				assert.Equal(t, Config{
					OrchestratorAddress: "http://testhost:8080",
					OrchestratorClient:  http.DefaultClient,
					PersistenceConfig:   persistence.DefaultConfig,
				}, svc.cfg)
				assert.NotEmpty(t, svc.scheduler)
				assert.NotEmpty(t, orchestratorconnect.NewOrchestratorClient(svc.cfg.OrchestratorClient, "http:://testhost:8080"), svc.orchestratorClient)
				assert.Equal(t, make(map[string]map[string]*orchestrator.Control), svc.catalogControls)
				assert.NotNil(t, &svc.streamMutex)
				return assert.NotNil(t, &svc.catalogsMutex)
			},
			wantErr: assert.NoError,
		},
		{
			name: "happy path",
			args: args{
				opts: []service.Option[Service]{},
			},
			want: func(t *testing.T, got evaluationconnect.EvaluationHandler, msgAndArgs ...any) bool {
				assert.NotNil(t, got)
				svc, ok := got.(*Service)
				if !ok {
					t.Fatalf("expected *Service, got %T", got)
				}
				assert.NotEmpty(t, svc.db)
				assert.Equal(t, DefaultConfig, svc.cfg)
				assert.NotEmpty(t, svc.scheduler)
				assert.NotEmpty(t, orchestratorconnect.NewOrchestratorClient(svc.cfg.OrchestratorClient, svc.cfg.OrchestratorAddress), svc.orchestratorClient)
				assert.Equal(t, make(map[string]map[string]*orchestrator.Control), svc.catalogControls)
				assert.NotNil(t, &svc.streamMutex)
				return assert.NotNil(t, &svc.catalogsMutex)
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := NewService(tt.args.opts...)

			tt.want(t, got)
			tt.wantErr(t, gotErr)
		})
	}
}

func TestService_Shutdown(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "happy path",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewService()
			assert.NoError(t, err)

			svc, ok := got.(*Service)
			if !ok {
				t.Fatalf("expected *Service, got %T", got)
			}

			svc.Shutdown()
			assert.False(t, svc.scheduler.IsRunning())

		})
	}
}

func Test_validateCreateEvaluationResultRequest(t *testing.T) {
	type args struct {
		req *connect.Request[evaluation.CreateEvaluationResultRequest]
	}
	tests := []struct {
		name    string
		args    args
		wantErr assert.WantErr
	}{
		{
			name: "error: nil result",
			args: args{
				req: connect.NewRequest(&evaluation.CreateEvaluationResultRequest{}),
			},
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				assert.IsConnectError(t, err, connect.CodeInvalidArgument)
				return assert.ErrorContains(t, err, "invalid request")
			},
		},
		{
			name: "error: non-manual status",
			args: args{
				req: connect.NewRequest(&evaluation.CreateEvaluationResultRequest{
					Result: &evaluation.EvaluationResult{
						Id:                   "", // Empty ID - will be set by prep function
						TargetOfEvaluationId: evaluationtest.MockToeId1,
						AuditScopeId:         evaluationtest.MockAuditScopeId1,
						ControlId:            evaluationtest.MockControlId1,
						ControlCategoryName:  evaluationtest.MockCategoryName1,
						ControlCatalogId:     evaluationtest.MockCatalogId1,
						Status:               evaluation.EvaluationStatus_EVALUATION_STATUS_COMPLIANT,
						Timestamp:            timestamppb.Now(),
						ValidUntil:           timestamppb.Now(),
					},
				}),
			},
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.ErrorContains(t, err, "only manually set statuses are allowed")
			},
		},
		{
			name: "error: missing ValidUntil",
			args: args{
				req: connect.NewRequest(&evaluation.CreateEvaluationResultRequest{
					Result: &evaluation.EvaluationResult{
						Id:                   "", // Empty ID - will be set by prep function
						TargetOfEvaluationId: evaluationtest.MockToeId1,
						AuditScopeId:         evaluationtest.MockAuditScopeId1,
						ControlId:            evaluationtest.MockControlId1,
						ControlCategoryName:  evaluationtest.MockCategoryName1,
						ControlCatalogId:     evaluationtest.MockCatalogId1,
						Status:               evaluation.EvaluationStatus_EVALUATION_STATUS_COMPLIANT_MANUALLY,
						Timestamp:            timestamppb.Now(),
					},
				}),
			},
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.ErrorContains(t, err, "validity must be set")
			},
		},
		{
			name: "happy path: compliant manually",
			args: args{
				req: connect.NewRequest(&evaluation.CreateEvaluationResultRequest{
					Result: &evaluation.EvaluationResult{
						Id:                   "", // Empty ID - will be set by prep function
						TargetOfEvaluationId: evaluationtest.MockToeId1,
						AuditScopeId:         evaluationtest.MockAuditScopeId1,
						ControlId:            evaluationtest.MockControlId1,
						ControlCategoryName:  evaluationtest.MockCategoryName1,
						ControlCatalogId:     evaluationtest.MockCatalogId1,
						Status:               evaluation.EvaluationStatus_EVALUATION_STATUS_COMPLIANT_MANUALLY,
						Timestamp:            timestamppb.Now(),
						ValidUntil:           timestamppb.Now(),
						Comment:              util.Ref("Manual evaluation"),
					},
				}),
			},
			wantErr: assert.NoError,
		},
		{
			name: "happy path: not compliant manually",
			args: args{
				req: connect.NewRequest(&evaluation.CreateEvaluationResultRequest{
					Result: &evaluation.EvaluationResult{
						Id:                   "", // Empty ID - will be set by prep function
						TargetOfEvaluationId: evaluationtest.MockToeId2,
						AuditScopeId:         evaluationtest.MockAuditScopeId2,
						ControlId:            evaluationtest.MockControlId2,
						ControlCategoryName:  evaluationtest.MockCategoryName2,
						ControlCatalogId:     evaluationtest.MockCatalogId2,
						Status:               evaluation.EvaluationStatus_EVALUATION_STATUS_NOT_COMPLIANT_MANUALLY,
						Timestamp:            timestamppb.Now(),
						ValidUntil:           timestamppb.Now(),
					},
				}),
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := validateCreateEvaluationResultRequest(tt.args.req)

			tt.wantErr(t, gotErr)

			// Verify that ID was set by prep function when Result is not nil and validation passes
			if gotErr == nil && tt.args.req.Msg.Result != nil {
				assert.NotEmpty(t, tt.args.req.Msg.Result.Id)
			}
		})
	}
}

func TestService_ListEvaluationResults(t *testing.T) {
	type args struct {
		req *connect.Request[evaluation.ListEvaluationResultsRequest]
	}
	type fields struct {
		db persistence.DB
	}
	tests := []struct {
		name    string
		args    args
		fields  fields
		want    assert.Want[*connect.Response[evaluation.ListEvaluationResultsResponse]]
		wantErr assert.WantErr
	}{
		{
			name: "error: pagination error",
			args: args{
				req: &connect.Request[evaluation.ListEvaluationResultsRequest]{
					Msg: &evaluation.ListEvaluationResultsRequest{
						PageToken: "!!!invalid-base64!!!",
					},
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, []persistence.CustomJoinTable{}),
			},
			want: assert.Nil[*connect.Response[evaluation.ListEvaluationResultsResponse]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				assert.IsConnectError(t, err, connect.CodeInvalidArgument)
				return assert.ErrorContains(t, err, "could not decode page token")
			},
		},
		{
			name:   "error: validation error",
			args:   args{},
			fields: fields{},
			want:   assert.Nil[*connect.Response[evaluation.ListEvaluationResultsResponse]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeInvalidArgument) &&
					assert.ErrorContains(t, err, "empty request")
			},
		},
		// TODO(all): Results in an error when filtering by latest by control id -> Why! Error message: "Parsing error near <with>"
		// {
		// 	name: "happy path: filter by `get latest by control id`",
		// 	args: args{
		// 		req: connect.NewRequest(&evaluation.ListEvaluationResultsRequest{
		// 			LatestByControlId: util.Ref(true),
		// 		}),
		// 	},
		// 	fields: field{
		// 		db: persistencetest.NewInMemoryDB(t, types, []persistence.CustomJoinTable{}, func(d persistence.DB) {
		// 			err := d.Create(evaluationtest.MockEvaluationResult1)
		// 			assert.NoError(t, err)
		// 			err = d.Create(evaluationtest.MockEvaluationResult2)
		// 			assert.NoError(t, err)
		// 			err = d.Create(evaluationtest.MockEvaluationResult3)
		// 			assert.NoError(t, err)
		// 			err = d.Create(evaluationtest.MockEvaluationResult4)
		// 			assert.NoError(t, err)
		// 		}),
		// 	},
		// 	want: func(t *testing.T, got *connect.Response[evaluation.ListEvaluationResultsResponse], msgAndArgs ...any) bool {
		// 		assert.NotNil(t, got)
		// 		assert.Equal(t, 1, len(got.Msg.Results))
		// 		return assert.Equal(t, evaluationtest.MockEvaluationResult1, got.Msg.Results[0])
		// 	},
		// 	wantErr: assert.NoError,
		// },
		{
			name: "happy path: filter by `valid manual only`",
			args: args{
				req: connect.NewRequest(&evaluation.ListEvaluationResultsRequest{
					Filter: &evaluation.ListEvaluationResultsRequest_Filter{
						ValidManualOnly: util.Ref(true),
					},
				}),
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, []persistence.CustomJoinTable{}, func(d persistence.DB) {
					err := d.Create(evaluationtest.MockEvaluationResult1)
					assert.NoError(t, err)
					err = d.Create(evaluationtest.MockEvaluationResult2)
					assert.NoError(t, err)
					err = d.Create(evaluationtest.MockEvaluationResult3)
					assert.NoError(t, err)
					err = d.Create(evaluationtest.MockEvaluationResult4)
					assert.NoError(t, err)
				}),
			},
			want: func(t *testing.T, got *connect.Response[evaluation.ListEvaluationResultsResponse], msgAndArgs ...any) bool {
				assert.NotNil(t, got)
				assert.Equal(t, 1, len(got.Msg.Results))
				return assert.Equal(t, evaluationtest.MockEvaluationResult4, got.Msg.Results[0])
			},
			wantErr: assert.NoError,
		},
		{
			name: "happy path: filter by `parents only`",
			args: args{
				req: connect.NewRequest(&evaluation.ListEvaluationResultsRequest{
					Filter: &evaluation.ListEvaluationResultsRequest_Filter{
						ParentsOnly: util.Ref(true),
					},
				}),
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, []persistence.CustomJoinTable{}, func(d persistence.DB) {
					err := d.Create(evaluationtest.MockEvaluationResult1)
					assert.NoError(t, err)
					err = d.Create(evaluationtest.MockEvaluationResult2)
					assert.NoError(t, err)
					err = d.Create(evaluationtest.MockEvaluationResult3)
					assert.NoError(t, err)
				}),
			},
			want: func(t *testing.T, got *connect.Response[evaluation.ListEvaluationResultsResponse], msgAndArgs ...any) bool {
				assert.NotNil(t, got)
				assert.Equal(t, 2, len(got.Msg.Results))
				assert.Equal(t, evaluationtest.MockEvaluationResult1, got.Msg.Results[0])
				return assert.Equal(t, evaluationtest.MockEvaluationResult2, got.Msg.Results[1])
			},
			wantErr: assert.NoError,
		},
		{
			name: "happy path: filter by sub-control",
			args: args{
				req: connect.NewRequest(&evaluation.ListEvaluationResultsRequest{
					Filter: &evaluation.ListEvaluationResultsRequest_Filter{
						SubControls: util.Ref(string(evaluationtest.MockControlId11)),
					},
				}),
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, []persistence.CustomJoinTable{}, func(d persistence.DB) {
					err := d.Create(evaluationtest.MockEvaluationResult1)
					assert.NoError(t, err)
					err = d.Create(evaluationtest.MockEvaluationResult2)
					assert.NoError(t, err)
					err = d.Create(evaluationtest.MockEvaluationResult3)
					assert.NoError(t, err)
				}),
			},
			want: func(t *testing.T, got *connect.Response[evaluation.ListEvaluationResultsResponse], msgAndArgs ...any) bool {
				assert.NotNil(t, got)
				assert.Equal(t, 1, len(got.Msg.Results))
				return assert.Equal(t, evaluationtest.MockEvaluationResult3, got.Msg.Results[0])
			},
			wantErr: assert.NoError,
		},
		{
			name: "happy path: filter by control ID",
			args: args{
				req: connect.NewRequest(&evaluation.ListEvaluationResultsRequest{
					Filter: &evaluation.ListEvaluationResultsRequest_Filter{
						ControlId: util.Ref(string(evaluationtest.MockControlId2)),
					},
				}),
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, []persistence.CustomJoinTable{}, func(d persistence.DB) {
					err := d.Create(evaluationtest.MockEvaluationResult1)
					assert.NoError(t, err)
					err = d.Create(evaluationtest.MockEvaluationResult2)
					assert.NoError(t, err)
				}),
			},
			want: func(t *testing.T, got *connect.Response[evaluation.ListEvaluationResultsResponse], msgAndArgs ...any) bool {
				assert.NotNil(t, got)
				assert.Equal(t, 1, len(got.Msg.Results))
				return assert.Equal(t, evaluationtest.MockEvaluationResult2, got.Msg.Results[0])
			},
			wantErr: assert.NoError,
		},
		{
			name: "happy path: filter by catalog ID",
			args: args{
				req: connect.NewRequest(&evaluation.ListEvaluationResultsRequest{
					Filter: &evaluation.ListEvaluationResultsRequest_Filter{
						CatalogId: util.Ref(string(evaluationtest.MockCatalogId2)),
					},
				}),
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, []persistence.CustomJoinTable{}, func(d persistence.DB) {
					err := d.Create(evaluationtest.MockEvaluationResult1)
					assert.NoError(t, err)
					err = d.Create(evaluationtest.MockEvaluationResult2)
					assert.NoError(t, err)
				}),
			},
			want: func(t *testing.T, got *connect.Response[evaluation.ListEvaluationResultsResponse], msgAndArgs ...any) bool {
				assert.NotNil(t, got)
				assert.Equal(t, 1, len(got.Msg.Results))
				return assert.Equal(t, evaluationtest.MockEvaluationResult2, got.Msg.Results[0])
			},
			wantErr: assert.NoError,
		},
		{
			name: "happy path: filter by ToE",
			args: args{
				req: connect.NewRequest(&evaluation.ListEvaluationResultsRequest{
					Filter: &evaluation.ListEvaluationResultsRequest_Filter{
						TargetOfEvaluationId: util.Ref(evaluationtest.MockToeId2),
					},
				}),
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, []persistence.CustomJoinTable{}, func(d persistence.DB) {
					err := d.Create(evaluationtest.MockEvaluationResult1)
					assert.NoError(t, err)
					err = d.Create(evaluationtest.MockEvaluationResult2)
					assert.NoError(t, err)
				}),
			},
			want: func(t *testing.T, got *connect.Response[evaluation.ListEvaluationResultsResponse], msgAndArgs ...any) bool {
				assert.NotNil(t, got)
				assert.Equal(t, 1, len(got.Msg.Results))
				return assert.Equal(t, evaluationtest.MockEvaluationResult2, got.Msg.Results[0])
			},
			wantErr: assert.NoError,
		},
		{
			name: "happy path",
			args: args{
				req: connect.NewRequest(&evaluation.ListEvaluationResultsRequest{}),
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, []persistence.CustomJoinTable{}, func(d persistence.DB) {
					err := d.Create(evaluationtest.MockEvaluationResult1)
					assert.NoError(t, err)
					err = d.Create(evaluationtest.MockEvaluationResult2)
					assert.NoError(t, err)
				}),
			},
			want: func(t *testing.T, got *connect.Response[evaluation.ListEvaluationResultsResponse], msgAndArgs ...any) bool {
				assert.NotNil(t, got)
				assert.Equal(t, 2, len(got.Msg.Results))
				assert.Equal(t, evaluationtest.MockEvaluationResult1, got.Msg.Results[0])
				return assert.Equal(t, evaluationtest.MockEvaluationResult2, got.Msg.Results[1])
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &Service{
				db: tt.fields.db,
			}
			got, gotErr := svc.ListEvaluationResults(context.Background(), tt.args.req)

			tt.want(t, got)
			tt.wantErr(t, gotErr)
		})
	}
}
