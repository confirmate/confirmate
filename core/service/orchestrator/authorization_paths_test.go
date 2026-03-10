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

	api "confirmate.io/core/api"
	"confirmate.io/core/api/assessment"
	"confirmate.io/core/api/orchestrator"
	"confirmate.io/core/persistence"
	"confirmate.io/core/persistence/persistencetest"
	"confirmate.io/core/service/orchestrator/orchestratortest"
	"confirmate.io/core/util/assert"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/structpb"
)

type denyAuthorizationStrategy struct{}

func (*denyAuthorizationStrategy) CheckAccess(context.Context, orchestrator.RequestType, api.HasTargetOfEvaluationId) bool {
	return false
}

func (*denyAuthorizationStrategy) AllowedTargetOfEvaluations(context.Context) (bool, []string) {
	return false, nil
}

func TestService_AuthzDeniedBranches(t *testing.T) {
	type args struct {
		call func(*Service) (any, error)
	}
	type fields struct {
		db persistence.DB
	}

	tests := []struct {
		name    string
		args    args
		fields  fields
		wantErr assert.WantErr
	}{
		{
			name: "create audit scope returns permission denied",
			args: args{call: func(s *Service) (any, error) {
				return s.CreateAuditScope(context.Background(), connect.NewRequest(&orchestrator.CreateAuditScopeRequest{
					AuditScope: &orchestrator.AuditScope{
						TargetOfEvaluationId: orchestratortest.MockAuditScope1.TargetOfEvaluationId,
						CatalogId:            orchestratortest.MockAuditScope1.CatalogId,
					},
				}))
			}},
			fields: fields{db: persistencetest.NewInMemoryDB(t, types, joinTables)},
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodePermissionDenied)
			},
		},
		{
			name: "create certificate returns permission denied",
			args: args{call: func(s *Service) (any, error) {
				return s.CreateCertificate(context.Background(), connect.NewRequest(&orchestrator.CreateCertificateRequest{
					Certificate: &orchestrator.Certificate{
						Id:                   orchestratortest.MockCertificate1.Id,
						Name:                 orchestratortest.MockCertificate1.Name,
						TargetOfEvaluationId: orchestratortest.MockCertificate1.TargetOfEvaluationId,
					},
				}))
			}},
			fields: fields{db: persistencetest.NewInMemoryDB(t, types, joinTables)},
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodePermissionDenied)
			},
		},
		{
			name: "get target of evaluation returns permission denied",
			args: args{call: func(s *Service) (any, error) {
				return s.GetTargetOfEvaluation(context.Background(), connect.NewRequest(&orchestrator.GetTargetOfEvaluationRequest{
					TargetOfEvaluationId: orchestratortest.MockToeId1,
				}))
			}},
			fields: fields{db: persistencetest.NewInMemoryDB(t, types, joinTables)},
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodePermissionDenied)
			},
		},
		{
			name: "update metric configuration returns permission denied",
			args: args{call: func(s *Service) (any, error) {
				return s.UpdateMetricConfiguration(context.Background(), connect.NewRequest(&orchestrator.UpdateMetricConfigurationRequest{
					Configuration: &assessment.MetricConfiguration{
						Operator:             "==",
						TargetValue:          structpb.NewBoolValue(true),
						MetricId:             orchestratortest.MockMetric1.Id,
						TargetOfEvaluationId: orchestratortest.MockToeId1,
					},
				}))
			}},
			fields: fields{db: persistencetest.NewInMemoryDB(t, types, joinTables)},
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodePermissionDenied)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &Service{
				db:    tt.fields.db,
				authz: &denyAuthorizationStrategy{},
			}
			_, err := tt.args.call(svc)
			tt.wantErr(t, err)
		})
	}
}
