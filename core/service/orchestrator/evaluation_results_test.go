package orchestrator

import (
	"context"
	"slices"
	"testing"

	"confirmate.io/core/api/evaluation"
	"confirmate.io/core/api/orchestrator"
	"confirmate.io/core/persistence"
	"confirmate.io/core/persistence/persistencetest"
	"confirmate.io/core/service/evaluation/evaluationtest"
	"confirmate.io/core/util/assert"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestService_StoreEvaluationResult(t *testing.T) {
	type args struct {
		req *connect.Request[orchestrator.StoreEvaluationResultRequest]
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
			name: "error: empty request",
			args: args{
				req: &connect.Request[orchestrator.StoreEvaluationResultRequest]{},
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
			name: "error: database error on create",
			args: args{
				req: connect.NewRequest(&orchestrator.StoreEvaluationResultRequest{
					Result: evaluationtest.MockManualEvaluationResult1,
				}),
			},
			fields: fields{
				db: persistencetest.CreateErrorDB(t, persistence.ErrUniqueConstraintFailed, types, []persistence.CustomJoinTable{}),
			},
			want: assert.Nil[*connect.Response[evaluation.EvaluationResult]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				assert.IsConnectError(t, err, connect.CodeAlreadyExists)
				return assert.ErrorContains(t, err, "resource already exists")
			},
		},
		{
			name: "happy path: with all fields populated",
			args: args{
				req: connect.NewRequest(&orchestrator.StoreEvaluationResultRequest{
					Result: evaluationtest.MockManualEvaluationResult1,
				}),
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, []persistence.CustomJoinTable{}),
			},
			want: func(t *testing.T, got *connect.Response[evaluation.EvaluationResult], msgAndArgs ...any) bool {
				assert.NotEmpty(t, got.Msg.GetId())
				assert.NotEmpty(t, got.Msg.GetTimestamp())
				return assert.Equal(t, evaluationtest.MockManualEvaluationResult1, got.Msg, protocmp.IgnoreFields(&evaluation.EvaluationResult{}, "id", "timestamp"))
			},
			wantErr: assert.NoError,
		},
		{
			name: "happy path: ID missing",
			args: args{
				req: connect.NewRequest(&orchestrator.StoreEvaluationResultRequest{
					Result: evaluationtest.MockManualEvaluationResult2,
				}),
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, []persistence.CustomJoinTable{}),
			},
			want: func(t *testing.T, got *connect.Response[evaluation.EvaluationResult], msgAndArgs ...any) bool {
				assert.Equal(t, evaluationtest.MockToeId1, got.Msg.TargetOfEvaluationId)
				assert.Equal(t, evaluationtest.MockAuditScopeId1, got.Msg.AuditScopeId)
				assert.Equal(t, evaluationtest.MockSubcontrolId11, got.Msg.ControlId)
				assert.Equal(t, evaluationtest.MockCategoryName1, got.Msg.ControlCategoryName)
				assert.Equal(t, evaluationtest.MockCatalogId1, got.Msg.ControlCatalogId)
				assert.Equal(t, evaluation.EvaluationStatus_EVALUATION_STATUS_COMPLIANT_MANUALLY, got.Msg.Status)
				assert.Equal(t, "Mock manual evaluation result 1", got.Msg.GetComment())
				assert.Equal(t, 2, len(got.Msg.AssessmentResultIds))
				return assert.NotEmpty(t, got.Msg.Data)
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &Service{
				db: tt.fields.db,
			}
			got, gotErr := svc.StoreEvaluationResult(context.Background(), tt.args.req)

			tt.want(t, got)
			tt.wantErr(t, gotErr)
		})
	}
}

// TestService_StoreEvaluationResult_LargeBlobIntegration is an integration test that verifies large binary data blobs
// are correctly stored and retrieved from the database. This test ensures the persistence layer (GORM) properly handles
// large binary data (1MB+) in the evaluation result's Data field, which is used for giving additional justification.
func TestService_StoreEvaluationResult_LargeBlobIntegration(t *testing.T) {
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
	req := connect.NewRequest(&orchestrator.StoreEvaluationResultRequest{
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
			Comment:              new("Integration test for large blob storage"),
			Data:                 largeBlob,
		},
	})

	// Call StoreEvaluationResult to store the result with large blob
	res, err := svc.StoreEvaluationResult(context.Background(), req)

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

func TestService_ListEvaluationResults(t *testing.T) {
	type args struct {
		req *connect.Request[orchestrator.ListEvaluationResultsRequest]
	}
	type fields struct {
		db persistence.DB
	}
	tests := []struct {
		name    string
		args    args
		fields  fields
		want    assert.Want[*connect.Response[orchestrator.ListEvaluationResultsResponse]]
		wantErr assert.WantErr
	}{
		{
			name: "error: pagination error",
			args: args{
				req: &connect.Request[orchestrator.ListEvaluationResultsRequest]{
					Msg: &orchestrator.ListEvaluationResultsRequest{
						PageToken: "!!!invalid-base64!!!",
					},
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, []persistence.CustomJoinTable{}),
			},
			want: assert.Nil[*connect.Response[orchestrator.ListEvaluationResultsResponse]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				assert.IsConnectError(t, err, connect.CodeInvalidArgument)
				return assert.ErrorContains(t, err, "invalid page_token:")
			},
		},
		{
			name:   "error: validation error",
			args:   args{},
			fields: fields{},
			want:   assert.Nil[*connect.Response[orchestrator.ListEvaluationResultsResponse]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeInvalidArgument) &&
					assert.ErrorContains(t, err, "empty request")
			},
		},
		{
			name: "error: database error",
			args: args{
				req: connect.NewRequest(&orchestrator.ListEvaluationResultsRequest{
					LatestByControlId: new(true),
				}),
			},
			fields: fields{
				db: persistencetest.RawErrorDB(t, persistence.ErrDatabase, types, []persistence.CustomJoinTable{}),
			},
			want: assert.Nil[*connect.Response[orchestrator.ListEvaluationResultsResponse]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.ErrorContains(t, err, persistence.ErrDatabase.Error())
			},
		},
		{
			name: "happy path: filter by `get latest by control id` and filter by ToE",
			args: args{
				req: connect.NewRequest(&orchestrator.ListEvaluationResultsRequest{
					LatestByControlId: new(true),
					Filter: &orchestrator.ListEvaluationResultsRequest_Filter{
						TargetOfEvaluationId: new(evaluationtest.MockToeId1),
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
					err = d.Create(evaluationtest.MockManualEvaluationResult1)
					assert.NoError(t, err)
				}),
			},
			want: func(t *testing.T, got *connect.Response[orchestrator.ListEvaluationResultsResponse], msgAndArgs ...any) bool {
				assert.NotNil(t, got)
				assert.Equal(t, 2, len(got.Msg.Results))
				// MockManualEvaluationResult1 is the latest one targeting Control 1 (and fulfilling the filters)
				mockResult1IsIn := slices.ContainsFunc(got.Msg.Results, func(result *evaluation.EvaluationResult) bool {
					isSame := result.Id == evaluationtest.MockManualEvaluationResult1.Id
					return isSame
				})
				assert.True(t, mockResult1IsIn)

				// MockEvaluationResult3 is the only (=latest) one targeting Control 1.1 (and fulfilling the filters)
				result3IsIn := slices.ContainsFunc(got.Msg.Results, func(result *evaluation.EvaluationResult) bool {
					isSame := result.Id == evaluationtest.MockEvaluationResult3.Id
					return isSame
				})
				return assert.True(t, result3IsIn)
			},
			wantErr: assert.NoError,
		},
		{
			name: "happy path: filter by `valid manual only`",
			args: args{
				req: connect.NewRequest(&orchestrator.ListEvaluationResultsRequest{
					Filter: &orchestrator.ListEvaluationResultsRequest_Filter{
						ValidManualOnly: new(true),
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
			want: func(t *testing.T, got *connect.Response[orchestrator.ListEvaluationResultsResponse], msgAndArgs ...any) bool {
				assert.NotNil(t, got)
				assert.Equal(t, 1, len(got.Msg.Results))
				return assert.Equal(t, evaluationtest.MockEvaluationResult4, got.Msg.Results[0])
			},
			wantErr: assert.NoError,
		},
		{
			name: "happy path: filter by `parents only`",
			args: args{
				req: connect.NewRequest(&orchestrator.ListEvaluationResultsRequest{
					Filter: &orchestrator.ListEvaluationResultsRequest_Filter{
						ParentsOnly: new(true),
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
			want: func(t *testing.T, got *connect.Response[orchestrator.ListEvaluationResultsResponse], msgAndArgs ...any) bool {
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
				req: connect.NewRequest(&orchestrator.ListEvaluationResultsRequest{
					Filter: &orchestrator.ListEvaluationResultsRequest_Filter{
						SubControls: new(string(evaluationtest.MockSubcontrolId11)),
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
			want: func(t *testing.T, got *connect.Response[orchestrator.ListEvaluationResultsResponse], msgAndArgs ...any) bool {
				assert.NotNil(t, got)
				assert.Equal(t, 1, len(got.Msg.Results))
				return assert.Equal(t, evaluationtest.MockEvaluationResult3, got.Msg.Results[0])
			},
			wantErr: assert.NoError,
		},
		{
			name: "happy path: filter by control ID",
			args: args{
				req: connect.NewRequest(&orchestrator.ListEvaluationResultsRequest{
					Filter: &orchestrator.ListEvaluationResultsRequest_Filter{
						ControlId: new(string(evaluationtest.MockControlId2)),
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
			want: func(t *testing.T, got *connect.Response[orchestrator.ListEvaluationResultsResponse], msgAndArgs ...any) bool {
				assert.NotNil(t, got)
				assert.Equal(t, 1, len(got.Msg.Results))
				return assert.Equal(t, evaluationtest.MockEvaluationResult2, got.Msg.Results[0])
			},
			wantErr: assert.NoError,
		},
		{
			name: "happy path: filter by catalog ID",
			args: args{
				req: connect.NewRequest(&orchestrator.ListEvaluationResultsRequest{
					Filter: &orchestrator.ListEvaluationResultsRequest_Filter{
						CatalogId: new(string(evaluationtest.MockCatalogId2)),
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
			want: func(t *testing.T, got *connect.Response[orchestrator.ListEvaluationResultsResponse], msgAndArgs ...any) bool {
				assert.NotNil(t, got)
				assert.Equal(t, 1, len(got.Msg.Results))
				return assert.Equal(t, evaluationtest.MockEvaluationResult2, got.Msg.Results[0])
			},
			wantErr: assert.NoError,
		},
		{
			name: "happy path: filter by ToE",
			args: args{
				req: connect.NewRequest(&orchestrator.ListEvaluationResultsRequest{
					Filter: &orchestrator.ListEvaluationResultsRequest_Filter{
						TargetOfEvaluationId: new(evaluationtest.MockToeId2),
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
			want: func(t *testing.T, got *connect.Response[orchestrator.ListEvaluationResultsResponse], msgAndArgs ...any) bool {
				assert.NotNil(t, got)
				assert.Equal(t, 1, len(got.Msg.Results))
				return assert.Equal(t, evaluationtest.MockEvaluationResult2, got.Msg.Results[0])
			},
			wantErr: assert.NoError,
		},
		{
			name: "happy path",
			args: args{
				req: connect.NewRequest(&orchestrator.ListEvaluationResultsRequest{}),
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, []persistence.CustomJoinTable{}, func(d persistence.DB) {
					err := d.Create(evaluationtest.MockEvaluationResult1)
					assert.NoError(t, err)
					err = d.Create(evaluationtest.MockEvaluationResult2)
					assert.NoError(t, err)
				}),
			},
			want: func(t *testing.T, got *connect.Response[orchestrator.ListEvaluationResultsResponse], msgAndArgs ...any) bool {
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
