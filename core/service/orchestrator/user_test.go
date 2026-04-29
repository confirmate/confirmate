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

	"confirmate.io/core/api/orchestrator"
	"confirmate.io/core/auth"
	"confirmate.io/core/persistence"
	"confirmate.io/core/persistence/persistencetest"
	"confirmate.io/core/service"
	"confirmate.io/core/service/orchestrator/orchestratortest"
	"confirmate.io/core/util/assert"

	"connectrpc.com/connect"
	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/protobuf/types/known/emptypb"
)

func TestService_GetCurrentUser(t *testing.T) {
	type fields struct {
		db persistence.DB
	}
	tests := []struct {
		name    string
		ctx     context.Context
		fields  fields
		want    assert.Want[*connect.Response[orchestrator.User]]
		wantErr assert.WantErr
	}{
		{
			name: "err: unauthenticated - no claims",
			ctx:  context.Background(),
			want: assert.Nil[*connect.Response[orchestrator.User]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeUnauthenticated)
			},
		},
		{
			name: "err: not found - user not in DB",
			ctx: auth.WithClaims(context.Background(), &auth.OAuthClaims{
				RegisteredClaims: jwt.RegisteredClaims{
					Subject: "user-1",
					Issuer:  "test",
				},
			}),
			fields: fields{db: persistencetest.NewInMemoryDB(t, types, joinTables)},
			want:   assert.Nil[*connect.Response[orchestrator.User]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeNotFound)
			},
		},
		{
			name: "happy path",
			ctx: auth.WithClaims(context.Background(), &auth.OAuthClaims{
				RegisteredClaims: jwt.RegisteredClaims{
					Subject: "user-1",
					Issuer:  "test",
				},
			}),
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					assert.NoError(t, d.Create(&orchestrator.User{Id: "test|user-1", Enabled: true}))
				}),
			},
			want: func(t *testing.T, got *connect.Response[orchestrator.User], _ ...any) bool {
				return assert.NotNil(t, got) && assert.Equal(t, "test|user-1", got.Msg.Id)
			},
			wantErr: assert.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &Service{db: tt.fields.db}

			res, err := svc.GetCurrentUser(tt.ctx, connect.NewRequest(&orchestrator.GetCurrentUserRequest{}))
			assert.True(t, tt.wantErr(t, err))
			assert.True(t, tt.want(t, res))
		})
	}
}

func TestService_UpsertUserPermission(t *testing.T) {
	type args struct {
		ctx context.Context
		req *connect.Request[orchestrator.UpsertUserPermissionRequest]
	}
	type fields struct {
		db    persistence.DB
		authz service.AuthorizationStrategy
	}
	tests := []struct {
		name    string
		args    args
		fields  fields
		want    assert.Want[*connect.Response[orchestrator.UpsertUserPermissionResponse]]
		wantErr assert.WantErr
	}{
		{
			name: "err: invalid request",
			args: args{
				ctx: context.Background(),
				req: connect.NewRequest(&orchestrator.UpsertUserPermissionRequest{}),
			},
			want: assert.Nil[*connect.Response[orchestrator.UpsertUserPermissionResponse]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeInvalidArgument) &&
					assert.ErrorContains(t, err, "invalid request:")
			},
		},
		{
			name: "err: permission denied - non-admin",
			args: args{
				ctx: context.Background(),
				req: connect.NewRequest(&orchestrator.UpsertUserPermissionRequest{
					UserPermission: &orchestrator.UserPermission{
						UserId:       orchestratortest.MockUserId1,
						ResourceId:   orchestratortest.MockTargetOfEvaluation1.Id,
						ResourceType: orchestrator.ObjectType_OBJECT_TYPE_TARGET_OF_EVALUATION,
						Permission:   orchestrator.UserPermission_PERMISSION_READER,
					},
				}),
			},
			fields: fields{
				db:    persistencetest.NewInMemoryDB(t, types, joinTables),
				authz: &service.AuthorizationStrategyPermissionStore{},
			},
			want: assert.Nil[*connect.Response[orchestrator.UpsertUserPermissionResponse]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodePermissionDenied)
			},
		},
		{
			name: "happy path: with allow-all authorization strategy",
			args: args{
				req: connect.NewRequest(&orchestrator.UpsertUserPermissionRequest{
					UserPermission: &orchestrator.UserPermission{
						UserId:       orchestratortest.MockUserId1,
						ResourceId:   orchestratortest.MockTargetOfEvaluation1.Id,
						ResourceType: orchestrator.ObjectType_OBJECT_TYPE_TARGET_OF_EVALUATION,
						Permission:   orchestrator.UserPermission_PERMISSION_READER,
					},
				}),
			},
			fields: fields{
				db:    persistencetest.NewInMemoryDB(t, types, joinTables),
				authz: &service.AuthorizationStrategyAllowAll{},
			},
			want: func(t *testing.T, got *connect.Response[orchestrator.UpsertUserPermissionResponse], _ ...any) bool {
				return assert.NotNil(t, got)
			},
			wantErr: assert.NoError,
		},
		{
			name: "happy path: with authorization strategy with permission store and admin token",
			args: args{
				ctx: auth.WithClaims(
					context.Background(),
					&auth.OAuthClaims{
						IsAdminToken: true,
					},
				),
				req: connect.NewRequest(&orchestrator.UpsertUserPermissionRequest{
					UserPermission: &orchestrator.UserPermission{
						UserId:       orchestratortest.MockUserId1,
						ResourceId:   orchestratortest.MockTargetOfEvaluation1.Id,
						ResourceType: orchestrator.ObjectType_OBJECT_TYPE_TARGET_OF_EVALUATION,
						Permission:   orchestrator.UserPermission_PERMISSION_READER,
					},
				}),
			},
			fields: fields{
				db:    persistencetest.NewInMemoryDB(t, types, joinTables),
				authz: &service.AuthorizationStrategyPermissionStore{},
			},
			want: func(t *testing.T, got *connect.Response[orchestrator.UpsertUserPermissionResponse], _ ...any) bool {
				return assert.NotNil(t, got)
			},
			wantErr: assert.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &Service{db: tt.fields.db, authz: tt.fields.authz}

			res, err := svc.UpsertUserPermission(tt.args.ctx, tt.args.req)
			assert.True(t, tt.wantErr(t, err))
			assert.True(t, tt.want(t, res))
		})
	}
}

func TestService_ListUsers(t *testing.T) {
	type args struct {
		context context.Context
		req     *connect.Request[orchestrator.ListUsersRequest]
	}
	type fields struct {
		db    persistence.DB
		authz service.AuthorizationStrategy
	}
	tests := []struct {
		name    string
		args    args
		fields  fields
		want    assert.Want[*connect.Response[orchestrator.ListUsersResponse]]
		wantErr assert.WantErr
	}{
		{
			name: "err: database error",
			args: args{
				req: connect.NewRequest(&orchestrator.ListUsersRequest{}),
			},
			fields: fields{
				db:    persistencetest.ListErrorDB(t, persistence.ErrDatabase, types, joinTables),
				authz: &service.AuthorizationStrategyAllowAll{},
			},
			want: assert.Nil[*connect.Response[orchestrator.ListUsersResponse]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeInternal) &&
					assert.ErrorContains(t, err, "database error:")
			},
		},
		{
			name: "validation error",
			args: args{
				req: connect.NewRequest(&orchestrator.ListUsersRequest{PageToken: "!!!invalid-base64!!!"}),
			},
			fields: fields{
				db:    persistencetest.NewInMemoryDB(t, types, joinTables),
				authz: &service.AuthorizationStrategyAllowAll{},
			},
			want: assert.Nil[*connect.Response[orchestrator.ListUsersResponse]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeInvalidArgument) &&
					assert.ErrorContains(t, err, "invalid request:")
			},
		},
		{
			name: "authorization error",
			args: args{
				req: connect.NewRequest(&orchestrator.ListUsersRequest{PageSize: -1}),
			},
			fields: fields{
				db:    persistencetest.NewInMemoryDB(t, types, joinTables),
				authz: &denyAuthorizationStrategy{},
			},
			want: assert.Nil[*connect.Response[orchestrator.ListUsersResponse]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodePermissionDenied)
			},
		},
		{
			name: "happy path: with allow-all authorization strategy",
			args: args{
				req: connect.NewRequest(&orchestrator.ListUsersRequest{PageSize: -1}),
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					assert.NoError(t, d.Create(orchestratortest.MockUser1))
				}),
				authz: &service.AuthorizationStrategyAllowAll{},
			},
			want: func(t *testing.T, got *connect.Response[orchestrator.ListUsersResponse], _ ...any) bool {
				return assert.NotNil(t, got) &&
					assert.Equal(t, 1, len(got.Msg.Users)) &&
					assert.Equal(t, "test-issuer|00000000-0000-0000-0000-000000000001", got.Msg.Users[0].Id)
			},
			wantErr: assert.NoError,
		},
		{
			name: "happy path: with authorization strategy with permission store and admin token",
			args: args{
				req: connect.NewRequest(&orchestrator.ListUsersRequest{}),
				context: auth.WithClaims(context.Background(), &auth.OAuthClaims{
					IsAdminToken: true,
				}),
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					assert.NoError(t, d.Create(orchestratortest.MockUser1))
				}),
				authz: &service.AuthorizationStrategyPermissionStore{},
			},
			want: func(t *testing.T, got *connect.Response[orchestrator.ListUsersResponse], _ ...any) bool {
				return assert.NotNil(t, got) &&
					assert.Equal(t, 1, len(got.Msg.Users)) &&
					assert.Equal(t, "test-issuer|00000000-0000-0000-0000-000000000001", got.Msg.Users[0].Id)
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

			res, err := svc.ListUsers(tt.args.context, tt.args.req)
			assert.True(t, tt.wantErr(t, err))
			assert.True(t, tt.want(t, res))
		})
	}
}

func TestService_ListUserPermissions(t *testing.T) {
	type args struct {
		ctx    context.Context
		userId string
	}
	type fields struct {
		db    persistence.DB
		authz service.AuthorizationStrategy
	}
	tests := []struct {
		name    string
		args    args
		fields  fields
		want    assert.Want[*connect.Response[orchestrator.ListUserPermissionsResponse]]
		wantErr assert.WantErr
	}{
		{
			name: "err: invalid request - missing user id",
			args: args{ctx: context.Background()},
			want: assert.Nil[*connect.Response[orchestrator.ListUserPermissionsResponse]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeInvalidArgument) &&
					assert.ErrorContains(t, err, "invalid request: ") &&
					assert.ErrorContains(t, err, "id")
			},
		},
		{
			name: "err: database error",
			args: args{
				userId: "non-existent-user-id",
			},
			fields: fields{
				db:    persistencetest.ListErrorDB(t, persistence.ErrDatabase, types, joinTables),
				authz: &service.AuthorizationStrategyAllowAll{},
			},
			want: assert.Nil[*connect.Response[orchestrator.ListUserPermissionsResponse]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeInternal) &&
					assert.ErrorContains(t, err, "database error:")
			},
		},
		{
			name: "authorization error - deny strategy",
			args: args{
				userId: "any-user-id",
			},
			fields: fields{
				db:    persistencetest.NewInMemoryDB(t, types, joinTables),
				authz: &denyAuthorizationStrategy{},
			},
			want: assert.Nil[*connect.Response[orchestrator.ListUserPermissionsResponse]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodePermissionDenied)
			},
		},
		{
			name: "happy path: with allow-all authorization strategy",
			args: args{
				userId: orchestratortest.MockUserId1,
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					assert.NoError(t, d.Create(&orchestrator.UserPermission{
						UserId:       orchestratortest.MockUserId1,
						ResourceId:   orchestratortest.MockTargetOfEvaluation1.Id,
						ResourceType: orchestrator.ObjectType_OBJECT_TYPE_TARGET_OF_EVALUATION,
						Permission:   orchestrator.UserPermission_PERMISSION_READER,
					}))
				}),
				authz: &service.AuthorizationStrategyAllowAll{},
			},
			want: func(t *testing.T, got *connect.Response[orchestrator.ListUserPermissionsResponse], _ ...any) bool {
				return assert.NotNil(t, got) &&
					assert.Equal(t, 1, len(got.Msg.UserPermissions)) &&
					assert.Equal(t, orchestratortest.MockUserId1, got.Msg.UserPermissions[0].UserId)
			},
			wantErr: assert.NoError,
		},
		{
			name: "happy path: with authorization strategy with permission store and admin token",
			args: args{
				ctx:    auth.WithClaims(context.Background(), &auth.OAuthClaims{IsAdminToken: true}),
				userId: orchestratortest.MockUserId1,
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					assert.NoError(t, d.Create(&orchestrator.UserPermission{
						UserId:       orchestratortest.MockUserId1,
						ResourceId:   orchestratortest.MockTargetOfEvaluation1.Id,
						ResourceType: orchestrator.ObjectType_OBJECT_TYPE_TARGET_OF_EVALUATION,
						Permission:   orchestrator.UserPermission_PERMISSION_READER,
					}))
				}),
				authz: &service.AuthorizationStrategyPermissionStore{},
			},
			want: func(t *testing.T, got *connect.Response[orchestrator.ListUserPermissionsResponse], _ ...any) bool {
				return assert.NotNil(t, got) &&
					assert.Equal(t, 1, len(got.Msg.UserPermissions)) &&
					assert.Equal(t, orchestratortest.MockUserId1, got.Msg.UserPermissions[0].UserId)
			},
			wantErr: assert.NoError,
		},
		{
			name: "happy path: userId filter excludes other users' permissions",
			args: args{
				ctx:    auth.WithClaims(context.Background(), &auth.OAuthClaims{IsAdminToken: true}),
				userId: orchestratortest.MockUserId1,
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					assert.NoError(t, d.Create(&orchestrator.UserPermission{
						UserId:       orchestratortest.MockUserId1,
						ResourceId:   orchestratortest.MockTargetOfEvaluation1.Id,
						ResourceType: orchestrator.ObjectType_OBJECT_TYPE_TARGET_OF_EVALUATION,
						Permission:   orchestrator.UserPermission_PERMISSION_READER,
					}))
					assert.NoError(t, d.Create(&orchestrator.UserPermission{
						UserId:       "other-user",
						ResourceId:   orchestratortest.MockTargetOfEvaluation1.Id,
						ResourceType: orchestrator.ObjectType_OBJECT_TYPE_TARGET_OF_EVALUATION,
						Permission:   orchestrator.UserPermission_PERMISSION_ADMIN,
					}))
				}),
				authz: &service.AuthorizationStrategyAllowAll{},
			},
			want: func(t *testing.T, got *connect.Response[orchestrator.ListUserPermissionsResponse], _ ...any) bool {
				return assert.NotNil(t, got) &&
					assert.Equal(t, 1, len(got.Msg.UserPermissions)) &&
					assert.Equal(t, orchestratortest.MockUserId1, got.Msg.UserPermissions[0].UserId)
			},
			wantErr: assert.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &Service{db: tt.fields.db, authz: tt.fields.authz}

			res, err := svc.ListUserPermissions(tt.args.ctx, connect.NewRequest(&orchestrator.ListUserPermissionsRequest{UserId: tt.args.userId}))
			assert.True(t, tt.wantErr(t, err))
			assert.True(t, tt.want(t, res))
		})
	}
}

func TestService_ListUserRoles(t *testing.T) {
	tests := []struct {
		name    string
		want    assert.Want[*connect.Response[orchestrator.ListUserRolesResponse]]
		wantErr assert.WantErr
	}{
		{
			name: "happy path",
			want: func(t *testing.T, got *connect.Response[orchestrator.ListUserRolesResponse], _ ...any) bool {
				// Role_name includes ROLE_UNSPECIFIED, so subtract 1
				return assert.NotNil(t, got) &&
					assert.Equal(t, len(orchestrator.Role_name)-1, len(got.Msg.Roles))
			},
			wantErr: assert.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &Service{}

			res, err := svc.ListUserRoles(context.Background(), connect.NewRequest(&orchestrator.ListUserRolesRequest{}))
			assert.True(t, tt.wantErr(t, err))
			assert.True(t, tt.want(t, res))
		})
	}
}

func TestService_RemoveUser(t *testing.T) {
	type args struct {
		ctx context.Context
		req *connect.Request[orchestrator.RemoveUserRequest]
	}
	type fields struct {
		db    persistence.DB
		authz service.AuthorizationStrategy
	}
	tests := []struct {
		name    string
		args    args
		fields  fields
		want    assert.Want[*connect.Response[emptypb.Empty]]
		wantErr assert.WantErr
	}{
		{
			name: "err: invalid request - missing user id",
			args: args{ctx: context.Background(), req: connect.NewRequest(&orchestrator.RemoveUserRequest{})},
			want: assert.Nil[*connect.Response[emptypb.Empty]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeInvalidArgument)
			},
		},
		{
			name: "err: permission denied - non-admin",
			args: args{
				ctx: context.Background(),
				req: connect.NewRequest(&orchestrator.RemoveUserRequest{UserId: orchestratortest.MockUserId1}),
			},
			fields: fields{
				db:    persistencetest.NewInMemoryDB(t, types, joinTables),
				authz: &service.AuthorizationStrategyPermissionStore{},
			},
			want: assert.Nil[*connect.Response[emptypb.Empty]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodePermissionDenied)
			},
		},
		{
			name: "err: db error getting user",
			args: args{
				ctx: context.Background(),
				req: connect.NewRequest(&orchestrator.RemoveUserRequest{UserId: orchestratortest.MockUser1.GetId()}),
			},
			fields: fields{
				db: persistencetest.GetErrorDB(t, persistence.ErrDatabase, types, joinTables, func(d persistence.DB) {
					assert.NoError(t, d.Create(orchestratortest.MockUser1))
				}),
				authz: &service.AuthorizationStrategyAllowAll{},
			},
			want: assert.Nil[*connect.Response[emptypb.Empty]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeInternal) &&
					assert.ErrorContains(t, err, "database error:")
			},
		},
		{
			name: "err: db error saving user",
			args: args{
				ctx: context.Background(),
				req: connect.NewRequest(&orchestrator.RemoveUserRequest{UserId: orchestratortest.MockUser1.GetId()}),
			},
			fields: fields{
				db: persistencetest.SaveErrorDB(t, persistence.ErrDatabase, types, joinTables, func(d persistence.DB) {
					assert.NoError(t, d.Create(orchestratortest.MockUser1))
				}),
				authz: &service.AuthorizationStrategyAllowAll{},
			},
			want: assert.Nil[*connect.Response[emptypb.Empty]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeInternal) &&
					assert.ErrorContains(t, err, "database error:")
			},
		},
		{
			name: "happy path: with allow-all authorization strategy",
			args: args{
				ctx: context.Background(),
				req: connect.NewRequest(&orchestrator.RemoveUserRequest{UserId: orchestratortest.MockUser1.GetId()}),
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					assert.NoError(t, d.Create(orchestratortest.MockUser1))
				}),
				authz: &service.AuthorizationStrategyAllowAll{},
			},
			want: func(t *testing.T, got *connect.Response[emptypb.Empty], _ ...any) bool {
				return assert.NotNil(t, got)
			},
			wantErr: assert.NoError,
		},
		{
			name: "happy path: with authorization strategy with permission store and admin token",
			args: args{
				ctx: auth.WithClaims(context.Background(), &auth.OAuthClaims{IsAdminToken: true}),
				req: connect.NewRequest(&orchestrator.RemoveUserRequest{UserId: orchestratortest.MockUser1.GetId()}),
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					assert.NoError(t, d.Create(orchestratortest.MockUser1))
					assert.NoError(t, d.Create(orchestratortest.MockUser2))
				}),
				authz: &service.AuthorizationStrategyPermissionStore{},
			},
			want: func(t *testing.T, got *connect.Response[emptypb.Empty], _ ...any) bool {
				return assert.NotNil(t, got)
			},
			wantErr: assert.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &Service{db: tt.fields.db, authz: tt.fields.authz}

			res, err := svc.RemoveUser(tt.args.ctx, tt.args.req)
			assert.True(t, tt.wantErr(t, err))
			assert.True(t, tt.want(t, res))
		})
	}
}

func TestService_GetUser(t *testing.T) {
	type args struct {
		ctx context.Context
		req *connect.Request[orchestrator.GetUserRequest]
	}
	type fields struct {
		db    persistence.DB
		authz service.AuthorizationStrategy
	}
	tests := []struct {
		name    string
		args    args
		fields  fields
		want    assert.Want[*connect.Response[orchestrator.User]]
		wantErr assert.WantErr
	}{
		{
			name: "err: invalid request - missing user id",
			args: args{
				ctx: context.Background(),
				req: connect.NewRequest(&orchestrator.GetUserRequest{}),
			},
			want: assert.Nil[*connect.Response[orchestrator.User]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeInvalidArgument) &&
					assert.ErrorContains(t, err, "invalid request: ") &&
					assert.ErrorContains(t, err, "id")
			},
		},
		{
			name: "err: permission denied - deny strategy",
			args: args{

				req: connect.NewRequest(&orchestrator.GetUserRequest{
					UserId: orchestratortest.MockUserId1,
				}),
			},
			fields: fields{
				db:    persistencetest.NewInMemoryDB(t, types, joinTables),
				authz: &denyAuthorizationStrategy{},
			},
			want: assert.Nil[*connect.Response[orchestrator.User]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodePermissionDenied)
			},
		},
		{
			name: "err: database error getting user",
			args: args{
				req: connect.NewRequest(&orchestrator.GetUserRequest{
					UserId: orchestratortest.MockUser1.GetId(),
				}),
			},
			fields: fields{
				db: persistencetest.GetErrorDB(t, persistence.ErrDatabase, types, joinTables, func(d persistence.DB) {
					assert.NoError(t, d.Create(orchestratortest.MockUser1))
				}),
				authz: &service.AuthorizationStrategyAllowAll{},
			},
			want: assert.Nil[*connect.Response[orchestrator.User]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeInternal) &&
					assert.ErrorContains(t, err, "database error:")
			},
		},
		{
			name: "happy path: with allow-all authorization strategy",
			args: args{
				req: connect.NewRequest(&orchestrator.GetUserRequest{
					UserId: orchestratortest.MockUser1.GetId(),
				}),
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					err := d.Create(orchestratortest.MockUser1)
					assert.NoError(t, err)
					err = d.Create(orchestratortest.MockUser2)
					assert.NoError(t, err)
				}),
				authz: &service.AuthorizationStrategyAllowAll{},
			},
			want: func(t *testing.T, got *connect.Response[orchestrator.User], _ ...any) bool {
				return assert.Equal(t, orchestratortest.MockUser1, got.Msg)
			},
			wantErr: assert.NoError,
		},
		{
			name: "happy path: with authorization strategy with permission store and admin token",
			args: args{
				req: connect.NewRequest(&orchestrator.GetUserRequest{
					UserId: orchestratortest.MockUser1.GetId(),
				}),
				ctx: auth.WithClaims(context.Background(), &auth.OAuthClaims{IsAdminToken: true}),
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					err := d.Create(orchestratortest.MockUser1)
					assert.NoError(t, err)
					err = d.Create(orchestratortest.MockUser2)
					assert.NoError(t, err)
				}),
				authz: &service.AuthorizationStrategyPermissionStore{},
			},
			want: func(t *testing.T, got *connect.Response[orchestrator.User], _ ...any) bool {
				return assert.Equal(t, orchestratortest.MockUser1, got.Msg)
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

			res, err := svc.GetUser(tt.args.ctx, tt.args.req)
			assert.True(t, tt.wantErr(t, err))
			assert.True(t, tt.want(t, res))
		})
	}
}
