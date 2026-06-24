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
	"testing"

	"confirmate.io/core/api/evaluation"
	"confirmate.io/core/api/orchestrator"
	"confirmate.io/core/persistence"
	"confirmate.io/core/persistence/persistencetest"
	"confirmate.io/core/service/evaluation/evaluationtest"
	"confirmate.io/core/service/orchestrator/orchestratortest"
	"confirmate.io/core/util/assert"

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
			wantDB: func(t *testing.T, db persistence.DB, _ ...any) bool {
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
			wantDB: func(t *testing.T, db persistence.DB, _ ...any) bool {
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
			wantDB: func(t *testing.T, db persistence.DB, _ ...any) bool {
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
				return assert.Equal(t, 1, len(states)) &&
					assert.Equal(t, CertificateStateSuspended, states[0].State) &&
					assert.Equal(t, orchestratortest.MockCertificateId1, states[0].CertificateId)
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
						ControlId:       evaluationtest.MockSubcontrolId11,
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
			err := svc.updateCertificateLifecycle(tt.auditScopeId)
			tt.wantErr(t, err)
			tt.wantDB(t, tt.fields.db)
		})
	}
}
