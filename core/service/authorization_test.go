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

	"confirmate.io/core/api"
	"confirmate.io/core/api/orchestrator"
	"confirmate.io/core/auth"
	"confirmate.io/core/util/assert"

	"connectrpc.com/connect"
	"github.com/golang-jwt/jwt/v5"
)

type toeReq struct {
	targetID string
}

func (r *toeReq) GetTargetOfEvaluationId() string {
	return r.targetID
}

type denyAuthorizationStrategy struct{}

func (*denyAuthorizationStrategy) CheckAccess(context.Context, orchestrator.RequestType, api.HasTargetOfEvaluationId) bool {
	return false
}

func (*denyAuthorizationStrategy) AllowedTargetOfEvaluations(context.Context) (bool, []string) {
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
			got := CheckAccess(tt.authz, context.Background(), orchestrator.RequestType_REQUEST_TYPE_UPDATED, connect.NewRequest(&toeReq{targetID: "toe-1"}))
			tt.want(t, got)
		})
	}
}

func TestCheckAccess_AutoTargetResolution(t *testing.T) {
	strategy := &AuthorizationStrategyJWT{
		TargetOfEvaluationsKey: DefaultTargetOfEvaluationsClaim,
		AllowAllKey:            DefaultAllowAllClaim,
	}

	tests := []struct {
		name string
		call func(context.Context, *AuthorizationStrategyJWT) bool
		want assert.Want[bool]
	}{
		{
			name: "direct request target id",
			call: func(ctx context.Context, strategy *AuthorizationStrategyJWT) bool {
				return CheckAccess(strategy, ctx, orchestrator.RequestType_REQUEST_TYPE_UPDATED, connect.NewRequest(&toeReq{targetID: "toe-1"}))
			},
			want: func(t *testing.T, got bool, _ ...any) bool {
				return assert.True(t, got)
			},
		},
		{
			name: "payload target id",
			call: func(ctx context.Context, strategy *AuthorizationStrategyJWT) bool {
				return CheckAccess(strategy, ctx, orchestrator.RequestType_REQUEST_TYPE_UPDATED, connect.NewRequest(&orchestrator.CreateAuditScopeRequest{
					AuditScope: &orchestrator.AuditScope{TargetOfEvaluationId: "toe-1"},
				}))
			},
			want: func(t *testing.T, got bool, _ ...any) bool {
				return assert.True(t, got)
			},
		},
		{
			name: "payload without target id",
			call: func(ctx context.Context, strategy *AuthorizationStrategyJWT) bool {
				return CheckAccess(strategy, ctx, orchestrator.RequestType_REQUEST_TYPE_UPDATED, connect.NewRequest(&orchestrator.CreateCatalogRequest{
					Catalog: &orchestrator.Catalog{Id: "catalog-1"},
				}))
			},
			want: func(t *testing.T, got bool, _ ...any) bool {
				return assert.False(t, got)
			},
		},
	}

	ctx := auth.WithClaims(context.Background(), jwt.MapClaims{DefaultTargetOfEvaluationsClaim: []any{"toe-1"}})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.call(ctx, strategy)
			tt.want(t, got)
		})
	}
}

func TestAuthorizationStrategyJWT_AllowedTargetOfEvaluations(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	type fields struct {
		strategy *AuthorizationStrategyJWT
	}
	type want struct {
		all bool
		ids []string
	}

	tests := []struct {
		name    string
		args    args
		fields  fields
		want    assert.Want[want]
		wantErr assert.WantErr
	}{
		{
			name: "allow-all claim returns all true",
			args: args{ctx: auth.WithClaims(context.Background(), jwt.MapClaims{
				DefaultAllowAllClaim: true,
			})},
			fields: fields{strategy: &AuthorizationStrategyJWT{
				TargetOfEvaluationsKey: DefaultTargetOfEvaluationsClaim,
				AllowAllKey:            DefaultAllowAllClaim,
			}},
			want: func(t *testing.T, got want, _ ...any) bool {
				return assert.True(t, got.all) && assert.Nil(t, got.ids)
			},
			wantErr: assert.NoError,
		},
		{
			name: "target claim list parsed from generic slice",
			args: args{ctx: auth.WithClaims(context.Background(), jwt.MapClaims{
				DefaultTargetOfEvaluationsClaim: []any{"toe-1", "toe-2"},
			})},
			fields: fields{strategy: &AuthorizationStrategyJWT{
				TargetOfEvaluationsKey: DefaultTargetOfEvaluationsClaim,
				AllowAllKey:            DefaultAllowAllClaim,
			}},
			want: func(t *testing.T, got want, _ ...any) bool {
				return assert.False(t, got.all) && assert.Equal(t, []string{"toe-1", "toe-2"}, got.ids)
			},
			wantErr: assert.NoError,
		},
		{
			name: "missing claims returns no access",
			args: args{ctx: context.Background()},
			fields: fields{strategy: &AuthorizationStrategyJWT{
				TargetOfEvaluationsKey: DefaultTargetOfEvaluationsClaim,
				AllowAllKey:            DefaultAllowAllClaim,
			}},
			want: func(t *testing.T, got want, _ ...any) bool {
				return assert.False(t, got.all) && assert.Nil(t, got.ids)
			},
			wantErr: assert.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			all, ids := tt.fields.strategy.AllowedTargetOfEvaluations(tt.args.ctx)
			got := want{all: all, ids: ids}
			assert.True(t, tt.wantErr(t, nil))
			assert.True(t, tt.want(t, got))
		})
	}
}

func TestAuthorizationStrategyJWT_CheckAccess(t *testing.T) {
	type args struct {
		ctx context.Context
		req *toeReq
	}
	type fields struct {
		strategy *AuthorizationStrategyJWT
	}

	strategy := &AuthorizationStrategyJWT{
		TargetOfEvaluationsKey: DefaultTargetOfEvaluationsClaim,
		AllowAllKey:            DefaultAllowAllClaim,
	}

	tests := []struct {
		name    string
		args    args
		fields  fields
		want    assert.Want[bool]
		wantErr assert.WantErr
	}{
		{
			name: "allows matching target id",
			args: args{
				ctx: auth.WithClaims(context.Background(), jwt.MapClaims{DefaultTargetOfEvaluationsClaim: []any{"toe-1"}}),
				req: &toeReq{targetID: "toe-1"},
			},
			fields: fields{strategy: strategy},
			want: func(t *testing.T, got bool, _ ...any) bool {
				return assert.True(t, got)
			},
			wantErr: assert.NoError,
		},
		{
			name: "denies non-matching target id",
			args: args{
				ctx: auth.WithClaims(context.Background(), jwt.MapClaims{DefaultTargetOfEvaluationsClaim: []any{"toe-1"}}),
				req: &toeReq{targetID: "toe-2"},
			},
			fields: fields{strategy: strategy},
			want: func(t *testing.T, got bool, _ ...any) bool {
				return assert.False(t, got)
			},
			wantErr: assert.NoError,
		},
		{
			name: "allows when allow-all claim is true",
			args: args{
				ctx: auth.WithClaims(context.Background(), jwt.MapClaims{DefaultAllowAllClaim: true}),
				req: &toeReq{targetID: "any"},
			},
			fields: fields{strategy: strategy},
			want: func(t *testing.T, got bool, _ ...any) bool {
				return assert.True(t, got)
			},
			wantErr: assert.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.fields.strategy.CheckAccess(tt.args.ctx, 0, tt.args.req)
			assert.True(t, tt.wantErr(t, nil))
			assert.True(t, tt.want(t, got))
		})
	}
}
