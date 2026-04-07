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
	"testing"

	"confirmate.io/core/api/orchestrator"
	"confirmate.io/core/auth"
	"confirmate.io/core/util/assert"

	"github.com/golang-jwt/jwt/v5"
)

type denyAuthorizationStrategy struct{}

func (*denyAuthorizationStrategy) CheckAccess(_ context.Context, _ string, _ orchestrator.RequestType, _ orchestrator.UserPermission_Permission, _ string, _ orchestrator.ObjectType) (bool, []string) {
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
		strategy *AuthorizationStrategyJWT
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
			fields: fields{strategy: &AuthorizationStrategyJWT{
				AllowAllKey: DefaultAllowAllClaim,
			}},
			want: func(t *testing.T, got bool, _ ...any) bool {
				return assert.False(t, got)
			},
			wantErr: assert.NoError,
		},
		{
			name: "allows when allow-all claim is true",
			args: args{
				ctx:            auth.WithClaims(context.Background(), jwt.MapClaims{DefaultAllowAllClaim: true}),
				userId:         "user-1",
				reqType:        orchestrator.RequestType_REQUEST_TYPE_GET,
				userPermission: orchestrator.UserPermission_PERMISSION_READER,
				resourceId:     "any",
				objectType:     orchestrator.ObjectType_OBJECT_TYPE_TARGET_OF_EVALUATION,
			},
			fields: fields{strategy: &AuthorizationStrategyJWT{
				AllowAllKey: DefaultAllowAllClaim,
			}},
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
			fields: fields{strategy: &AuthorizationStrategyJWT{
				AllowAllKey: DefaultAllowAllClaim,
			}},
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
