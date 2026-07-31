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

package orchestrator

import (
	"context"
	"testing"
	"time"

	"confirmate.io/core/api/evaluation"
	"confirmate.io/core/api/orchestrator"
	"confirmate.io/core/persistence"
	"confirmate.io/core/persistence/persistencetest"
	"confirmate.io/core/service"
	"confirmate.io/core/service/evaluation/evaluationtest"
	"confirmate.io/core/service/orchestrator/orchestratortest"
	"confirmate.io/core/util/assert"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestService_updateCertificateLifecycle(t *testing.T) {
	type fields struct {
		db persistence.DB
	}
	tests := []struct {
		name         string
		fields       fields
		auditScopeId string
		wantErr      assert.WantErr
		wantDB       assert.Want[persistence.DB]
	}{
		{
			name:         "no-op when no certificate is linked to the audit scope",
			auditScopeId: orchestratortest.MockScopeId1,
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables),
			},
			wantErr: assert.NoError,
			wantDB: func(t *testing.T, db persistence.DB, msgAndArgs ...any) bool {
				auditScopeId := msgAndArgs[0].(string)

				// Confirm the actual reason for the no-op: no certificate is linked to this audit scope.
				var cert orchestrator.Certificate
				err := db.Get(&cert, "audit_scope_id = ?", auditScopeId)
				assert.ErrorIs(t, err, persistence.ErrRecordNotFound)

				var states []*orchestrator.State
				assert.NoError(t, db.List(&states, "id", true, 0, -1))
				return assert.Equal(t, 0, len(states))
			},
		},
		{
			name:         "no-op when there are no evaluation results yet",
			auditScopeId: orchestratortest.MockScopeId1,
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					assert.NoError(t, d.Create(orchestratortest.MockCertificate1))
				}),
			},
			wantErr: assert.NoError,
			wantDB: func(t *testing.T, db persistence.DB, msgAndArgs ...any) bool {
				auditScopeId := msgAndArgs[0].(string)

				// Confirm a certificate does exist (ruling out the "no certificate" case above)
				// but there are genuinely no evaluation results yet for this scope.
				var cert orchestrator.Certificate
				assert.NoError(t, db.Get(&cert, "audit_scope_id = ?", auditScopeId))

				var results []*evaluation.EvaluationResult
				assert.NoError(t, db.List(&results, "id", true, 0, -1, "audit_scope_id = ?", auditScopeId))
				assert.Equal(t, 0, len(results))

				var states []*orchestrator.State
				assert.NoError(t, db.List(&states, "id", true, 0, -1))
				return assert.Equal(t, 0, len(states))
			},
		},
		{
			name:         "no-op when all results are still PENDING",
			auditScopeId: orchestratortest.MockScopeId1,
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					assert.NoError(t, d.Create(orchestratortest.MockCertificate1))
					assert.NoError(t, d.Create(&evaluation.EvaluationResult{
						Id:           "00000000-0000-0000-0099-000000000001",
						AuditScopeId: orchestratortest.MockScopeId1,
						ControlId:    evaluationtest.MockControlId1,
						Status:       evaluation.EvaluationStatus_EVALUATION_STATUS_PENDING,
					}))
				}),
			},
			wantErr: assert.NoError,
			wantDB: func(t *testing.T, db persistence.DB, msgAndArgs ...any) bool {
				auditScopeId := msgAndArgs[0].(string)

				// Confirm results exist for this scope, and all of them are PENDING —
				// ruling out the "no results" case above and pinning down the actual reason.
				var results []*evaluation.EvaluationResult
				assert.NoError(t, db.List(&results, "id", true, 0, -1, "audit_scope_id = ?", auditScopeId))
				assert.NotEqual(t, 0, len(results))
				for _, r := range results {
					assert.Equal(t, evaluation.EvaluationStatus_EVALUATION_STATUS_PENDING, r.GetStatus())
				}

				var states []*orchestrator.State
				assert.NoError(t, db.List(&states, "id", true, 0, -1))
				return assert.Equal(t, 0, len(states))
			},
		},
		{
			name:         "appends 'suspended' when a NOT_COMPLIANT result exists",
			auditScopeId: orchestratortest.MockScopeId1,
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					assert.NoError(t, d.Create(orchestratortest.MockCertificate1))
					assert.NoError(t, d.Create(&evaluation.EvaluationResult{
						Id:           "00000000-0000-0000-0099-000000000001",
						AuditScopeId: orchestratortest.MockScopeId1,
						ControlId:    evaluationtest.MockControlId1,
						Status:       evaluation.EvaluationStatus_EVALUATION_STATUS_NOT_COMPLIANT,
						Timestamp:    timestamppb.Now(),
					}))
				}),
			},
			wantErr: assert.NoError,
			wantDB: func(t *testing.T, db persistence.DB, _ ...any) bool {
				var states []*orchestrator.State
				assert.NoError(t, db.List(&states, "id", true, 0, -1))
				if !assert.Equal(t, 1, len(states)) {
					return false
				}

				// Id and Timestamp are generated at runtime; check them for well-formedness
				// and compare the rest of the object as a whole.
				_, err := uuid.Parse(states[0].GetId())
				assert.NoError(t, err)
				_, err = time.Parse(time.RFC3339, states[0].GetTimestamp())
				assert.NoError(t, err)

				return assert.Equal(t, &orchestrator.State{
					State:         CertificateStateSuspended,
					CertificateId: orchestratortest.MockCertificateId1,
				}, states[0], protocmp.IgnoreFields(&orchestrator.State{}, "id", "timestamp"))
			},
		},
		{
			name:         "appends 'new' when all results are COMPLIANT",
			auditScopeId: orchestratortest.MockScopeId1,
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					assert.NoError(t, d.Create(orchestratortest.MockCertificate1))
					assert.NoError(t, d.Create(&evaluation.EvaluationResult{
						Id:           "00000000-0000-0000-0099-000000000001",
						AuditScopeId: orchestratortest.MockScopeId1,
						ControlId:    evaluationtest.MockControlId1,
						Status:       evaluation.EvaluationStatus_EVALUATION_STATUS_COMPLIANT,
						Timestamp:    timestamppb.Now(),
					}))
				}),
			},
			wantErr: assert.NoError,
			wantDB: func(t *testing.T, db persistence.DB, _ ...any) bool {
				var states []*orchestrator.State
				assert.NoError(t, db.List(&states, "id", true, 0, -1))
				return assert.Equal(t, 1, len(states)) &&
					assert.Equal(t, CertificateStateNew, states[0].State)
			},
		},
		{
			name:         "no duplicate state when the compliance posture has not changed",
			auditScopeId: orchestratortest.MockScopeId1,
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					cert := &orchestrator.Certificate{
						Id:                   orchestratortest.MockCertificateId1,
						Name:                 orchestratortest.MockCertifiateName1,
						TargetOfEvaluationId: orchestratortest.MockToeId1,
						AuditScopeId:         orchestratortest.MockScopeId1,
						States: []*orchestrator.State{
							{Id: "00000000-0000-0000-0088-000000000001", State: CertificateStateNew, Timestamp: "2026-01-01T00:00:00Z", CertificateId: orchestratortest.MockCertificateId1},
						},
					}
					assert.NoError(t, d.Create(cert))
					assert.NoError(t, d.Create(&evaluation.EvaluationResult{
						Id:           "00000000-0000-0000-0099-000000000001",
						AuditScopeId: orchestratortest.MockScopeId1,
						ControlId:    evaluationtest.MockControlId1,
						Status:       evaluation.EvaluationStatus_EVALUATION_STATUS_COMPLIANT,
						Timestamp:    timestamppb.Now(),
					}))
				}),
			},
			wantErr: assert.NoError,
			wantDB: func(t *testing.T, db persistence.DB, _ ...any) bool {
				var states []*orchestrator.State
				assert.NoError(t, db.List(&states, "id", true, 0, -1))
				return assert.Equal(t, 1, len(states))
			},
		},
		{
			name:         "does not override a 'withdrawn' state",
			auditScopeId: orchestratortest.MockScopeId1,
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					cert := &orchestrator.Certificate{
						Id:                   orchestratortest.MockCertificateId1,
						Name:                 orchestratortest.MockCertifiateName1,
						TargetOfEvaluationId: orchestratortest.MockToeId1,
						AuditScopeId:         orchestratortest.MockScopeId1,
						States: []*orchestrator.State{
							{Id: "00000000-0000-0000-0088-000000000001", State: CertificateStateWithdrawn, Timestamp: "2026-01-01T00:00:00Z", CertificateId: orchestratortest.MockCertificateId1},
						},
					}
					assert.NoError(t, d.Create(cert))
					assert.NoError(t, d.Create(&evaluation.EvaluationResult{
						Id:           "00000000-0000-0000-0099-000000000001",
						AuditScopeId: orchestratortest.MockScopeId1,
						ControlId:    evaluationtest.MockControlId1,
						Status:       evaluation.EvaluationStatus_EVALUATION_STATUS_NOT_COMPLIANT,
						Timestamp:    timestamppb.Now(),
					}))
				}),
			},
			wantErr: assert.NoError,
			wantDB: func(t *testing.T, db persistence.DB, _ ...any) bool {
				var states []*orchestrator.State
				assert.NoError(t, db.List(&states, "id", true, 0, -1))
				return assert.Equal(t, 1, len(states)) &&
					assert.Equal(t, CertificateStateWithdrawn, states[0].State)
			},
		},
		{
			name:         "sub-control results are ignored; only parent-level results drive the state",
			auditScopeId: orchestratortest.MockScopeId1,
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					assert.NoError(t, d.Create(orchestratortest.MockCertificate1))
					// Parent is compliant.
					assert.NoError(t, d.Create(&evaluation.EvaluationResult{
						Id:           "00000000-0000-0000-0099-000000000001",
						AuditScopeId: orchestratortest.MockScopeId1,
						ControlId:    evaluationtest.MockControlId1,
						Status:       evaluation.EvaluationStatus_EVALUATION_STATUS_COMPLIANT,
						Timestamp:    timestamppb.Now(),
					}))
					// Sub-control is not compliant — must not affect the certificate state.
					parent := evaluationtest.MockControlId1
					assert.NoError(t, d.Create(&evaluation.EvaluationResult{
						Id:              "00000000-0000-0000-0099-000000000002",
						AuditScopeId:    orchestratortest.MockScopeId1,
						ControlId:       evaluationtest.MockControl1SubcontrolId11,
						ParentControlId: &parent,
						Status:          evaluation.EvaluationStatus_EVALUATION_STATUS_NOT_COMPLIANT,
						Timestamp:       timestamppb.Now(),
					}))
				}),
			},
			wantErr: assert.NoError,
			wantDB: func(t *testing.T, db persistence.DB, _ ...any) bool {
				var states []*orchestrator.State
				assert.NoError(t, db.List(&states, "id", true, 0, -1))
				return assert.Equal(t, 1, len(states)) &&
					assert.Equal(t, CertificateStateNew, states[0].State)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &Service{db: tt.fields.db}
			err := svc.updateCertificateLifecycle(context.Background(), tt.auditScopeId)
			tt.wantErr(t, err)
			tt.wantDB(t, tt.fields.db, tt.auditScopeId)
		})
	}
}

func TestService_UpdateCertificateLifecycle(t *testing.T) {
	type fields struct {
		db    persistence.DB
		authz service.AuthorizationStrategy
	}
	tests := []struct {
		name    string
		fields  fields
		req     *orchestrator.UpdateCertificateLifecycleRequest
		wantErr assert.WantErr
		wantDB  assert.Want[persistence.DB]
	}{
		{
			name: "validation error",
			fields: fields{
				db:    persistencetest.NewInMemoryDB(t, types, joinTables),
				authz: &service.AuthorizationStrategyAllowAll{},
			},
			req: &orchestrator.UpdateCertificateLifecycleRequest{},
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeInvalidArgument)
			},
			wantDB: func(t *testing.T, db persistence.DB, _ ...any) bool {
				var states []*orchestrator.State
				assert.NoError(t, db.List(&states, "id", true, 0, -1))
				return assert.Equal(t, 0, len(states))
			},
		},
		{
			name: "permission denied",
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					assert.NoError(t, d.Create(orchestratortest.MockCertificate1))
					assert.NoError(t, d.Create(&evaluation.EvaluationResult{
						Id:           "00000000-0000-0000-0099-000000000001",
						AuditScopeId: orchestratortest.MockScopeId1,
						ControlId:    evaluationtest.MockControlId1,
						Status:       evaluation.EvaluationStatus_EVALUATION_STATUS_NOT_COMPLIANT,
						Timestamp:    timestamppb.Now(),
					}))
				}),
				authz: &denyAuthorizationStrategy{},
			},
			req: &orchestrator.UpdateCertificateLifecycleRequest{
				AuditScopeId: orchestratortest.MockScopeId1,
			},
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.ErrorIs(t, err, service.ErrPermissionDenied)
			},
			wantDB: func(t *testing.T, db persistence.DB, _ ...any) bool {
				var states []*orchestrator.State
				assert.NoError(t, db.List(&states, "id", true, 0, -1))
				return assert.Equal(t, 0, len(states))
			},
		},
		{
			name: "happy path: appends 'suspended' when a NOT_COMPLIANT result exists",
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					assert.NoError(t, d.Create(orchestratortest.MockCertificate1))
					assert.NoError(t, d.Create(&evaluation.EvaluationResult{
						Id:           "00000000-0000-0000-0099-000000000001",
						AuditScopeId: orchestratortest.MockScopeId1,
						ControlId:    evaluationtest.MockControlId1,
						Status:       evaluation.EvaluationStatus_EVALUATION_STATUS_NOT_COMPLIANT,
						Timestamp:    timestamppb.Now(),
					}))
				}),
				authz: &service.AuthorizationStrategyAllowAll{},
			},
			req: &orchestrator.UpdateCertificateLifecycleRequest{
				AuditScopeId: orchestratortest.MockScopeId1,
			},
			wantErr: assert.NoError,
			wantDB: func(t *testing.T, db persistence.DB, _ ...any) bool {
				var states []*orchestrator.State
				assert.NoError(t, db.List(&states, "id", true, 0, -1))
				return assert.Equal(t, 1, len(states)) &&
					assert.Equal(t, CertificateStateSuspended, states[0].State)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &Service{db: tt.fields.db, authz: tt.fields.authz}
			_, err := svc.UpdateCertificateLifecycle(context.Background(), connect.NewRequest(tt.req))
			tt.wantErr(t, err)
			tt.wantDB(t, tt.fields.db)
		})
	}
}
