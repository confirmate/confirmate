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

package service

import (
	"context"
	"errors"
	"testing"

	"confirmate.io/core/api/orchestrator"
	"confirmate.io/core/auth"
	"confirmate.io/core/service/orchestrator/orchestratortest"
	"confirmate.io/core/util/assert"
	"github.com/golang-jwt/jwt/v5"
)

// var (
// 	// Used by persistencetest.NewInMemoryDB to auto-migrate schema for tests.
// 	types      = []any{}
// 	joinTables = []any{}
// )

type fakePermissionStore struct {
	ids []string
	err error
}

func (f *fakePermissionStore) HasPermission(ctx context.Context, userId string, resourceId string, permission orchestrator.UserPermission_Permission, reqType orchestrator.RequestType, objectType orchestrator.ObjectType) (bool, error) {
	panic("not used")
}

func (f *fakePermissionStore) PermissionForResources(ctx context.Context, userId string, permission orchestrator.UserPermission_Permission, reqType orchestrator.RequestType, objectType orchestrator.ObjectType) ([]string, error) {
	return f.ids, f.err
}

type denyAuthorizationStrategy struct{}

func (*denyAuthorizationStrategy) CheckAccess(_ context.Context, _ string, _ orchestrator.RequestType, _ orchestrator.UserPermission_Permission, _ string, _ orchestrator.ObjectType) (bool, []string) {
	return false, nil
}

func (*denyAuthorizationStrategy) AllowedTargetOfEvaluations(_ context.Context) (bool, []string) {
	return false, nil
}

func (*denyAuthorizationStrategy) AllowedAuditScopes(_ context.Context) (bool, []string) {
	return false, nil
}

func TestCheckAccess(t *testing.T) {
	tests := []struct {
		name  string
		authz AuthorizationStrategy
		want  assert.Want[bool]
	}{
		{
			name:  "nil strategy allows",
			authz: nil,
			want: func(t *testing.T, got bool, msgAndArgs ...any) bool {
				return assert.True(t, got)
			},
		},
		{
			name:  "delegates to strategy",
			authz: &denyAuthorizationStrategy{},
			want: func(t *testing.T, got bool, msgAndArgs ...any) bool {
				return assert.False(t, got)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := CheckAccess[struct{}](tt.authz, context.Background(), "user-1", orchestrator.RequestType_REQUEST_TYPE_GET, orchestrator.UserPermission_PERMISSION_READER, "resource-1", orchestrator.ObjectType_OBJECT_TYPE_TARGET_OF_EVALUATION)
			tt.want(t, got)
		})
	}
}

func TestAuthorizationStrategyJWT_CheckAccess(t *testing.T) {
	type args struct {
		ctx            context.Context
		userId         string
		reqType        orchestrator.RequestType
		userPermission orchestrator.UserPermission_Permission
		resourceId     string
		objectType     orchestrator.ObjectType
	}
	type fields struct {
		strategy *AuthorizationStrategyPermissionStore
	}

	tests := []struct {
		name    string
		args    args
		fields  fields
		want    assert.Want[bool]
		wantErr assert.WantErr
	}{
		{
			name: "empty userId returns false",
			args: args{
				ctx:            context.Background(),
				userId:         "",
				reqType:        orchestrator.RequestType_REQUEST_TYPE_GET,
				userPermission: orchestrator.UserPermission_PERMISSION_READER,
				resourceId:     "resource-1",
				objectType:     orchestrator.ObjectType_OBJECT_TYPE_TARGET_OF_EVALUATION,
			},
			fields: fields{strategy: &AuthorizationStrategyPermissionStore{}},
			want: func(t *testing.T, got bool, _ ...any) bool {
				return assert.False(t, got)
			},
			wantErr: assert.NoError,
		},
		{
			name: "allows when allow-all claim is true",
			args: args{
				ctx:            auth.WithClaims(context.Background(), &auth.OAuthClaims{IsAdminToken: true}),
				userId:         "user-1",
				reqType:        orchestrator.RequestType_REQUEST_TYPE_GET,
				userPermission: orchestrator.UserPermission_PERMISSION_READER,
				resourceId:     "any",
				objectType:     orchestrator.ObjectType_OBJECT_TYPE_TARGET_OF_EVALUATION,
			},
			fields: fields{strategy: &AuthorizationStrategyPermissionStore{}},
			want: func(t *testing.T, got bool, _ ...any) bool {
				return assert.True(t, got)
			},
			wantErr: assert.NoError,
		},
		{
			name: "no permissions store returns false",
			args: args{
				ctx:            context.Background(),
				userId:         "user-1",
				reqType:        orchestrator.RequestType_REQUEST_TYPE_GET,
				userPermission: orchestrator.UserPermission_PERMISSION_READER,
				resourceId:     "resource-1",
				objectType:     orchestrator.ObjectType_OBJECT_TYPE_TARGET_OF_EVALUATION,
			},
			fields: fields{strategy: &AuthorizationStrategyPermissionStore{}},
			want: func(t *testing.T, got bool, _ ...any) bool {
				return assert.False(t, got)
			},
			wantErr: assert.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := tt.fields.strategy.CheckAccess(tt.args.ctx, tt.args.userId, tt.args.reqType, tt.args.userPermission, tt.args.resourceId, tt.args.objectType)
			assert.True(t, tt.wantErr(t, nil))
			assert.True(t, tt.want(t, got))
		})
	}
}

func TestAuthorizationStrategyPermissionStore_AllowedAuditScopes(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	type fields struct {
		authz *AuthorizationStrategyPermissionStore
	}
	tests := []struct {
		name            string
		fields          fields
		args            args
		wantAllowed     assert.Want[bool]
		wantResourceIds assert.Want[[]string]
	}{
		{
			name: "err: authz is nil",
			args: args{
				ctx: context.Background(),
			},
			fields: fields{authz: nil},
			wantAllowed: func(t *testing.T, got bool, msgAndArgs ...any) bool {
				return assert.False(t, got)
			},
			wantResourceIds: assert.Nil[[]string],
		},
		{
			name: "err: permission store returns error",
			args: args{
				ctx: auth.WithClaims(
					context.Background(),
					&auth.OAuthClaims{
						RegisteredClaims: jwt.RegisteredClaims{
							Subject: orchestratortest.MockUserId1,
							Issuer:  orchestratortest.MockUserIssuer1,
						},
					},
				),
			},
			fields: fields{
				authz: &AuthorizationStrategyPermissionStore{
					Permissions: &fakePermissionStore{
						ids: nil,
						err: errors.New("some error"),
					},
				},
			},
			wantAllowed: func(t *testing.T, got bool, msgAndArgs ...any) bool {
				return assert.False(t, got)
			},
			wantResourceIds: assert.Nil[[]string],
		},
		{
			name: "happy path: admin token allows all audit scopes",
			args: args{
				ctx: auth.WithClaims(
					context.Background(),
					&auth.OAuthClaims{IsAdminToken: true},
				),
			},
			fields: fields{authz: &AuthorizationStrategyPermissionStore{}},
			wantAllowed: func(t *testing.T, got bool, msgAndArgs ...any) bool {
				return assert.True(t, got)
			},
			wantResourceIds: assert.Nil[[]string],
		},
		{
			name: "happy path: permissions store allows specific audit scopes",
			args: args{
				auth.WithClaims(
					context.Background(),
					&auth.OAuthClaims{
						RegisteredClaims: jwt.RegisteredClaims{
							Subject: orchestratortest.MockUserId1,
							Issuer:  orchestratortest.MockUserIssuer1,
						},
					},
				),
			},
			fields: fields{
				authz: &AuthorizationStrategyPermissionStore{
					&fakePermissionStore{
						ids: []string{orchestratortest.MockToeId1, orchestratortest.MockToeId2},
						err: nil,
					},
				},
			},
			wantAllowed: func(t *testing.T, got bool, msgAndArgs ...any) bool {
				return assert.False(t, got)
			},
			wantResourceIds: func(t *testing.T, got []string, msgAndArgs ...any) bool {
				return assert.Equal(t, []string{orchestratortest.MockToeId1, orchestratortest.MockToeId2}, got)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := tt.fields.authz
			got, got2 := a.AllowedAuditScopes(tt.args.ctx)
			tt.wantAllowed(t, got)
			tt.wantResourceIds(t, got2)
		})
	}
}

// func TestAuthorizationStrategyAllowAll_CheckAccess(t *testing.T) {
// 	type fields struct {
// 		authz *AuthorizationStrategyAllowAll
// 	}
// 		{
// 			name: "happy path: allows all access",
// 			fields: fields{
// 				authz: &AuthorizationStrategyAllowAll{},
// 			},
// 			wantAllowed: func(t *testing.T, got bool, msgAndArgs ...any) bool {
// 				return assert.True(t, got)
// 			},
// 			wantResourceIds: func(t *testing.T, got []string, msgAndArgs ...any) bool {
// 				return assert.Nil(t, got)
// 			},
// 		},
// 				return assert.True(t, got)
// 			},
// 			wantResourceIds: func(t *testing.T, got []string, msgAndArgs ...any) bool {
// 				return assert.Nil(t, got)
// 			},
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			got, got2 := tt.fields.authz.CheckAccess(context.Background(), "", 0, 0, "", 0)
// 			assert.True(t, tt.wantAllowed(t, got))
// 			assert.True(t, tt.wantResourceIds(t, got2))
// 		})
// 	}
// }

// func TestAuthorizationStrategyAllowAll_AllowedTargetOfEvaluations(t *testing.T) {
// 		{
// 			name: "happy path: allows all access to ToEs",
// 			fields: fields{
// 				authz: &AuthorizationStrategyAllowAll{},
// 			},
// 			wantAllowed: func(t *testing.T, got bool, msgAndArgs ...any) bool {
// 				return assert.True(t, got)
// 			},
// 			wantResourceIds: func(t *testing.T, got []string, msgAndArgs ...any) bool {
// 				return assert.Nil(t, got)
// 			},
// 		},
// 		{
// 			name: "happy path: allows all access to ToEs",
// 			wantAllowed: func(t *testing.T, got bool, msgAndArgs ...any) bool {
// 				return assert.True(t, got)
// 			},
// 			wantResourceIds: func(t *testing.T, got []string, msgAndArgs ...any) bool {
// 				return assert.Nil(t, got)
// 			},
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			got, got2 := tt.fields.authz.AllowedTargetOfEvaluations(context.Background())
// 			tt.wantAllowed(t, got)
// 			tt.wantResourceIds(t, got2)
// 		})
// 	}
// 		{
// 			name: "happy path: allows all access to audit scopes",
// 			fields: fields{
// 				authz: &AuthorizationStrategyAllowAll{},
// 			},
// 			wantAllowed: func(t *testing.T, got bool, msgAndArgs ...any) bool {
// 				return assert.True(t, got)
// 			},
// 			wantResourceIds: func(t *testing.T, got []string, msgAndArgs ...any) bool {
// 				return assert.Nil(t, got)
// 			},
// 		},
// 		wantAllowed     assert.Want[bool]
// 		wantResourceIds assert.Want[[]string]
// 	}{
// 		{
// 			name: "happy path: allows all access to audit scopes",
// 			wantAllowed: func(t *testing.T, got bool, msgAndArgs ...any) bool {
// 				return assert.True(t, got)
// 			},
// 			wantResourceIds: func(t *testing.T, got []string, msgAndArgs ...any) bool {
// 				return assert.Nil(t, got)
// 			},
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			got, got2 := tt.fields.authz.AllowedAuditScopes(context.Background())
// 			tt.wantAllowed(t, got)
// 			tt.wantResourceIds(t, got2)
// 		})
// 	}
// }

func TestAuthorizationStrategyPermissionStore_AllowedTargetOfEvaluations(t *testing.T) {
	tests := []struct {
		name  string // description of this test case
		want  bool
		want2 []string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// TODO: construct the receiver type.
			var a AuthorizationStrategyPermissionStore
			got, got2 := a.AllowedTargetOfEvaluations(context.Background())
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("AllowedTargetOfEvaluations() = %v, want %v", got, tt.want)
			}
			if true {
				t.Errorf("AllowedTargetOfEvaluations() = %v, want %v", got2, tt.want2)
			}
		})
	}
}
