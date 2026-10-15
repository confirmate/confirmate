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

package orchestrator

import (
	"context"
	"testing"

	"confirmate.io/core/api/orchestrator"
	"confirmate.io/core/persistence"
	"confirmate.io/core/persistence/persistencetest"
	"confirmate.io/core/service"
	"confirmate.io/core/service/orchestrator/orchestratortest"
	"confirmate.io/core/util/assert"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	mockRiskAssessmentId1  = "00000000-0000-0000-0007-000000000001"
	mockRiskAssessmentId2  = "00000000-0000-0000-0007-000000000002"
	mockRiskJustification1 = "The product has no network interface."
)

// seedRiskAssessmentScope seeds the audit scope MockAuditScope1 with three controls of its catalog, a parent control
// with one sub-control and a second control, that are all in scope.
func seedRiskAssessmentScope(t *testing.T, d persistence.DB) {
	t.Helper()

	assert.NoError(t, d.Create(orchestratortest.MockAuditScope1))

	controls := []*orchestrator.Control{
		{Id: orchestratortest.MockControlId1, CatalogId: orchestratortest.MockCatalogId1, Name: orchestratortest.MockControlName1, ShortName: orchestratortest.MockControlShortName1},
		{Id: orchestratortest.MockControl2SubControlId1, CatalogId: orchestratortest.MockCatalogId1, Name: orchestratortest.MockSubControlName1, ShortName: orchestratortest.MockSubControlShortName1, ParentControlId: new(orchestratortest.MockControlId1)},
		{Id: orchestratortest.MockControlId2, CatalogId: orchestratortest.MockCatalogId1, Name: orchestratortest.MockControlName2, ShortName: orchestratortest.MockControlShortName2},
	}
	for _, c := range controls {
		assert.NoError(t, d.Create(c))
		assert.NoError(t, d.Create(&orchestrator.ControlInScope{
			Id:                   uuid.NewString(),
			AuditScopeId:         orchestratortest.MockScopeId1,
			TargetOfEvaluationId: orchestratortest.MockToeId1,
			ControlId:            c.Id,
			State:                orchestrator.ControlInScopeState_CONTROL_IN_SCOPE_STATE_OPEN,
		}))
	}
}

// mockRiskAssessment returns an assessment of MockAuditScope1 with the given ID, version, status and decisions.
func mockRiskAssessment(id string, version int32, status orchestrator.RiskAssessmentStatus, decisions ...*orchestrator.ScopeDecision) *orchestrator.RiskAssessment {
	return &orchestrator.RiskAssessment{
		Id:                   id,
		TargetOfEvaluationId: orchestratortest.MockToeId1,
		AuditScopeId:         orchestratortest.MockScopeId1,
		Version:              version,
		Status:               status,
		Method:               orchestrator.RiskAssessmentMethod_RISK_ASSESSMENT_METHOD_BUILT_IN,
		Decisions:            decisions,
		CreatedAt:            timestamppb.Now(),
	}
}

// outOfScope is a decision that takes a control out of scope.
func outOfScope(controlId, justification string) *orchestrator.ScopeDecision {
	return &orchestrator.ScopeDecision{
		ControlId:     controlId,
		InScope:       false,
		Justification: justification,
		Source:        orchestrator.ScopeDecisionSource_SCOPE_DECISION_SOURCE_RULE,
	}
}

// countControlInScope counts the controls in scope of MockAuditScope1 with the given control ID.
func countControlInScope(t *testing.T, db persistence.DB, controlId string) int64 {
	t.Helper()

	count, err := db.Count(&orchestrator.ControlInScope{}, "audit_scope_id = ? AND control_id = ?", orchestratortest.MockScopeId1, controlId)
	assert.NoError(t, err)

	return count
}

func TestService_CreateRiskAssessment(t *testing.T) {
	type args struct {
		req     *orchestrator.CreateRiskAssessmentRequest
		context context.Context
	}
	type fields struct {
		db    persistence.DB
		authz service.AuthorizationStrategy
	}
	valid := func(decisions ...*orchestrator.ScopeDecision) *orchestrator.CreateRiskAssessmentRequest {
		return &orchestrator.CreateRiskAssessmentRequest{
			RiskAssessment: &orchestrator.RiskAssessment{
				TargetOfEvaluationId: orchestratortest.MockToeId1,
				AuditScopeId:         orchestratortest.MockScopeId1,
				Method:               orchestrator.RiskAssessmentMethod_RISK_ASSESSMENT_METHOD_BUILT_IN,
				Decisions:            decisions,
			},
		}
	}
	tests := []struct {
		name    string
		args    args
		fields  fields
		want    assert.Want[*connect.Response[orchestrator.RiskAssessment]]
		wantErr assert.WantErr
		wantDB  assert.Want[persistence.DB]
	}{
		{
			name: "validation error - empty request",
			args: args{req: &orchestrator.CreateRiskAssessmentRequest{}},
			fields: fields{
				db:    persistencetest.NewInMemoryDB(t, types, joinTables),
				authz: &service.AuthorizationStrategyAllowAll{},
			},
			want: assert.Nil[*connect.Response[orchestrator.RiskAssessment]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeInvalidArgument)
			},
			wantDB: assert.NotNil[persistence.DB],
		},
		{
			name: "validation error - out of scope without a justification",
			args: args{req: valid(outOfScope(orchestratortest.MockControlId2, ""))},
			fields: fields{
				db:    persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) { seedRiskAssessmentScope(t, d) }),
				authz: &service.AuthorizationStrategyAllowAll{},
			},
			want: assert.Nil[*connect.Response[orchestrator.RiskAssessment]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeInvalidArgument) &&
					assert.ErrorContains(t, err, "needs a justification")
			},
			wantDB: assert.NotNil[persistence.DB],
		},
		{
			name: "audit scope not found",
			args: args{req: valid()},
			fields: fields{
				db:    persistencetest.NewInMemoryDB(t, types, joinTables),
				authz: &service.AuthorizationStrategyAllowAll{},
			},
			want: assert.Nil[*connect.Response[orchestrator.RiskAssessment]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeNotFound)
			},
			wantDB: assert.NotNil[persistence.DB],
		},
		{
			name: "audit scope of another target of evaluation",
			args: args{req: &orchestrator.CreateRiskAssessmentRequest{RiskAssessment: &orchestrator.RiskAssessment{
				TargetOfEvaluationId: orchestratortest.MockToeId2,
				AuditScopeId:         orchestratortest.MockScopeId1,
				Method:               orchestrator.RiskAssessmentMethod_RISK_ASSESSMENT_METHOD_BUILT_IN,
			}}},
			fields: fields{
				db:    persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) { seedRiskAssessmentScope(t, d) }),
				authz: &service.AuthorizationStrategyAllowAll{},
			},
			want: assert.Nil[*connect.Response[orchestrator.RiskAssessment]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeInvalidArgument) &&
					assert.ErrorContains(t, err, "does not belong")
			},
			wantDB: assert.NotNil[persistence.DB],
		},
		{
			name: "authorization failure",
			args: args{req: valid()},
			fields: fields{
				db:    persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) { seedRiskAssessmentScope(t, d) }),
				authz: &denyAuthorizationStrategy{},
			},
			want: assert.Nil[*connect.Response[orchestrator.RiskAssessment]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodePermissionDenied)
			},
			wantDB: assert.NotNil[persistence.DB],
		},
		{
			name: "decision for a control that is not part of the catalog",
			args: args{req: valid(outOfScope(orchestratortest.MockNonExistentId, mockRiskJustification1))},
			fields: fields{
				db:    persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) { seedRiskAssessmentScope(t, d) }),
				authz: &service.AuthorizationStrategyAllowAll{},
			},
			want: assert.Nil[*connect.Response[orchestrator.RiskAssessment]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeInvalidArgument) &&
					assert.ErrorContains(t, err, "not part of the catalog")
			},
			wantDB: assert.NotNil[persistence.DB],
		},
		{
			name: "a control with requirements cannot be taken out of scope as a whole",
			args: args{req: valid(outOfScope(orchestratortest.MockControlId1, mockRiskJustification1))},
			fields: fields{
				db:    persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) { seedRiskAssessmentScope(t, d) }),
				authz: &service.AuthorizationStrategyAllowAll{},
			},
			want: assert.Nil[*connect.Response[orchestrator.RiskAssessment]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeInvalidArgument) &&
					assert.ErrorContains(t, err, "decide on its requirements")
			},
			wantDB: assert.NotNil[persistence.DB],
		},
		{
			name: "all requirements of a control cannot be taken out of scope",
			args: args{req: valid(outOfScope(orchestratortest.MockControl2SubControlId1, mockRiskJustification1))},
			fields: fields{
				db:    persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) { seedRiskAssessmentScope(t, d) }),
				authz: &service.AuthorizationStrategyAllowAll{},
			},
			want: assert.Nil[*connect.Response[orchestrator.RiskAssessment]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeInvalidArgument) &&
					assert.ErrorContains(t, err, "at least one requirement")
			},
			wantDB: assert.NotNil[persistence.DB],
		},
		{
			name: "a requirement can be taken out of scope if another one of the control stays",
			args: args{req: valid(outOfScope(orchestratortest.MockControl2SubControlId1, mockRiskJustification1))},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					seedRiskAssessmentScope(t, d)
					assert.NoError(t, d.Create(&orchestrator.Control{Id: orchestratortest.MockControlId31, CatalogId: orchestratortest.MockCatalogId1, Name: orchestratortest.MockControlName2, ShortName: orchestratortest.MockControlShortName2, ParentControlId: new(orchestratortest.MockControlId1)}))
				}),
				authz: &service.AuthorizationStrategyAllowAll{},
			},
			want:    assert.NotNil[*connect.Response[orchestrator.RiskAssessment]],
			wantErr: assert.NoError,
			wantDB:  assert.NotNil[persistence.DB],
		},
		{
			name: "two decisions for the same control",
			args: args{req: valid(
				outOfScope(orchestratortest.MockControlId2, mockRiskJustification1),
				outOfScope(orchestratortest.MockControlId2, mockRiskJustification1),
			)},
			fields: fields{
				db:    persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) { seedRiskAssessmentScope(t, d) }),
				authz: &service.AuthorizationStrategyAllowAll{},
			},
			want: assert.Nil[*connect.Response[orchestrator.RiskAssessment]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeInvalidArgument) &&
					assert.ErrorContains(t, err, "more than one decision")
			},
			wantDB: assert.NotNil[persistence.DB],
		},
		{
			name: "database error",
			args: args{req: valid()},
			fields: fields{
				db:    persistencetest.CreateErrorDB(t, persistence.ErrUniqueConstraintFailed, types, joinTables, func(d persistence.DB) { seedRiskAssessmentScope(t, d) }),
				authz: &service.AuthorizationStrategyAllowAll{},
			},
			want: assert.Nil[*connect.Response[orchestrator.RiskAssessment]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeAlreadyExists)
			},
			wantDB: assert.NotNil[persistence.DB],
		},
		{
			name: "happy path: first assessment of the audit scope is a draft of version 1",
			args: args{req: valid(outOfScope(orchestratortest.MockControlId2, mockRiskJustification1))},
			fields: fields{
				db:    persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) { seedRiskAssessmentScope(t, d) }),
				authz: &service.AuthorizationStrategyAllowAll{},
			},
			want: func(t *testing.T, got *connect.Response[orchestrator.RiskAssessment], msgAndArgs ...any) bool {
				return assert.NotEmpty(t, got.Msg.GetId()) &&
					assert.Equal(t, int32(1), got.Msg.GetVersion()) &&
					assert.Equal(t, orchestrator.RiskAssessmentStatus_RISK_ASSESSMENT_STATUS_DRAFT, got.Msg.GetStatus()) &&
					assert.Equal(t, 1, len(got.Msg.GetDecisions())) &&
					assert.NotNil(t, got.Msg.GetCreatedAt())
			},
			wantErr: assert.NoError,
			wantDB: func(t *testing.T, db persistence.DB, msgAndArgs ...any) bool {
				// Creating the draft does not change the scope.
				count, err := db.Count(&orchestrator.RiskAssessment{}, "audit_scope_id = ?", orchestratortest.MockScopeId1)
				return assert.NoError(t, err) && assert.Equal(t, int64(1), count) &&
					assert.Equal(t, int64(1), countControlInScope(t, db, orchestratortest.MockControlId2))
			},
		},
		{
			name: "happy path: the version counts up",
			args: args{req: valid()},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					seedRiskAssessmentScope(t, d)
					assert.NoError(t, d.Create(mockRiskAssessment(mockRiskAssessmentId1, 1, orchestrator.RiskAssessmentStatus_RISK_ASSESSMENT_STATUS_APPROVED)))
				}),
				authz: &service.AuthorizationStrategyAllowAll{},
			},
			want: func(t *testing.T, got *connect.Response[orchestrator.RiskAssessment], msgAndArgs ...any) bool {
				return assert.Equal(t, int32(2), got.Msg.GetVersion())
			},
			wantErr: assert.NoError,
			wantDB:  assert.NotNil[persistence.DB],
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &Service{
				db:    tt.fields.db,
				authz: tt.fields.authz,
			}
			res, err := svc.CreateRiskAssessment(tt.args.context, connect.NewRequest(tt.args.req))
			tt.want(t, res)
			tt.wantErr(t, err)
			tt.wantDB(t, tt.fields.db)
		})
	}
}

func TestService_ApproveRiskAssessment(t *testing.T) {
	type args struct {
		req     *orchestrator.ApproveRiskAssessmentRequest
		context context.Context
	}
	type fields struct {
		db    persistence.DB
		authz service.AuthorizationStrategy
	}
	approve := func(id string) *orchestrator.ApproveRiskAssessmentRequest {
		return &orchestrator.ApproveRiskAssessmentRequest{RiskAssessmentId: id}
	}
	draft := orchestrator.RiskAssessmentStatus_RISK_ASSESSMENT_STATUS_DRAFT
	approved := orchestrator.RiskAssessmentStatus_RISK_ASSESSMENT_STATUS_APPROVED
	tests := []struct {
		name    string
		args    args
		fields  fields
		want    assert.Want[*connect.Response[orchestrator.RiskAssessment]]
		wantErr assert.WantErr
		wantDB  assert.Want[persistence.DB]
	}{
		{
			name: "validation error - empty request",
			args: args{req: &orchestrator.ApproveRiskAssessmentRequest{}},
			fields: fields{
				db:    persistencetest.NewInMemoryDB(t, types, joinTables),
				authz: &service.AuthorizationStrategyAllowAll{},
			},
			want: assert.Nil[*connect.Response[orchestrator.RiskAssessment]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeInvalidArgument)
			},
			wantDB: assert.NotNil[persistence.DB],
		},
		{
			name: "risk assessment not found",
			args: args{req: approve(mockRiskAssessmentId1)},
			fields: fields{
				db:    persistencetest.NewInMemoryDB(t, types, joinTables),
				authz: &service.AuthorizationStrategyAllowAll{},
			},
			want: assert.Nil[*connect.Response[orchestrator.RiskAssessment]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeNotFound)
			},
			wantDB: assert.NotNil[persistence.DB],
		},
		{
			name: "audit scope not found",
			args: args{req: approve(mockRiskAssessmentId1)},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					assert.NoError(t, d.Create(mockRiskAssessment(mockRiskAssessmentId1, 1, draft)))
				}),
				authz: &service.AuthorizationStrategyAllowAll{},
			},
			want: assert.Nil[*connect.Response[orchestrator.RiskAssessment]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeNotFound)
			},
			wantDB: assert.NotNil[persistence.DB],
		},
		{
			name: "authorization failure",
			args: args{req: approve(mockRiskAssessmentId1)},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					seedRiskAssessmentScope(t, d)
					assert.NoError(t, d.Create(mockRiskAssessment(mockRiskAssessmentId1, 1, draft)))
				}),
				authz: &denyAuthorizationStrategy{},
			},
			want: assert.Nil[*connect.Response[orchestrator.RiskAssessment]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodePermissionDenied)
			},
			wantDB: assert.NotNil[persistence.DB],
		},
		{
			name: "only a draft can be approved",
			args: args{req: approve(mockRiskAssessmentId1)},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					seedRiskAssessmentScope(t, d)
					assert.NoError(t, d.Create(mockRiskAssessment(mockRiskAssessmentId1, 1, approved)))
				}),
				authz: &service.AuthorizationStrategyAllowAll{},
			},
			want: assert.Nil[*connect.Response[orchestrator.RiskAssessment]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeFailedPrecondition)
			},
			wantDB: assert.NotNil[persistence.DB],
		},
		{
			name: "happy path: a control that is out of scope is removed with its reason in the audit trail",
			args: args{req: approve(mockRiskAssessmentId1)},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					seedRiskAssessmentScope(t, d)
					assert.NoError(t, d.Create(mockRiskAssessment(mockRiskAssessmentId1, 1, draft,
						outOfScope(orchestratortest.MockControlId2, mockRiskJustification1))))
				}),
				authz: &service.AuthorizationStrategyAllowAll{},
			},
			want: func(t *testing.T, got *connect.Response[orchestrator.RiskAssessment], msgAndArgs ...any) bool {
				return assert.Equal(t, orchestrator.RiskAssessmentStatus_RISK_ASSESSMENT_STATUS_APPROVED, got.Msg.GetStatus()) &&
					assert.NotNil(t, got.Msg.GetApprovedAt()) &&
					assert.NotNil(t, got.Msg.GetReviewDueAt())
			},
			wantErr: assert.NoError,
			wantDB: func(t *testing.T, db persistence.DB, msgAndArgs ...any) bool {
				var event orchestrator.AuditTrailEvent
				err := db.Get(&event, "audit_scope_id = ?", orchestratortest.MockScopeId1)

				return assert.Equal(t, int64(0), countControlInScope(t, db, orchestratortest.MockControlId2)) &&
					assert.Equal(t, int64(1), countControlInScope(t, db, orchestratortest.MockControlId1)) &&
					assert.NoError(t, err) &&
					assert.Equal(t, "Risk assessment, version 1: "+mockRiskJustification1, event.GetComment())
			},
		},
		{
			name: "happy path: a parent control that is out of scope takes its sub-controls with it",
			args: args{req: approve(mockRiskAssessmentId1)},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					seedRiskAssessmentScope(t, d)
					assert.NoError(t, d.Create(mockRiskAssessment(mockRiskAssessmentId1, 1, draft,
						outOfScope(orchestratortest.MockControlId1, mockRiskJustification1),
						outOfScope(orchestratortest.MockControl2SubControlId1, mockRiskJustification1))))
				}),
				authz: &service.AuthorizationStrategyAllowAll{},
			},
			want:    assert.NotNil[*connect.Response[orchestrator.RiskAssessment]],
			wantErr: assert.NoError,
			wantDB: func(t *testing.T, db persistence.DB, msgAndArgs ...any) bool {
				return assert.Equal(t, int64(0), countControlInScope(t, db, orchestratortest.MockControlId1)) &&
					assert.Equal(t, int64(0), countControlInScope(t, db, orchestratortest.MockControl2SubControlId1)) &&
					assert.Equal(t, int64(1), countControlInScope(t, db, orchestratortest.MockControlId2))
			},
		},
		{
			name: "happy path: the previous assessment is superseded and its controls that are not out of scope anymore come back",
			args: args{req: approve(mockRiskAssessmentId2)},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					seedRiskAssessmentScope(t, d)
					// The first assessment took the parent control with its sub-control out of scope
					assert.NoError(t, d.Delete(&orchestrator.ControlInScope{}, "control_id IN ?", []string{orchestratortest.MockControlId1, orchestratortest.MockControl2SubControlId1}))
					assert.NoError(t, d.Create(mockRiskAssessment(mockRiskAssessmentId1, 1, approved,
						outOfScope(orchestratortest.MockControlId1, mockRiskJustification1))))
					assert.NoError(t, d.Create(mockRiskAssessment(mockRiskAssessmentId2, 2, draft)))
				}),
				authz: &service.AuthorizationStrategyAllowAll{},
			},
			want:    assert.NotNil[*connect.Response[orchestrator.RiskAssessment]],
			wantErr: assert.NoError,
			wantDB: func(t *testing.T, db persistence.DB, msgAndArgs ...any) bool {
				first := assert.InDBGet[orchestrator.RiskAssessment](t, db, mockRiskAssessmentId1)
				second := assert.InDBGet[orchestrator.RiskAssessment](t, db, mockRiskAssessmentId2)

				return assert.Equal(t, orchestrator.RiskAssessmentStatus_RISK_ASSESSMENT_STATUS_SUPERSEDED, first.GetStatus()) &&
					assert.Equal(t, orchestrator.RiskAssessmentStatus_RISK_ASSESSMENT_STATUS_APPROVED, second.GetStatus()) &&
					assert.Equal(t, int64(1), countControlInScope(t, db, orchestratortest.MockControlId1)) &&
					assert.Equal(t, int64(1), countControlInScope(t, db, orchestratortest.MockControl2SubControlId1))
			},
		},
		{
			name: "database error",
			args: args{req: approve(mockRiskAssessmentId1)},
			fields: fields{
				db: persistencetest.SaveErrorDB(t, persistence.ErrDatabase, types, joinTables, func(d persistence.DB) {
					seedRiskAssessmentScope(t, d)
					assert.NoError(t, d.Create(mockRiskAssessment(mockRiskAssessmentId1, 1, draft)))
				}),
				authz: &service.AuthorizationStrategyAllowAll{},
			},
			want: assert.Nil[*connect.Response[orchestrator.RiskAssessment]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeInternal)
			},
			wantDB: assert.NotNil[persistence.DB],
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &Service{
				db:    tt.fields.db,
				authz: tt.fields.authz,
			}
			res, err := svc.ApproveRiskAssessment(tt.args.context, connect.NewRequest(tt.args.req))
			tt.want(t, res)
			tt.wantErr(t, err)
			tt.wantDB(t, tt.fields.db)
		})
	}
}

func TestService_GetRiskAssessment(t *testing.T) {
	type args struct {
		req     *orchestrator.GetRiskAssessmentRequest
		context context.Context
	}
	type fields struct {
		db    persistence.DB
		authz service.AuthorizationStrategy
	}
	seed := func(d persistence.DB) {
		assert.NoError(t, d.Create(mockRiskAssessment(mockRiskAssessmentId1, 1, orchestrator.RiskAssessmentStatus_RISK_ASSESSMENT_STATUS_DRAFT,
			outOfScope(orchestratortest.MockControlId2, mockRiskJustification1))))
	}
	tests := []struct {
		name    string
		args    args
		fields  fields
		want    assert.Want[*connect.Response[orchestrator.RiskAssessment]]
		wantErr assert.WantErr
	}{
		{
			name: "validation error - empty request",
			args: args{req: &orchestrator.GetRiskAssessmentRequest{}},
			fields: fields{
				db:    persistencetest.NewInMemoryDB(t, types, joinTables),
				authz: &service.AuthorizationStrategyAllowAll{},
			},
			want: assert.Nil[*connect.Response[orchestrator.RiskAssessment]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeInvalidArgument)
			},
		},
		{
			name: "not found",
			args: args{req: &orchestrator.GetRiskAssessmentRequest{RiskAssessmentId: mockRiskAssessmentId1}},
			fields: fields{
				db:    persistencetest.NewInMemoryDB(t, types, joinTables),
				authz: &service.AuthorizationStrategyAllowAll{},
			},
			want: assert.Nil[*connect.Response[orchestrator.RiskAssessment]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeNotFound)
			},
		},
		{
			name: "authorization failure",
			args: args{req: &orchestrator.GetRiskAssessmentRequest{RiskAssessmentId: mockRiskAssessmentId1}},
			fields: fields{
				db:    persistencetest.NewInMemoryDB(t, types, joinTables, seed),
				authz: &denyAuthorizationStrategy{},
			},
			want: assert.Nil[*connect.Response[orchestrator.RiskAssessment]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodePermissionDenied)
			},
		},
		{
			name: "happy path",
			args: args{req: &orchestrator.GetRiskAssessmentRequest{RiskAssessmentId: mockRiskAssessmentId1}},
			fields: fields{
				db:    persistencetest.NewInMemoryDB(t, types, joinTables, seed),
				authz: &service.AuthorizationStrategyAllowAll{},
			},
			want: func(t *testing.T, got *connect.Response[orchestrator.RiskAssessment], msgAndArgs ...any) bool {
				return assert.Equal(t, mockRiskAssessmentId1, got.Msg.GetId()) &&
					assert.Equal(t, 1, len(got.Msg.GetDecisions())) &&
					assert.Equal(t, mockRiskJustification1, got.Msg.GetDecisions()[0].GetJustification())
			},
			wantErr: assert.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &Service{
				db:    tt.fields.db,
				authz: tt.fields.authz,
			}
			res, err := svc.GetRiskAssessment(tt.args.context, connect.NewRequest(tt.args.req))
			tt.want(t, res)
			tt.wantErr(t, err)
		})
	}
}

func TestService_ListRiskAssessments(t *testing.T) {
	type args struct {
		req     *orchestrator.ListRiskAssessmentsRequest
		context context.Context
	}
	type fields struct {
		db    persistence.DB
		authz service.AuthorizationStrategy
	}
	seed := func(d persistence.DB) {
		assert.NoError(t, d.Create(mockRiskAssessment(mockRiskAssessmentId1, 1, orchestrator.RiskAssessmentStatus_RISK_ASSESSMENT_STATUS_SUPERSEDED)))
		assert.NoError(t, d.Create(mockRiskAssessment(mockRiskAssessmentId2, 2, orchestrator.RiskAssessmentStatus_RISK_ASSESSMENT_STATUS_APPROVED)))
	}
	tests := []struct {
		name    string
		args    args
		fields  fields
		want    assert.Want[*connect.Response[orchestrator.ListRiskAssessmentsResponse]]
		wantErr assert.WantErr
	}{
		{
			name: "no access to any audit scope",
			args: args{req: &orchestrator.ListRiskAssessmentsRequest{}},
			fields: fields{
				db:    persistencetest.NewInMemoryDB(t, types, joinTables, seed),
				authz: &denyAuthorizationStrategy{},
			},
			want: func(t *testing.T, got *connect.Response[orchestrator.ListRiskAssessmentsResponse], msgAndArgs ...any) bool {
				return assert.Empty(t, got.Msg.GetRiskAssessments())
			},
			wantErr: assert.NoError,
		},
		{
			name: "happy path: all assessments",
			args: args{req: &orchestrator.ListRiskAssessmentsRequest{}},
			fields: fields{
				db:    persistencetest.NewInMemoryDB(t, types, joinTables, seed),
				authz: &service.AuthorizationStrategyAllowAll{},
			},
			want: func(t *testing.T, got *connect.Response[orchestrator.ListRiskAssessmentsResponse], msgAndArgs ...any) bool {
				return assert.Equal(t, 2, len(got.Msg.GetRiskAssessments()))
			},
			wantErr: assert.NoError,
		},
		{
			name: "happy path: filtered by audit scope and target of evaluation",
			args: args{req: &orchestrator.ListRiskAssessmentsRequest{Filter: &orchestrator.ListRiskAssessmentsRequest_Filter{
				AuditScopeId:         new(orchestratortest.MockScopeId2),
				TargetOfEvaluationId: new(orchestratortest.MockToeId1),
			}}},
			fields: fields{
				db:    persistencetest.NewInMemoryDB(t, types, joinTables, seed),
				authz: &service.AuthorizationStrategyAllowAll{},
			},
			want: func(t *testing.T, got *connect.Response[orchestrator.ListRiskAssessmentsResponse], msgAndArgs ...any) bool {
				return assert.Empty(t, got.Msg.GetRiskAssessments())
			},
			wantErr: assert.NoError,
		},
		{
			name: "database error",
			args: args{req: &orchestrator.ListRiskAssessmentsRequest{}},
			fields: fields{
				db:    persistencetest.ListErrorDB(t, persistence.ErrDatabase, types, joinTables),
				authz: &service.AuthorizationStrategyAllowAll{},
			},
			want: assert.Nil[*connect.Response[orchestrator.ListRiskAssessmentsResponse]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeInternal)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &Service{
				db:    tt.fields.db,
				authz: tt.fields.authz,
			}
			res, err := svc.ListRiskAssessments(tt.args.context, connect.NewRequest(tt.args.req))
			tt.want(t, res)
			tt.wantErr(t, err)
		})
	}
}
