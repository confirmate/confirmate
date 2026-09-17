// Copyright 2016-2026 Fraunhofer AISEC
//
// SPDX-License-Identifier: Apache-2.0
//
//                                 /$$$$$$  /$$                                     /$$
//                               /$$__  $$|__/                                    | $$
//   /$$$$$$$  /$$$$$$  /$$$$$$$ | $$  \__/ /$$  /$$$$$$  /$$$$$$/$$$$   /$$$$$$  /$$$$$$    /$$$$$$
//  /$$_____/ /$$__  $$| $$__  $$| $$$$    | $$ /$$__  $$| $$_  $$_  $$ |____  $$|_  $$_/   /$$__  $$
// | $$      | $$  \ $$| $$  \ $$| $$_/    | $$| $$  \__/| $$ \ $$ \ $$  /$$$$$$$  | $$    | $$$$$$$$
// | $$      | $$  | $$| $$      | $$| $$      | $$ | $$ | $$ /$$__  $$  | $$ /$$| $$_____/
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
	"confirmate.io/core/auth"
	"confirmate.io/core/persistence"
	"confirmate.io/core/persistence/persistencetest"
	"confirmate.io/core/service"
	"confirmate.io/core/service/orchestrator/orchestratortest"
	"confirmate.io/core/util/assert"
	"github.com/golang-jwt/jwt/v5"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	statsCatalogId = "00000000-0000-0000-0009-000000000010"
	statsCtrl1Id   = "00000000-0000-0000-000a-000000000011"
	statsCtrl1Sub  = "00000000-0000-0000-000a-000000000012"
	statsCtrl2Id   = "00000000-0000-0000-000a-000000000013"
)

// newAuditScopeStatisticsTestDB seeds a catalog with two top-level controls (one of which has a
// sub-control), an audit scope (reusing orchestratortest.MockScopeId1 so the shared permission
// fixtures apply), ControlInScope records at mixed workflow states, and EvaluationResult records
// at mixed compliance statuses (including a superseded one, to exercise "latest per control").
func newAuditScopeStatisticsTestDB(t *testing.T) persistence.DB {
	return persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
		assert.NoError(t, d.Create(&orchestrator.Catalog{Id: statsCatalogId, Name: "Test Catalog"}))
		assert.NoError(t, d.Create(&orchestrator.Control{Id: statsCtrl1Id, ShortName: "C-1", Name: "Control 1", CatalogId: statsCatalogId}))
		assert.NoError(t, d.Create(&orchestrator.Control{Id: statsCtrl1Sub, ShortName: "C-1.1", Name: "Sub-control 1.1", CatalogId: statsCatalogId, ParentControlId: new(statsCtrl1Id)}))
		assert.NoError(t, d.Create(&orchestrator.Control{Id: statsCtrl2Id, ShortName: "C-2", Name: "Control 2", CatalogId: statsCatalogId}))
		assert.NoError(t, d.Create(&orchestrator.AuditScope{
			Id:                   orchestratortest.MockScopeId1,
			Name:                 "Test Scope",
			TargetOfEvaluationId: orchestratortest.MockToeId1,
			CatalogId:            statsCatalogId,
			Status:               orchestrator.AuditScopeStatus_AUDIT_SCOPE_STATUS_SETUP,
		}))

		// Workflow status: only top-level controls (ctrl1, ctrl2) should be counted. The
		// sub-control's state (OPEN) must not leak into the aggregation.
		assert.NoError(t, d.Create(&orchestrator.ControlInScope{
			Id: "00000000-0000-0000-000b-000000000001", AuditScopeId: orchestratortest.MockScopeId1, TargetOfEvaluationId: orchestratortest.MockToeId1,
			ControlId: statsCtrl1Id, State: orchestrator.ControlInScopeState_CONTROL_IN_SCOPE_STATE_IMPLEMENTED,
		}))
		assert.NoError(t, d.Create(&orchestrator.ControlInScope{
			Id: "00000000-0000-0000-000b-000000000002", AuditScopeId: orchestratortest.MockScopeId1, TargetOfEvaluationId: orchestratortest.MockToeId1,
			ControlId: statsCtrl1Sub, State: orchestrator.ControlInScopeState_CONTROL_IN_SCOPE_STATE_OPEN,
		}))
		assert.NoError(t, d.Create(&orchestrator.ControlInScope{
			Id: "00000000-0000-0000-000b-000000000003", AuditScopeId: orchestratortest.MockScopeId1, TargetOfEvaluationId: orchestratortest.MockToeId1,
			ControlId: statsCtrl2Id, State: orchestrator.ControlInScopeState_CONTROL_IN_SCOPE_STATE_ACCEPTED,
		}))

		// Compliance status: an evaluation on the sub-control must be counted (metrics
		// typically attach there), and only the latest of ctrl2's two results should count.
		assert.NoError(t, d.Create(&evaluation.EvaluationResult{
			Id: "00000000-0000-0000-000c-000000000001", TargetOfEvaluationId: orchestratortest.MockToeId1, AuditScopeId: orchestratortest.MockScopeId1,
			ControlId: statsCtrl1Id, ControlCatalogId: statsCatalogId,
			Status:    evaluation.EvaluationStatus_EVALUATION_STATUS_COMPLIANT,
			Timestamp: timestamppb.New(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)),
		}))
		assert.NoError(t, d.Create(&evaluation.EvaluationResult{
			Id: "00000000-0000-0000-000c-000000000002", TargetOfEvaluationId: orchestratortest.MockToeId1, AuditScopeId: orchestratortest.MockScopeId1,
			ControlId: statsCtrl1Sub, ControlCatalogId: statsCatalogId, ParentControlId: new(statsCtrl1Id),
			Status:    evaluation.EvaluationStatus_EVALUATION_STATUS_NOT_COMPLIANT,
			Timestamp: timestamppb.New(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)),
		}))
		assert.NoError(t, d.Create(&evaluation.EvaluationResult{
			Id: "00000000-0000-0000-000c-000000000003", TargetOfEvaluationId: orchestratortest.MockToeId1, AuditScopeId: orchestratortest.MockScopeId1,
			ControlId: statsCtrl2Id, ControlCatalogId: statsCatalogId,
			Status:    evaluation.EvaluationStatus_EVALUATION_STATUS_PENDING,
			Timestamp: timestamppb.New(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)),
		}))
		assert.NoError(t, d.Create(&evaluation.EvaluationResult{
			Id: "00000000-0000-0000-000c-000000000004", TargetOfEvaluationId: orchestratortest.MockToeId1, AuditScopeId: orchestratortest.MockScopeId1,
			ControlId: statsCtrl2Id, ControlCatalogId: statsCatalogId,
			Status:    evaluation.EvaluationStatus_EVALUATION_STATUS_COMPLIANT,
			Timestamp: timestamppb.New(time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)),
		}))
	})
}

func TestService_GetAuditScopeStatistics(t *testing.T) {
	type args struct {
		req     *orchestrator.GetAuditScopeStatisticsRequest
		context context.Context
	}
	type fields struct {
		db    persistence.DB
		authz service.AuthorizationStrategy
	}
	tests := []struct {
		name    string
		args    args
		fields  fields
		want    assert.Want[*connect.Response[orchestrator.GetAuditScopeStatisticsResponse]]
		wantErr assert.WantErr
	}{
		{
			name: "happy path: with allow-all authorization strategy",
			args: args{
				req: &orchestrator.GetAuditScopeStatisticsRequest{
					AuditScopeId: orchestratortest.MockScopeId1,
				},
			},
			fields: fields{
				db:    newAuditScopeStatisticsTestDB(t),
				authz: &service.AuthorizationStrategyAllowAll{},
			},
			want: func(t *testing.T, got *connect.Response[orchestrator.GetAuditScopeStatisticsResponse], args ...any) bool {
				return assert.NotNil(t, got.Msg) &&
					assert.Equal(t, map[string]int64{
						"CONTROL_IN_SCOPE_STATE_IMPLEMENTED": 1,
						"CONTROL_IN_SCOPE_STATE_ACCEPTED":    1,
					}, got.Msg.GetCountsByControlState()) &&
					assert.Equal(t, map[string]int64{
						"EVALUATION_STATUS_COMPLIANT":     2,
						"EVALUATION_STATUS_NOT_COMPLIANT": 1,
					}, got.Msg.GetCountsByComplianceStatus())
			},
			wantErr: assert.NoError,
		},
		{
			name: "happy path: with authorization strategy with permission store and admin token",
			args: args{
				req: &orchestrator.GetAuditScopeStatisticsRequest{
					AuditScopeId: orchestratortest.MockScopeId1,
				},
				context: auth.WithClaims(context.Background(), &auth.OAuthClaims{
					IsAdminToken: true,
				}),
			},
			fields: fields{
				db:    newAuditScopeStatisticsTestDB(t),
				authz: &service.AuthorizationStrategyPermissionStore{},
			},
			want: func(t *testing.T, got *connect.Response[orchestrator.GetAuditScopeStatisticsResponse], args ...any) bool {
				return assert.NotNil(t, got.Msg) &&
					assert.Equal(t, map[string]int64{
						"CONTROL_IN_SCOPE_STATE_IMPLEMENTED": 1,
						"CONTROL_IN_SCOPE_STATE_ACCEPTED":    1,
					}, got.Msg.GetCountsByControlState())
			},
			wantErr: assert.NoError,
		},
		{
			name: "happy path: with authorization strategy with permission store and user permissions allowing access",
			args: args{
				req: &orchestrator.GetAuditScopeStatisticsRequest{
					AuditScopeId: orchestratortest.MockScopeId1,
				},
				context: auth.WithClaims(context.Background(), &auth.OAuthClaims{
					RegisteredClaims: jwt.RegisteredClaims{
						Subject: orchestratortest.MockUserId1,
						Issuer:  orchestratortest.MockUserIssuer1,
					},
				}),
			},
			fields: fields{
				db: newAuditScopeStatisticsTestDB(t),
				authz: &service.AuthorizationStrategyPermissionStore{
					Permissions: service.DBPermissionStore{
						DB: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
							err := d.Create(orchestratortest.MockUserPermissionsAuditScopeAdmin)
							assert.NoError(t, err)
						}),
					},
				},
			},
			want: func(t *testing.T, got *connect.Response[orchestrator.GetAuditScopeStatisticsResponse], args ...any) bool {
				return assert.NotNil(t, got.Msg) &&
					assert.Equal(t, map[string]int64{
						"CONTROL_IN_SCOPE_STATE_IMPLEMENTED": 1,
						"CONTROL_IN_SCOPE_STATE_ACCEPTED":    1,
					}, got.Msg.GetCountsByControlState())
			},
			wantErr: assert.NoError,
		},
		{
			name: "authorization failure",
			args: args{
				req: &orchestrator.GetAuditScopeStatisticsRequest{
					AuditScopeId: orchestratortest.MockScopeId1,
				},
			},
			fields: fields{
				db:    newAuditScopeStatisticsTestDB(t),
				authz: &denyAuthorizationStrategy{},
			},
			want: assert.Nil[*connect.Response[orchestrator.GetAuditScopeStatisticsResponse]],
			wantErr: func(t *testing.T, err error, args ...any) bool {
				return assert.IsConnectError(t, err, connect.CodePermissionDenied) &&
					assert.ErrorContains(t, err, service.ErrPermissionDenied.Error())
			},
		},
		{
			name: "validation error - empty request",
			args: args{
				req: &orchestrator.GetAuditScopeStatisticsRequest{},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables),
			},
			want: assert.Nil[*connect.Response[orchestrator.GetAuditScopeStatisticsResponse]],
			wantErr: func(t *testing.T, err error, args ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeInvalidArgument) &&
					assert.IsValidationError(t, err, "audit_scope_id")
			},
		},
		{
			name: "error: database error listing top-level controls",
			args: args{
				req: &orchestrator.GetAuditScopeStatisticsRequest{
					AuditScopeId: orchestratortest.MockScopeId1,
				},
			},
			fields: fields{
				db: persistencetest.ListErrorDB(t, persistence.ErrDatabase, types, joinTables, func(d persistence.DB) {
					err := d.Create(&orchestrator.AuditScope{Id: orchestratortest.MockScopeId1, Name: "Test Scope", TargetOfEvaluationId: orchestratortest.MockToeId1, CatalogId: statsCatalogId})
					assert.NoError(t, err)
				}),
				authz: &service.AuthorizationStrategyAllowAll{},
			},
			want: assert.Nil[*connect.Response[orchestrator.GetAuditScopeStatisticsResponse]],
			wantErr: func(t *testing.T, err error, args ...any) bool {
				return assert.ErrorContains(t, err, persistence.ErrDatabase.Error())
			},
		},
		{
			name: "error: database error listing evaluation results",
			args: args{
				req: &orchestrator.GetAuditScopeStatisticsRequest{
					AuditScopeId: orchestratortest.MockScopeId1,
				},
			},
			fields: fields{
				db: persistencetest.RawErrorDB(t, persistence.ErrDatabase, types, joinTables, func(d persistence.DB) {
					err := d.Create(&orchestrator.AuditScope{Id: orchestratortest.MockScopeId1, Name: "Test Scope", TargetOfEvaluationId: orchestratortest.MockToeId1, CatalogId: statsCatalogId})
					assert.NoError(t, err)
				}),
				authz: &service.AuthorizationStrategyAllowAll{},
			},
			want: assert.Nil[*connect.Response[orchestrator.GetAuditScopeStatisticsResponse]],
			wantErr: func(t *testing.T, err error, args ...any) bool {
				return assert.ErrorContains(t, err, persistence.ErrDatabase.Error())
			},
		},
		{
			name: "not found",
			args: args{
				req: &orchestrator.GetAuditScopeStatisticsRequest{
					AuditScopeId: orchestratortest.MockNonExistentId,
				},
			},
			fields: fields{
				db:    persistencetest.NewInMemoryDB(t, types, joinTables),
				authz: &service.AuthorizationStrategyAllowAll{},
			},
			want: assert.Nil[*connect.Response[orchestrator.GetAuditScopeStatisticsResponse]],
			wantErr: func(t *testing.T, err error, args ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeNotFound)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &Service{
				db:    tt.fields.db,
				authz: tt.fields.authz,
			}
			res, err := svc.GetAuditScopeStatistics(tt.args.context, connect.NewRequest(tt.args.req))
			tt.want(t, res)
			tt.wantErr(t, err)
		})
	}
}
