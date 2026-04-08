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
				return assert.IsConnectError(t, err, connect.CodeInvalidArgument)
			},
		},
		{
			name: "err: permission denied - deny strategy",
			args: args{
				ctx: auth.WithClaims(context.Background(), &auth.OAuthClaims{
					RegisteredClaims: jwt.RegisteredClaims{
						Subject: "caller",
						Issuer:  "test",
					},
				}),
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
			name: "happy path",
			args: args{
				ctx: auth.WithClaims(context.Background(), &auth.OAuthClaims{
					RegisteredClaims: jwt.RegisteredClaims{
						Subject: "caller",
						Issuer:  "test",
					},
					IsAdminToken: true,
				}),
				req: connect.NewRequest(&orchestrator.GetUserRequest{
					UserId: orchestratortest.MockUserId1,
				}),
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					err := d.Create(orchestratortest.MockUser1)
					assert.NoError(t, err)
				}),
				authz: &service.AuthorizationStrategyPermissionStore{},
			},
			want: func(t *testing.T, got *connect.Response[orchestrator.User], _ ...any) bool {
				return assert.Equal(t, orchestratortest.MockUserId1, got.Msg.Id)
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
