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

package evaluation

import (
	"context"
	"errors"
	"testing"
	"time"

	"confirmate.io/core/api/evaluation"
	"confirmate.io/core/api/orchestrator"
	"confirmate.io/core/auth"
	"confirmate.io/core/service"
	"confirmate.io/core/service/evaluation/evaluationtest"
	"confirmate.io/core/service/orchestrator/orchestratortest"
	"confirmate.io/core/util/assert"

	"connectrpc.com/connect"
	"github.com/go-co-op/gocron"
	"github.com/golang-jwt/jwt/v5"
)

// TestService_StopEvaluation_permissionStore checks the authorization of [Service.StopEvaluation] with the permission
// store that is backed by the orchestrator (as in a real deployment): only users with a permission for the audit scope
// may stop its evaluation.
func TestService_StopEvaluation_permissionStore(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name        string
		permissions []*orchestrator.UserPermission
		listError   error
		args        args
		wantErr     assert.WantErr
	}{
		{
			name:        "user with a permission for the audit scope: access is granted, but there is no running evaluation",
			permissions: []*orchestrator.UserPermission{orchestratortest.MockUserPermissionsAuditScopeAdmin},
			args: args{ctx: auth.WithClaims(context.Background(), &auth.OAuthClaims{
				RegisteredClaims: jwt.RegisteredClaims{Subject: orchestratortest.MockUserId1, Issuer: orchestratortest.MockUserIssuer1},
			})},
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeFailedPrecondition)
			},
		},
		{
			name: "user without any permission: access is denied",
			args: args{ctx: auth.WithClaims(context.Background(), &auth.OAuthClaims{
				RegisteredClaims: jwt.RegisteredClaims{Subject: orchestratortest.MockUserId1, Issuer: orchestratortest.MockUserIssuer1},
			})},
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodePermissionDenied)
			},
		},
		{
			name: "user without a permission for the audit scope, only another user has one: access is denied",
			permissions: []*orchestrator.UserPermission{
				{
					UserId:     orchestratortest.GetConfirmateUserID(orchestratortest.MockUserIssuer1, orchestratortest.MockNonExistentId),
					ObjectId:   evaluationtest.MockAuditScopeId1,
					ObjectType: orchestrator.ObjectType_OBJECT_TYPE_AUDIT_SCOPE,
					Permission: orchestrator.UserPermission_PERMISSION_ADMIN,
				},
			},
			args: args{ctx: auth.WithClaims(context.Background(), &auth.OAuthClaims{
				RegisteredClaims: jwt.RegisteredClaims{Subject: orchestratortest.MockUserId1, Issuer: orchestratortest.MockUserIssuer1},
			})},
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodePermissionDenied)
			},
		},
		{
			name:        "the permissions cannot be listed: access is denied",
			permissions: []*orchestrator.UserPermission{orchestratortest.MockUserPermissionsAuditScopeAdmin},
			listError:   errors.New("boom"),
			args: args{ctx: auth.WithClaims(context.Background(), &auth.OAuthClaims{
				RegisteredClaims: jwt.RegisteredClaims{Subject: orchestratortest.MockUserId1, Issuer: orchestratortest.MockUserIssuer1},
			})},
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodePermissionDenied)
			},
		},
		{
			name:        "no user at all: access is denied",
			permissions: []*orchestrator.UserPermission{orchestratortest.MockUserPermissionsAuditScopeAdmin},
			args:        args{ctx: context.Background()},
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodePermissionDenied)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := newOrchestratorClient(t, WithUserPermissions(tt.permissions), func(m *mockOrchestratorHandler) {
				m.listUserPermissionsError = tt.listError
			})
			svc := Service{
				orchestratorClient: client,
				scheduler:          gocron.NewScheduler(time.UTC),
				authz: &service.AuthorizationStrategyPermissionStore{
					Permissions: &service.OrchestratorPermissionStore{Client: client},
				},
			}

			_, err := svc.StopEvaluation(tt.args.ctx, connect.NewRequest(&evaluation.StopEvaluationRequest{
				AuditScopeId: evaluationtest.MockAuditScopeId1,
			}))

			tt.wantErr(t, err)
		})
	}
}

// TestService_StartEvaluation_orchestratorErrors covers failing requests to the orchestrator while starting an
// evaluation.
func TestService_StartEvaluation_orchestratorErrors(t *testing.T) {
	tests := []struct {
		name    string
		opts    []func(*mockOrchestratorHandler)
		wantErr assert.WantErr
	}{
		{
			name: "the catalog cannot be fetched",
			opts: []func(*mockOrchestratorHandler){
				WithAuditScope(evaluationtest.MockAuditScope1),
				WithControls([]*orchestrator.Control{evaluationtest.MockControl1}),
				WithGetCatalogError(errors.New("boom")),
			},
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeInternal) &&
					assert.ErrorContains(t, err, "could not get catalog from the orchestrator")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := Service{
				orchestratorClient: newOrchestratorClient(t, tt.opts...),
				scheduler:          gocron.NewScheduler(time.UTC),
				catalogControls:    make(map[string]map[string]*orchestrator.Control),
			}
			t.Cleanup(svc.Shutdown)

			_, err := svc.StartEvaluation(context.Background(), connect.NewRequest(&evaluation.StartEvaluationRequest{
				AuditScopeId: evaluationtest.MockAuditScopeId1,
			}))

			tt.wantErr(t, err)
		})
	}
}

// TestService_triggerEvaluation_oneShotErrors covers an immediate evaluation without a scheduled job (a one-shot run)
// that fails because of the orchestrator.
func TestService_triggerEvaluation_oneShotErrors(t *testing.T) {
	tests := []struct {
		name    string
		opts    []func(*mockOrchestratorHandler)
		wantErr assert.WantErr
	}{
		{
			name: "the controls cannot be cached",
			opts: []func(*mockOrchestratorHandler){
				WithAuditScope(evaluationtest.MockAuditScope1),
				WithCatalog(evaluationtest.MockCatalog1),
				func(m *mockOrchestratorHandler) { m.listError = errors.New("boom") },
			},
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeInternal) &&
					assert.ErrorContains(t, err, "could not cache controls")
			},
		},
		{
			name: "the evaluation fails",
			opts: []func(*mockOrchestratorHandler){
				WithAuditScope(evaluationtest.MockAuditScope1),
				WithCatalog(evaluationtest.MockCatalog1),
				WithAdditionalControls(evaluationtest.MockControl1),
				func(m *mockOrchestratorHandler) { m.listEvalError = errors.New("boom") },
			},
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeInternal) &&
					assert.ErrorContains(t, err, "evaluation failed")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := Service{
				orchestratorClient: newOrchestratorClient(t, tt.opts...),
				scheduler:          gocron.NewScheduler(time.UTC),
				catalogControls:    make(map[string]map[string]*orchestrator.Control),
				evaluating:         make(map[string]struct{}),
			}
			t.Cleanup(svc.Shutdown)

			err := svc.triggerEvaluation(context.Background(), evaluationtest.MockAuditScopeId1)

			tt.wantErr(t, err)
		})
	}
}
