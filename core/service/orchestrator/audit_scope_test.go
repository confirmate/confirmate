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
	"confirmate.io/core/persistence/persistencetest"
	"confirmate.io/core/util/assert"

	"connectrpc.com/connect"
)

func TestService_CreateAuditScope(t *testing.T) {
	var (
		tests = []struct {
			name    string
			req     *orchestrator.CreateAuditScopeRequest
			wantErr bool
		}{
			{
				name: "happy path",
				req: &orchestrator.CreateAuditScopeRequest{
					AuditScope: &orchestrator.AuditScope{
						TargetOfEvaluationId: "toe-1",
						CatalogId:            "catalog-1",
					},
				},
				wantErr: false,
			},
			{
				name: "with existing ID",
				req: &orchestrator.CreateAuditScopeRequest{
					AuditScope: &orchestrator.AuditScope{
						Id:                   "existing-id",
						TargetOfEvaluationId: "toe-2",
						CatalogId:            "catalog-2",
					},
				},
				wantErr: false,
			},
		}
	)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				svc = &service{
					db: persistencetest.NewInMemoryDB(t, types, joinTables),
				}
				req = connect.NewRequest(tt.req)
			)

			res, err := svc.CreateAuditScope(context.Background(), req)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, res)
			assert.NotEmpty(t, res.Msg.Id)
		})
	}
}

func TestService_GetAuditScope(t *testing.T) {
	var (
		tests = []struct {
			name    string
			id      string
			setup   func(*service)
			wantErr bool
		}{
			{
				name: "happy path",
				id:   "scope-1",
				setup: func(svc *service) {
					err := svc.db.Create(&orchestrator.AuditScope{
						Id:                   "scope-1",
						TargetOfEvaluationId: "toe-1",
						CatalogId:            "catalog-1",
					})
					assert.NoError(t, err)
				},
				wantErr: false,
			},
			{
				name:    "not found",
				id:      "non-existent",
				setup:   func(svc *service) {},
				wantErr: true,
			},
		}
	)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				svc = &service{
					db: persistencetest.NewInMemoryDB(t, types, joinTables),
				}
				req = connect.NewRequest(&orchestrator.GetAuditScopeRequest{
					AuditScopeId: tt.id,
				})
			)

			tt.setup(svc)

			res, err := svc.GetAuditScope(context.Background(), req)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, res)
			assert.Equal(t, tt.id, res.Msg.Id)
		})
	}
}

func TestService_ListAuditScopes(t *testing.T) {
	var (
		tests = []struct {
			name      string
			filter    *orchestrator.ListAuditScopesRequest_Filter
			setup     func(*service)
			wantCount int
		}{
			{
				name:   "list all",
				filter: nil,
				setup: func(svc *service) {
					err := svc.db.Create(&orchestrator.AuditScope{
						Id:                   "scope-1",
						TargetOfEvaluationId: "toe-1",
						CatalogId:            "catalog-1",
					})
					assert.NoError(t, err)

					err = svc.db.Create(&orchestrator.AuditScope{
						Id:                   "scope-2",
						TargetOfEvaluationId: "toe-2",
						CatalogId:            "catalog-2",
					})
					assert.NoError(t, err)
				},
				wantCount: 2,
			},
			{
				name: "filter by target of evaluation",
				filter: &orchestrator.ListAuditScopesRequest_Filter{
					TargetOfEvaluationId: stringPtr("toe-1"),
				},
				setup: func(svc *service) {
					err := svc.db.Create(&orchestrator.AuditScope{
						Id:                   "scope-1",
						TargetOfEvaluationId: "toe-1",
						CatalogId:            "catalog-1",
					})
					assert.NoError(t, err)

					err = svc.db.Create(&orchestrator.AuditScope{
						Id:                   "scope-2",
						TargetOfEvaluationId: "toe-2",
						CatalogId:            "catalog-2",
					})
					assert.NoError(t, err)
				},
				wantCount: 1,
			},
			{
				name: "filter by catalog",
				filter: &orchestrator.ListAuditScopesRequest_Filter{
					CatalogId: stringPtr("catalog-1"),
				},
				setup: func(svc *service) {
					err := svc.db.Create(&orchestrator.AuditScope{
						Id:                   "scope-1",
						TargetOfEvaluationId: "toe-1",
						CatalogId:            "catalog-1",
					})
					assert.NoError(t, err)

					err = svc.db.Create(&orchestrator.AuditScope{
						Id:                   "scope-2",
						TargetOfEvaluationId: "toe-2",
						CatalogId:            "catalog-2",
					})
					assert.NoError(t, err)
				},
				wantCount: 1,
			},
		}
	)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				svc = &service{
					db: persistencetest.NewInMemoryDB(t, types, joinTables),
				}
				req = connect.NewRequest(&orchestrator.ListAuditScopesRequest{
					Filter: tt.filter,
				})
			)

			tt.setup(svc)

			res, err := svc.ListAuditScopes(context.Background(), req)

			assert.NoError(t, err)
			assert.NotNil(t, res)
			assert.Equal(t, tt.wantCount, len(res.Msg.AuditScopes))
		})
	}
}

func TestService_UpdateAuditScope(t *testing.T) {
	var (
		tests = []struct {
			name    string
			req     *orchestrator.UpdateAuditScopeRequest
			setup   func(*service)
			wantErr bool
		}{
			{
				name: "happy path",
				req: &orchestrator.UpdateAuditScopeRequest{
					AuditScope: &orchestrator.AuditScope{
						Id:                   "scope-1",
						TargetOfEvaluationId: "toe-1-updated",
						CatalogId:            "catalog-1-updated",
					},
				},
				setup: func(svc *service) {
					err := svc.db.Create(&orchestrator.AuditScope{
						Id:                   "scope-1",
						TargetOfEvaluationId: "toe-1",
						CatalogId:            "catalog-1",
					})
					assert.NoError(t, err)
				},
				wantErr: false,
			},
			{
				name: "not found",
				req: &orchestrator.UpdateAuditScopeRequest{
					AuditScope: &orchestrator.AuditScope{
						Id:                   "non-existent",
						TargetOfEvaluationId: "toe-1",
						CatalogId:            "catalog-1",
					},
				},
				setup:   func(svc *service) {},
				wantErr: true,
			},
		}
	)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				svc = &service{
					db: persistencetest.NewInMemoryDB(t, types, joinTables),
				}
				req = connect.NewRequest(tt.req)
			)

			tt.setup(svc)

			res, err := svc.UpdateAuditScope(context.Background(), req)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, res)
			assert.Equal(t, tt.req.AuditScope.Id, res.Msg.Id)
		})
	}
}

func TestService_RemoveAuditScope(t *testing.T) {
	var (
		tests = []struct {
			name    string
			id      string
			setup   func(*service)
			wantErr bool
		}{
			{
				name: "happy path",
				id:   "scope-1",
				setup: func(svc *service) {
					err := svc.db.Create(&orchestrator.AuditScope{
						Id:                   "scope-1",
						TargetOfEvaluationId: "toe-1",
						CatalogId:            "catalog-1",
					})
					assert.NoError(t, err)
				},
				wantErr: false,
			},
		}
	)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				svc = &service{
					db: persistencetest.NewInMemoryDB(t, types, joinTables),
				}
				req = connect.NewRequest(&orchestrator.RemoveAuditScopeRequest{
					AuditScopeId: tt.id,
				})
			)

			tt.setup(svc)

			res, err := svc.RemoveAuditScope(context.Background(), req)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, res)
		})
	}
}
