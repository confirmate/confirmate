package evaluation

import (
	"context"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"testing"

	"confirmate.io/core/api/assessment"
	"confirmate.io/core/api/evaluation"
	"confirmate.io/core/api/evaluation/evaluationconnect"
	"confirmate.io/core/api/orchestrator"
	"confirmate.io/core/api/orchestrator/orchestratorconnect"
	"confirmate.io/core/persistence"
	"confirmate.io/core/persistence/persistencetest"
	"confirmate.io/core/service"
	"confirmate.io/core/service/evaluation/evaluationtest"
	"confirmate.io/core/service/evidence/evidencetest"
	"confirmate.io/core/service/orchestrator/orchestratortest"
	"confirmate.io/core/util"
	"confirmate.io/core/util/assert"
	"connectrpc.com/connect"
	"github.com/google/uuid"
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

func TestService_CreateEvaluationResult(t *testing.T) {
	type args struct {
		req *connect.Request[evaluation.CreateEvaluationResultRequest]
	}
	type fields struct {
		db persistence.DB
	}
	tests := []struct {
		name    string
		args    args
		fields  fields
		want    assert.Want[*connect.Response[evaluation.EvaluationResult]]
		wantErr assert.WantErr
	}{
		{
			name: "error: nil request",
			args: args{
				req: nil,
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, []persistence.CustomJoinTable{}),
			},
			want: assert.Nil[*connect.Response[evaluation.EvaluationResult]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				assert.IsConnectError(t, err, connect.CodeInvalidArgument)
				return assert.ErrorContains(t, err, "empty request")
			},
		},
		{
			name: "error: non-manual status (compliant)",
			args: args{
				req: connect.NewRequest(&evaluation.CreateEvaluationResultRequest{
					Result: &evaluation.EvaluationResult{
						Id:                   "",
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
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, []persistence.CustomJoinTable{}),
			},
			want: assert.Nil[*connect.Response[evaluation.EvaluationResult]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				assert.IsConnectError(t, err, connect.CodeInvalidArgument)
				return assert.ErrorContains(t, err, "only manually set statuses are allowed")
			},
		},
		{
			name: "happy path: compliant manually",
			args: args{
				req: connect.NewRequest(&evaluation.CreateEvaluationResultRequest{
					Result: &evaluation.EvaluationResult{
						Id:                   "",
						TargetOfEvaluationId: evaluationtest.MockToeId1,
						AuditScopeId:         evaluationtest.MockAuditScopeId1,
						ControlId:            evaluationtest.MockControlId1,
						ControlCategoryName:  evaluationtest.MockCategoryName1,
						ControlCatalogId:     evaluationtest.MockCatalogId1,
						Status:               evaluation.EvaluationStatus_EVALUATION_STATUS_COMPLIANT_MANUALLY,
						Timestamp:            timestamppb.Now(),
						ValidUntil:           timestamppb.Now(),
						Comment:              util.Ref("Manual evaluation - compliant"),
					},
				}),
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, []persistence.CustomJoinTable{}),
			},
			want: func(t *testing.T, got *connect.Response[evaluation.EvaluationResult], msgAndArgs ...any) bool {
				assert.NotNil(t, got)
				assert.NotEmpty(t, got.Msg.Id)
				assert.NoError(t, uuid.Validate(got.Msg.Id))
				assert.Equal(t, evaluationtest.MockToeId1, got.Msg.TargetOfEvaluationId)
				assert.Equal(t, evaluationtest.MockAuditScopeId1, got.Msg.AuditScopeId)
				assert.Equal(t, evaluationtest.MockControlId1, got.Msg.ControlId)
				assert.Equal(t, evaluationtest.MockCategoryName1, got.Msg.ControlCategoryName)
				assert.Equal(t, evaluationtest.MockCatalogId1, got.Msg.ControlCatalogId)
				assert.Equal(t, evaluation.EvaluationStatus_EVALUATION_STATUS_COMPLIANT_MANUALLY, got.Msg.Status)
				assert.Equal(t, "Manual evaluation - compliant", util.Deref(got.Msg.Comment))
				return true
			},
			wantErr: assert.NoError,
		},
		{
			name: "happy path: not compliant manually",
			args: args{
				req: connect.NewRequest(&evaluation.CreateEvaluationResultRequest{
					Result: &evaluation.EvaluationResult{
						Id:                   "",
						TargetOfEvaluationId: evaluationtest.MockToeId2,
						AuditScopeId:         evaluationtest.MockAuditScopeId2,
						ControlId:            evaluationtest.MockControlId2,
						ControlCategoryName:  evaluationtest.MockCategoryName2,
						ControlCatalogId:     evaluationtest.MockCatalogId2,
						Status:               evaluation.EvaluationStatus_EVALUATION_STATUS_NOT_COMPLIANT_MANUALLY,
						Timestamp:            timestamppb.Now(),
						ValidUntil:           timestamppb.Now(),
						Comment:              util.Ref("Manual evaluation - not compliant"),
					},
				}),
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, []persistence.CustomJoinTable{}),
			},
			want: func(t *testing.T, got *connect.Response[evaluation.EvaluationResult], msgAndArgs ...any) bool {
				assert.NotNil(t, got)
				assert.NotEmpty(t, got.Msg.Id)
				assert.Equal(t, evaluationtest.MockToeId2, got.Msg.TargetOfEvaluationId)
				assert.Equal(t, evaluationtest.MockAuditScopeId2, got.Msg.AuditScopeId)
				assert.Equal(t, evaluationtest.MockControlId2, got.Msg.ControlId)
				assert.Equal(t, evaluationtest.MockCategoryName2, got.Msg.ControlCategoryName)
				assert.Equal(t, evaluationtest.MockCatalogId2, got.Msg.ControlCatalogId)
				assert.Equal(t, evaluation.EvaluationStatus_EVALUATION_STATUS_NOT_COMPLIANT_MANUALLY, got.Msg.Status)
				assert.Equal(t, "Manual evaluation - not compliant", util.Deref(got.Msg.Comment))
				return true
			},
			wantErr: assert.NoError,
		},
		{
			name: "happy path: with large file blob in data field",
			args: args{
				req: connect.NewRequest(&evaluation.CreateEvaluationResultRequest{
					Result: &evaluation.EvaluationResult{
						Id:                   "",
						TargetOfEvaluationId: evaluationtest.MockToeId1,
						AuditScopeId:         evaluationtest.MockAuditScopeId1,
						ControlId:            evaluationtest.MockControlId1,
						ControlCategoryName:  evaluationtest.MockCategoryName1,
						ControlCatalogId:     evaluationtest.MockCatalogId1,
						Status:               evaluation.EvaluationStatus_EVALUATION_STATUS_COMPLIANT_MANUALLY,
						Timestamp:            timestamppb.Now(),
						ValidUntil:           timestamppb.Now(),
						Comment:              util.Ref("Manual evaluation with large data"),
						Data:                 make([]byte, 1024*1024), // 1MB data blob
					},
				}),
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, []persistence.CustomJoinTable{}),
			},
			want: func(t *testing.T, got *connect.Response[evaluation.EvaluationResult], msgAndArgs ...any) bool {
				assert.NotNil(t, got)
				assert.NotEmpty(t, got.Msg.Id)
				assert.Equal(t, evaluationtest.MockToeId1, got.Msg.TargetOfEvaluationId)
				assert.Equal(t, evaluationtest.MockAuditScopeId1, got.Msg.AuditScopeId)
				assert.Equal(t, evaluationtest.MockControlId1, got.Msg.ControlId)
				assert.Equal(t, evaluation.EvaluationStatus_EVALUATION_STATUS_COMPLIANT_MANUALLY, got.Msg.Status)
				assert.Equal(t, 1024*1024, len(got.Msg.Data))
				return true
			},
			wantErr: assert.NoError,
		},
		{
			name: "happy path: with assessment result IDs",
			args: args{
				req: connect.NewRequest(&evaluation.CreateEvaluationResultRequest{
					Result: &evaluation.EvaluationResult{
						Id:                   "",
						TargetOfEvaluationId: evaluationtest.MockToeId1,
						AuditScopeId:         evaluationtest.MockAuditScopeId1,
						ControlId:            evaluationtest.MockControlId1,
						ControlCategoryName:  evaluationtest.MockCategoryName1,
						ControlCatalogId:     evaluationtest.MockCatalogId1,
						Status:               evaluation.EvaluationStatus_EVALUATION_STATUS_COMPLIANT_MANUALLY,
						Timestamp:            timestamppb.Now(),
						ValidUntil:           timestamppb.Now(),
						AssessmentResultIds:  []string{evaluationtest.MockAssessmentResultId1, evaluationtest.MockAssessmentResultId2},
						Comment:              util.Ref("Manual evaluation with assessment results"),
					},
				}),
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, []persistence.CustomJoinTable{}),
			},
			want: func(t *testing.T, got *connect.Response[evaluation.EvaluationResult], msgAndArgs ...any) bool {
				assert.NotNil(t, got)
				assert.NotEmpty(t, got.Msg.Id)
				assert.Equal(t, evaluationtest.MockToeId1, got.Msg.TargetOfEvaluationId)
				assert.Equal(t, 2, len(got.Msg.AssessmentResultIds))
				assert.Equal(t, evaluationtest.MockAssessmentResultId1, got.Msg.AssessmentResultIds[0])
				assert.Equal(t, evaluationtest.MockAssessmentResultId2, got.Msg.AssessmentResultIds[1])
				return true
			},
			wantErr: assert.NoError,
		},
		{
			name: "happy path: with parent control ID",
			args: args{
				req: connect.NewRequest(&evaluation.CreateEvaluationResultRequest{
					Result: &evaluation.EvaluationResult{
						Id:                   "",
						TargetOfEvaluationId: evaluationtest.MockToeId1,
						AuditScopeId:         evaluationtest.MockAuditScopeId1,
						ControlId:            evaluationtest.MockControlId11,
						ParentControlId:      util.Ref(evaluationtest.MockControlId1),
						ControlCategoryName:  evaluationtest.MockCategoryName1,
						ControlCatalogId:     evaluationtest.MockCatalogId1,
						Status:               evaluation.EvaluationStatus_EVALUATION_STATUS_COMPLIANT_MANUALLY,
						Timestamp:            timestamppb.Now(),
						ValidUntil:           timestamppb.Now(),
						Comment:              util.Ref("Manual evaluation for sub-control"),
					},
				}),
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, []persistence.CustomJoinTable{}),
			},
			want: func(t *testing.T, got *connect.Response[evaluation.EvaluationResult], msgAndArgs ...any) bool {
				assert.NotNil(t, got)
				assert.NotEmpty(t, got.Msg.Id)
				assert.Equal(t, evaluationtest.MockControlId11, got.Msg.ControlId)
				assert.Equal(t, evaluationtest.MockControlId1, util.Deref(got.Msg.ParentControlId))
				assert.Equal(t, evaluation.EvaluationStatus_EVALUATION_STATUS_COMPLIANT_MANUALLY, got.Msg.Status)
				return true
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &Service{
				db: tt.fields.db,
			}
			got, gotErr := svc.CreateEvaluationResult(context.Background(), tt.args.req)

			tt.want(t, got)
			tt.wantErr(t, gotErr)
		})
	}
}

// TestService_CreateEvaluationResult_LargeBlobIntegration is an integration test that verifies large binary data blobs
// are correctly stored and retrieved from the database. This test ensures the persistence layer (GORM) properly handles
// large binary data (1MB+) in the evaluation result's Data field, which is used for giving additional justification.
func TestService_CreateEvaluationResult_LargeBlobIntegration(t *testing.T) {
	// Create an in-memory database for testing
	db := persistencetest.NewInMemoryDB(t, types, []persistence.CustomJoinTable{})

	// Initialize the service with the test database
	svc := &Service{
		db: db,
	}

	// Create a 1MB blob with known pattern (alternating 0xAA and 0x55)
	// This pattern helps verify data integrity - if the blob is corrupted,
	// the pattern will be broken
	largeBlob := make([]byte, 1024*1024)
	for i := range largeBlob {
		if i%2 == 0 {
			largeBlob[i] = 0xAA
		} else {
			largeBlob[i] = 0x55
		}
	}

	// Create a request with a large blob in the Data field
	req := connect.NewRequest(&evaluation.CreateEvaluationResultRequest{
		Result: &evaluation.EvaluationResult{
			Id:                   "", // Empty ID - will be generated by validation
			TargetOfEvaluationId: evaluationtest.MockToeId1,
			AuditScopeId:         evaluationtest.MockAuditScopeId1,
			ControlId:            evaluationtest.MockControlId1,
			ControlCategoryName:  evaluationtest.MockCategoryName1,
			ControlCatalogId:     evaluationtest.MockCatalogId1,
			Status:               evaluation.EvaluationStatus_EVALUATION_STATUS_COMPLIANT_MANUALLY,
			Timestamp:            timestamppb.Now(),
			ValidUntil:           timestamppb.Now(),
			Comment:              util.Ref("Integration test for large blob storage"),
			Data:                 largeBlob,
		},
	})

	// Call CreateEvaluationResult to store the result with large blob
	res, err := svc.CreateEvaluationResult(context.Background(), req)

	// Verify the creation was successful
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.NotEmpty(t, res.Msg.Id)

	// Verify the response contains the correct blob size
	assert.Equal(t, 1024*1024, len(res.Msg.Data))

	// Now retrieve the record directly from the database to verify persistence
	// This ensures the blob was actually written to disk/storage, not just kept in memory
	var retrieved evaluation.EvaluationResult
	err = db.Get(&retrieved, "id = ?", res.Msg.Id)
	assert.NoError(t, err)

	// Verify the blob size is preserved in the database
	assert.Equal(t, 1024*1024, len(retrieved.Data))

	// Verify the blob content is identical (byte-for-byte comparison)
	// This ensures no data corruption occurred during storage
	assert.Equal(t, largeBlob, retrieved.Data)
}

func TestService_evaluateSubcontrol(t *testing.T) {
	type fields struct {
		orchestratorClient orchestratorconnect.OrchestratorClient
		db                 persistence.DB
		catalogControls    map[string]map[string]*orchestrator.Control
	}
	type args struct {
		ctx        context.Context
		auditScope *orchestrator.AuditScope
		control    *orchestrator.Control
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    assert.Want[*evaluation.EvaluationResult]
		wantSvc assert.Want[*Service]
	}{
		{
			name: "Audit Scope input empty", // we do not check the other input parameters
			fields: fields{
				orchestratorClient: nil,
				db:                 persistencetest.NewInMemoryDB(t, types, []persistence.CustomJoinTable{}),
				catalogControls: map[string]map[string]*orchestrator.Control{
					orchestratortest.MockControl1.GetCategoryCatalogId(): {
						fmt.Sprintf("%s-%s", orchestratortest.MockControl1.GetCategoryName(), orchestratortest.MockControl1.GetId()): orchestratortest.MockControl1,
					},
				},
			},
			args: args{
				control: orchestratortest.MockControl1,
			},
			wantSvc: func(t *testing.T, got *Service, _ ...any) bool {
				evalResults, err := got.ListEvaluationResults(context.Background(), connect.NewRequest(&evaluation.ListEvaluationResultsRequest{}))
				assert.NoError(t, err)
				return assert.Equal(t, 0, len(evalResults.Msg.Results))
			},
		},
		{
			name: "no assessment results available",
			fields: fields{
				orchestratorClient: nil,
				db:                 persistencetest.NewInMemoryDB(t, types, []persistence.CustomJoinTable{}),
			},
			args: args{
				auditScope: &orchestrator.AuditScope{
					TargetOfEvaluationId: orchestratortest.MockToeId1,
					CatalogId:            orchestratortest.MockCatalogId1,
					AssuranceLevel:       util.Ref("high"),
				},
				control: &orchestrator.Control{
					Id:           orchestratortest.MockControlId1,
					CategoryName: orchestratortest.MockCategoryName1,
				},
			},
			wantSvc: func(t *testing.T, got *Service, _ ...any) bool {
				evalResults, err := got.ListEvaluationResults(context.Background(), connect.NewRequest(&evaluation.ListEvaluationResultsRequest{}))
				assert.NoError(t, err)
				return assert.Equal(t, 0, len(evalResults.Msg.Results))
			},
		},
		{
			name: "error getting metrics",
			fields: fields{
				orchestratorClient: nil,
				db:                 persistencetest.NewInMemoryDB(t, types, []persistence.CustomJoinTable{}),
			},
			args: args{
				auditScope: &orchestrator.AuditScope{
					TargetOfEvaluationId: orchestratortest.MockToeId1,
					CatalogId:            orchestratortest.MockCatalogId1,
					AssuranceLevel:       util.Ref("high"),
				},
				control: &orchestrator.Control{
					Id:           "Cont1.1",
					CategoryName: orchestratortest.MockCategoryName1,
				},
			},
			wantSvc: func(t *testing.T, got *Service, _ ...any) bool {
				evalResults, err := got.ListEvaluationResults(context.Background(), connect.NewRequest(&evaluation.ListEvaluationResultsRequest{}))
				assert.NoError(t, err)
				return assert.Equal(t, 0, len(evalResults.Msg.Results))
			},
		},
		{
			name: "error getting assessment results",
			fields: fields{
				orchestratorClient: nil,
				db:                 persistencetest.NewInMemoryDB(t, types, []persistence.CustomJoinTable{}),
			},
			args: args{
				auditScope: &orchestrator.AuditScope{
					CatalogId:      orchestratortest.MockCatalogId1,
					AssuranceLevel: util.Ref("high"),
				},
				control: &orchestrator.Control{
					Id:           "Cont1.1",
					CategoryName: orchestratortest.MockCategoryName1,
				},
			},
			wantSvc: func(t *testing.T, got *Service, _ ...any) bool {
				evalResults, err := got.ListEvaluationResults(context.Background(), connect.NewRequest(&evaluation.ListEvaluationResultsRequest{}))
				assert.NoError(t, err)
				return assert.Equal(t, 0, len(evalResults.Msg.Results))
			},
		},
		{
			name: "Happy path",
			fields: fields{
				orchestratorClient: nil,
				db: persistencetest.NewInMemoryDB(t, types, []persistence.CustomJoinTable{}, func(db persistence.DB) {
					assert.NoError(t, db.Create(orchestratortest.MockAssessmentResult1))
					assert.NoError(t, db.Create(orchestratortest.MockAssessmentResult2))
				}),
				catalogControls: map[string]map[string]*orchestrator.Control{
					orchestratortest.MockControl1.GetCategoryCatalogId(): {
						fmt.Sprintf("%s-%s", orchestratortest.MockControl1.GetCategoryName(), orchestratortest.MockControl1.GetId()):       orchestratortest.MockControl1,
						fmt.Sprintf("%s-%s", orchestratortest.MockSubControl1.GetCategoryName(), orchestratortest.MockSubControl1.GetId()): orchestratortest.MockSubControl1,
					},
				},
			},
			args: args{
				auditScope: &orchestrator.AuditScope{
					TargetOfEvaluationId: orchestratortest.MockToeId1,
					CatalogId:            orchestratortest.MockCatalogId1,
					AssuranceLevel:       util.Ref("high"),
				},
				control: &orchestrator.Control{
					Id:           "Cont1.1",
					CategoryName: orchestratortest.MockCategoryName1,
				},
			},
			wantSvc: func(t *testing.T, got *Service, _ ...any) bool {
				evalResults, err := got.ListEvaluationResults(context.Background(), connect.NewRequest(&evaluation.ListEvaluationResultsRequest{}))
				assert.NoError(t, err)
				return assert.Equal(t, 1, len(evalResults.Msg.Results))
			},
			want: func(t *testing.T, got *evaluation.EvaluationResult, _ ...any) bool {
				return assert.Equal(t, "Cont1.1", got.ControlId)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &Service{
				orchestratorClient: tt.fields.orchestratorClient,
				db:                 tt.fields.db,
				catalogControls:    tt.fields.catalogControls,
			}

			got, _ := svc.evaluateSubcontrol(tt.args.ctx, tt.args.auditScope, tt.args.control)
			assert.Optional(t, tt.want, got)
			assert.Optional(t, tt.wantSvc, svc)
		})
	}
}

func TestService_getAllMetricsFromControl(t *testing.T) {
	type fields struct {
		catalogControls map[string]map[string]*orchestrator.Control
	}
	type args struct {
		catalogId    string
		categoryName string
		controlId    string
	}
	tests := []struct {
		name        string
		fields      fields
		args        args
		wantMetrics []*assessment.Metric
		wantErr     assert.WantErr
	}{
		{
			name:        "Input empty",
			fields:      fields{},
			wantMetrics: nil,
			wantErr: func(t *testing.T, err error, i ...interface{}) bool {
				return assert.ErrorContains(t, err, "could not get control for control id")
			},
		},
		{
			name: "no sub-controls available",
			fields: fields{
				catalogControls: map[string]map[string]*orchestrator.Control{
					orchestratortest.MockCatalogId1: {
						fmt.Sprintf("%s-%s", orchestratortest.MockControl2.GetCategoryName(), orchestratortest.MockControl2.GetId()): orchestratortest.MockControl2,
					},
				},
			},
			args: args{
				catalogId:    orchestratortest.MockCatalogId1,
				categoryName: orchestratortest.MockCategoryName2,
				controlId:    orchestratortest.MockControlId2,
			},
			wantMetrics: nil,
			wantErr:     assert.NoError,
		},
		{
			name: "error getting metrics from sub-controls",
			fields: fields{
				catalogControls: map[string]map[string]*orchestrator.Control{
					orchestratortest.MockCatalogId1: {
						fmt.Sprintf("%s-%s", orchestratortest.MockControl1.GetCategoryName(), orchestratortest.MockControl1.GetId()): orchestratortest.MockControl1,
					},
				},
			},
			args: args{
				catalogId:    orchestratortest.MockCatalogId1,
				categoryName: orchestratortest.MockCategoryName1,
				controlId:    orchestratortest.MockControlId1,
			},
			wantMetrics: nil,
			wantErr: func(t *testing.T, err error, i ...interface{}) bool {
				return assert.ErrorContains(t, err, "error getting metrics from sub-controls")
			},
		},
		{
			name: "Happy path: control with metrics directly (no sub-controls)",
			fields: fields{
				catalogControls: map[string]map[string]*orchestrator.Control{
					orchestratortest.MockCatalogId2: {
						fmt.Sprintf("%s-%s", orchestratortest.MockControl2.GetCategoryName(), orchestratortest.MockControl2.GetId()): {
							Id:                orchestratortest.MockControlId2,
							CategoryName:      orchestratortest.MockCategoryName2,
							CategoryCatalogId: orchestratortest.MockCatalogId2,
							Name:              "Mock Control 2",
							Metrics: []*assessment.Metric{
								{
									Id:       orchestratortest.MockMetricId2,
									Version:  "2.0",
									Comments: "Direct metric on control",
								},
							},
						},
					},
				},
			},
			args: args{
				catalogId:    orchestratortest.MockCatalogId2,
				categoryName: orchestratortest.MockCategoryName2,
				controlId:    orchestratortest.MockControlId2,
			},
			wantMetrics: []*assessment.Metric{
				{
					Id:       orchestratortest.MockMetricId2,
					Version:  "2.0",
					Comments: "Direct metric on control",
				},
			},
			wantErr: assert.NoError,
		},
		{
			name: "Happy path: control with sub-controls",
			fields: fields{
				catalogControls: map[string]map[string]*orchestrator.Control{
					orchestratortest.MockControl1.GetCategoryCatalogId(): {
						fmt.Sprintf("%s-%s", orchestratortest.MockControl1.GetCategoryName(), orchestratortest.MockControl1.GetId()):       orchestratortest.MockControl1,
						fmt.Sprintf("%s-%s", orchestratortest.MockSubControl1.GetCategoryName(), orchestratortest.MockSubControl1.GetId()): orchestratortest.MockSubControl1,
					},
				},
			},
			args: args{
				catalogId:    orchestratortest.MockCatalogId1,
				categoryName: orchestratortest.MockCategoryName1,
				controlId:    orchestratortest.MockControlId1,
			},
			wantMetrics: []*assessment.Metric{
				{
					Id:       orchestratortest.MockMetricId1,
					Version:  "1.0",
					Comments: "This is a comment",
				},
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Service{
				catalogControls: tt.fields.catalogControls,
			}
			gotMetrics, err := s.getAllMetricsFromControl(tt.args.catalogId, tt.args.categoryName, tt.args.controlId)
			tt.wantErr(t, err)

			if assert.Equal(t, len(gotMetrics), len(tt.wantMetrics)) {
				for i := range gotMetrics {
					assert.Equal(t, tt.wantMetrics[i], gotMetrics[i])
				}
			}
		})
	}
}

func TestService_getMetricsFromSubControls(t *testing.T) {
	type fields struct {
		catalogControls map[string]map[string]*orchestrator.Control
	}
	type args struct {
		control *orchestrator.Control
	}
	tests := []struct {
		name        string
		fields      fields
		args        args
		wantMetrics []*assessment.Metric
		wantErr     assert.WantErr
	}{
		{
			name:        "Control is missing",
			args:        args{},
			wantMetrics: nil,
			wantErr: func(t *testing.T, err error, i ...interface{}) bool {
				return assert.ErrorContains(t, err, "control is missing")
			},
		},
		{
			name: "Sub-control level and metrics missing",
			args: args{
				control: &orchestrator.Control{
					Id:                orchestratortest.MockControlId1,
					CategoryName:      orchestratortest.MockCategoryName1,
					CategoryCatalogId: orchestratortest.MockCatalogId1,
					Name:              "Mock Control 1",
					Description:       "Mock control description",
				},
			},
			wantMetrics: nil,
			wantErr:     assert.NoError,
		},
		{
			name:   "Error getting control",
			fields: fields{},
			args: args{
				control: &orchestrator.Control{
					Id:                "testId",
					CategoryName:      orchestratortest.MockCategoryName1,
					CategoryCatalogId: orchestratortest.MockCatalogId1,
					Name:              "testId",
					Controls: []*orchestrator.Control{
						{
							Id:                "testId-subcontrol",
							CategoryName:      orchestratortest.MockCategoryName1,
							CategoryCatalogId: orchestratortest.MockCatalogId1,
							Name:              "testId-subcontrol",
						},
					},
					Metrics: []*assessment.Metric{},
				},
			},
			wantMetrics: nil,
			wantErr: func(t *testing.T, err error, i ...interface{}) bool {
				return assert.ErrorIs(t, err, service.ErrControlNotAvailable)
			},
		},
		{
			name: "Happy path",
			fields: fields{
				catalogControls: map[string]map[string]*orchestrator.Control{
					orchestratortest.MockControl1.GetCategoryCatalogId(): {
						fmt.Sprintf("%s-%s", orchestratortest.MockControl1.GetCategoryName(), orchestratortest.MockControl1.GetId()):       orchestratortest.MockControl1,
						fmt.Sprintf("%s-%s", orchestratortest.MockSubControl1.GetCategoryName(), orchestratortest.MockSubControl1.GetId()): orchestratortest.MockSubControl1,
					},
				},
			},
			args: args{
				control: orchestratortest.MockControl1,
			},
			wantMetrics: orchestratortest.MockControl1.Controls[0].GetMetrics(),
			wantErr:     assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Service{
				catalogControls: tt.fields.catalogControls,
			}
			gotMetrics, err := s.getMetricsFromSubcontrols(tt.args.control)

			tt.wantErr(t, err)

			assert.Equal(t, len(gotMetrics), len(tt.wantMetrics))
			for i := range gotMetrics {
				assert.Equal(t, tt.wantMetrics[i], gotMetrics[i])
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
		//{
		//	name: "happy path: filter by `get latest by control id`",
		//	args: args{
		//		req: connect.NewRequest(&evaluation.ListEvaluationResultsRequest{
		//			LatestByControlId: util.Ref(true),
		//		}),
		//	},
		//	fields: fields{
		//		db: persistencetest.NewInMemoryDB(t, types, []persistence.CustomJoinTable{}, func(d persistence.DB) {
		//			err := d.Create(evaluationtest.MockEvaluationResult1)
		//			assert.NoError(t, err)
		//			err = d.Create(evaluationtest.MockEvaluationResult2)
		//			assert.NoError(t, err)
		//			err = d.Create(evaluationtest.MockEvaluationResult3)
		//			assert.NoError(t, err)
		//			err = d.Create(evaluationtest.MockEvaluationResult4)
		//			assert.NoError(t, err)
		//		}),
		//	},
		//	want: func(t *testing.T, got *connect.Response[evaluation.ListEvaluationResultsResponse], msgAndArgs ...any) bool {
		//		assert.NotNil(t, got)
		//		assert.Equal(t, 1, len(got.Msg.Results))
		//		return assert.Equal(t, evaluationtest.MockEvaluationResult1, got.Msg.Results[0])
		//	},
		//	wantErr: assert.NoError,
		//},
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

func TestService_cacheControls(t *testing.T) {
	type fields struct {
		orchestratorClient orchestratorconnect.OrchestratorClient
		catalogControls    map[string]map[string]*orchestrator.Control
	}
	type args struct {
		catalogId string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantSvc assert.Want[*Service]
		wantErr assert.WantErr
	}{
		{
			name: "catalog_id missing",
			fields: fields{
				catalogControls: make(map[string]map[string]*orchestrator.Control),
			},
			args: args{
				catalogId: "",
			},
			wantErr: func(t *testing.T, err error, i ...interface{}) bool {
				return assert.ErrorContains(t, err, "catalog ID is missing")
			},
		},
		{
			name: "orchestrator client returns error",
			fields: func() fields {
				// Create test server that returns an error
				_, _, testSrv := newOrchestratorTestServerWithError(t, connect.NewError(connect.CodeInternal, fmt.Errorf("orchestrator service unavailable")))
				t.Cleanup(testSrv.Close)
				return fields{
					orchestratorClient: newOrchestratorClientForTest(testSrv),
					catalogControls:    make(map[string]map[string]*orchestrator.Control),
				}
			}(),
			args: args{
				catalogId: orchestratortest.MockCatalogId1,
			},
			wantErr: func(t *testing.T, err error, i ...interface{}) bool {
				return assert.ErrorContains(t, err, "orchestrator service unavailable")
			},
		},
		{
			name: "no controls available",
			fields: func() fields {
				// Create test server with empty control list
				_, _, testSrv := newOrchestratorTestServer(t, []*orchestrator.Control{})
				t.Cleanup(testSrv.Close)
				return fields{
					orchestratorClient: newOrchestratorClientForTest(testSrv),
					catalogControls:    make(map[string]map[string]*orchestrator.Control),
				}
			}(),
			args: args{
				catalogId: orchestratortest.MockCatalogId1,
			},
			wantErr: func(t *testing.T, err error, i ...interface{}) bool {
				return assert.ErrorContains(t, err, fmt.Sprintf("no controls for catalog '%s' available", orchestratortest.MockCatalogId1))
			},
		},
		{
			name: "Happy path",
			fields: func() fields {
				controls := mockControlsForCatalog(orchestratortest.MockCatalogId1)
				_, _, testSrv := newOrchestratorTestServer(t, controls)
				t.Cleanup(testSrv.Close)
				return fields{
					orchestratorClient: newOrchestratorClientForTest(testSrv),
					catalogControls:    make(map[string]map[string]*orchestrator.Control),
				}
			}(),
			args: args{
				catalogId: orchestratortest.MockCatalogId1,
			},
			wantSvc: func(t *testing.T, got *Service, msgAndArgs ...any) bool {
				assert.Equal(t, 1, len(got.catalogControls))
				return assert.Equal(t, 4, len(got.catalogControls[orchestratortest.MockCatalogId1]))
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &Service{
				orchestratorClient: tt.fields.orchestratorClient,
				catalogControls:    tt.fields.catalogControls,
			}
			err := svc.cacheControls(tt.args.catalogId)
			tt.wantErr(t, err)
			assert.Optional(t, tt.wantSvc, svc)
		})
	}
}

func TestService_getControl(t *testing.T) {
	type fields struct {
		catalogControls map[string]map[string]*orchestrator.Control
	}
	type args struct {
		catalogId    string
		categoryName string
		controlId    string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    assert.Want[*orchestrator.Control]
		wantErr assert.WantErr
	}{
		{
			name:   "catalog_id is missing",
			fields: fields{},
			args: args{
				categoryName: evidencetest.MockCategoryName,
				controlId:    evidencetest.MockControlID1,
			},
			want: nil,
			wantErr: func(t *testing.T, err error, _ ...any) bool {
				return assert.ErrorContains(t, err, "catalog id is missing")
			},
		},
		{
			name:   "category_name is missing",
			fields: fields{},
			args: args{
				catalogId: evidencetest.MockCatalogID1,
				controlId: evidencetest.MockControlID1,
			},
			want: nil,
			wantErr: func(t *testing.T, err error, i ...interface{}) bool {
				return assert.ErrorContains(t, err, "category name is missing")
			},
		},
		{
			name:   "control_id is missing",
			fields: fields{},
			args: args{
				catalogId:    evidencetest.MockCatalogID1,
				categoryName: evidencetest.MockCategoryName,
			},
			want: nil,
			wantErr: func(t *testing.T, err error, i ...interface{}) bool {
				return assert.ErrorContains(t, err, "control id is missing")
			},
		},
		{
			name:   "control does not exist",
			fields: fields{},
			args: args{
				catalogId:    "wrong_catalog_id",
				categoryName: "wrong_category_id",
				controlId:    "wrong_control_id",
			},
			want: nil,
			wantErr: func(t *testing.T, err error, i ...interface{}) bool {
				return assert.ErrorIs(t, err, service.ErrControlNotAvailable)
			},
		},
		{
			name: "Happy path",
			fields: fields{
				catalogControls: map[string]map[string]*orchestrator.Control{
					orchestratortest.MockControl1.GetCategoryCatalogId(): {
						fmt.Sprintf("%s-%s", orchestratortest.MockControl1.GetCategoryName(), orchestratortest.MockControl1.GetId()): orchestratortest.MockControl1,
						fmt.Sprintf("%s-%s", orchestratortest.MockControl1.GetCategoryName(), orchestratortest.MockControl2.GetId()): orchestratortest.MockControl2,
					},
				},
			},
			args: args{
				catalogId:    orchestratortest.MockControl1.GetCategoryCatalogId(),
				categoryName: orchestratortest.MockControl1.GetCategoryName(),
				controlId:    orchestratortest.MockControl1.GetId(),
			},
			want: func(t *testing.T, got *orchestrator.Control, _ ...any) bool {
				// We need to truncate the metric from the control because the control is only returned with its
				// sub-control but without the sub-control's metric.
				// TODO(oxisto): Use ignore fields instead
				wantControl := orchestratortest.MockControl1
				tmpMetrics := wantControl.Controls[0].Metrics
				wantControl.Controls[0].Metrics = nil

				if !assert.Equal(t, wantControl, got) {
					t.Errorf("Service.GetControl() = %v, want %v", got, wantControl)
					wantControl.Controls[0].Metrics = tmpMetrics
					return false
				}

				wantControl.Controls[0].Metrics = tmpMetrics
				return true
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Service{
				catalogControls: tt.fields.catalogControls,
			}

			gotControl, err := s.getControl(tt.args.catalogId, tt.args.categoryName, tt.args.controlId)
			tt.wantErr(t, err)

			if gotControl != nil {
				tt.want(t, gotControl)
			}
		})
	}
}

func Test_handlePending(t *testing.T) {
	type args struct {
		eval *evaluation.EvaluationResult
	}
	tests := []struct {
		name string
		args args
		want assert.Want[evaluation.EvaluationStatus]
	}{
		{
			name: "Status: Pending",
			args: args{
				eval: &evaluation.EvaluationResult{
					Status: evaluation.EvaluationStatus_EVALUATION_STATUS_PENDING,
				},
			},
			want: func(t *testing.T, got evaluation.EvaluationStatus, _ ...any) bool {
				return assert.Equal(t, evaluation.EvaluationStatus_EVALUATION_STATUS_PENDING, got)
			},
		},
		{
			name: "Status: Compliant",
			args: args{
				eval: &evaluation.EvaluationResult{
					Status: evaluation.EvaluationStatus_EVALUATION_STATUS_COMPLIANT,
				},
			},
			want: func(t *testing.T, got evaluation.EvaluationStatus, _ ...any) bool {
				return assert.Equal(t, evaluation.EvaluationStatus_EVALUATION_STATUS_COMPLIANT, got)
			},
		},
		{
			name: "Status: Compliant manually",
			args: args{
				eval: &evaluation.EvaluationResult{
					Status: evaluation.EvaluationStatus_EVALUATION_STATUS_COMPLIANT_MANUALLY,
				},
			},
			want: func(t *testing.T, got evaluation.EvaluationStatus, _ ...any) bool {
				return assert.Equal(t, evaluation.EvaluationStatus_EVALUATION_STATUS_COMPLIANT, got)
			},
		},
		{
			name: "Status: Not compliant manually without failing assessment results",
			args: args{
				eval: &evaluation.EvaluationResult{
					Status: evaluation.EvaluationStatus_EVALUATION_STATUS_NOT_COMPLIANT_MANUALLY,
				},
			},
			want: func(t *testing.T, got evaluation.EvaluationStatus, _ ...any) bool {
				return assert.Equal(t, evaluation.EvaluationStatus_EVALUATION_STATUS_NOT_COMPLIANT, got)
			},
		},
		{
			name: "Status: Not compliant with failing assessment results",
			args: args{
				eval: &evaluation.EvaluationResult{
					Status: evaluation.EvaluationStatus_EVALUATION_STATUS_NOT_COMPLIANT,
				},
			},
			want: func(t *testing.T, got evaluation.EvaluationStatus, _ ...any) bool {
				return assert.Equal(t, evaluation.EvaluationStatus_EVALUATION_STATUS_NOT_COMPLIANT, got)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := handlePending(tt.args.eval)
			tt.want(t, got)
		})
	}
}

func Test_handleCompliant(t *testing.T) {
	type args struct {
		er *evaluation.EvaluationResult
	}
	tests := []struct {
		name string
		args args
		want assert.Want[evaluation.EvaluationStatus]
	}{
		{
			name: "Status: Pending",
			args: args{
				er: &evaluation.EvaluationResult{
					Status: evaluation.EvaluationStatus_EVALUATION_STATUS_PENDING,
				},
			},
			want: func(t *testing.T, got evaluation.EvaluationStatus, msgAndArgs ...any) bool {
				return assert.Equal(t, evaluation.EvaluationStatus_EVALUATION_STATUS_COMPLIANT, got)
			},
		},
		{
			name: "Status: Compliant",
			args: args{
				er: &evaluation.EvaluationResult{
					Status: evaluation.EvaluationStatus_EVALUATION_STATUS_COMPLIANT,
				},
			},
			want: func(t *testing.T, got evaluation.EvaluationStatus, msgAndArgs ...any) bool {
				return assert.Equal(t, evaluation.EvaluationStatus_EVALUATION_STATUS_COMPLIANT, got)
			},
		},
		{
			name: "Status: Compliant manually",
			args: args{
				er: &evaluation.EvaluationResult{
					Status: evaluation.EvaluationStatus_EVALUATION_STATUS_COMPLIANT_MANUALLY,
				},
			},
			want: func(t *testing.T, got evaluation.EvaluationStatus, msgAndArgs ...any) bool {
				return assert.Equal(t, evaluation.EvaluationStatus_EVALUATION_STATUS_COMPLIANT, got)
			},
		},
		{
			name: "Status: Not compliant manually",
			args: args{
				er: &evaluation.EvaluationResult{
					Status:              evaluation.EvaluationStatus_EVALUATION_STATUS_NOT_COMPLIANT_MANUALLY,
					AssessmentResultIds: []string{"fail1", "fail2"},
				},
			},
			want: func(t *testing.T, got evaluation.EvaluationStatus, msgAndArgs ...any) bool {
				return assert.Equal(t, evaluation.EvaluationStatus_EVALUATION_STATUS_NOT_COMPLIANT, got)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := handleCompliant(tt.args.er)
			tt.want(t, got)
		})
	}
}

func Test_getMetricIds(t *testing.T) {
	type args struct {
		metrics []*assessment.Metric
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "Empty input",
			args: args{},
			want: nil,
		},
		{
			name: "Happy path",
			args: args{
				metrics: []*assessment.Metric{
					{
						Id: evidencetest.MockSubControlID11,
					},
					{
						Id: evidencetest.MockSubControlID,
					},
				},
			},
			want: []string{evidencetest.MockSubControlID11, evidencetest.MockSubControlID},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getMetricIds(tt.args.metrics)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_values(t *testing.T) {
	var (
		control1 = orchestratortest.MockControl1
		control2 = orchestratortest.MockControl2
	)

	// Ensure deterministic comparisons regardless of map iteration order.
	sortControls := func(controls []*orchestrator.Control) {
		slices.SortFunc(controls, func(a *orchestrator.Control, b *orchestrator.Control) int {
			return strings.Compare(a.Id, b.Id)
		})
	}

	type args struct {
		m map[string]*orchestrator.Control
	}

	tests := []struct {
		name string
		args args
		want assert.Want[[]*orchestrator.Control]
	}{
		{
			name: "empty map",
			args: args{
				m: map[string]*orchestrator.Control{},
			},
			want: func(t *testing.T, got []*orchestrator.Control, msgAndArgs ...any) bool {
				sortControls(got)
				return assert.Equal(t, []*orchestrator.Control{}, got)
			},
		},
		{
			name: "single control",
			args: args{
				m: map[string]*orchestrator.Control{
					control1.Id: control1,
				},
			},
			want: func(t *testing.T, got []*orchestrator.Control, msgAndArgs ...any) bool {
				sortControls(got)
				return assert.Equal(t, []*orchestrator.Control{control1}, got)
			},
		},
		{
			name: "multiple controls",
			args: args{
				m: map[string]*orchestrator.Control{
					control1.Id: control1,
					control2.Id: control2,
				},
			},
			want: func(t *testing.T, got []*orchestrator.Control, msgAndArgs ...any) bool {
				sortControls(got)
				return assert.Equal(t, []*orchestrator.Control{control1, control2}, got)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := values(tt.args.m)

			tt.want(t, got)
		})
	}
}
