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
	"confirmate.io/core/persistence"
	"confirmate.io/core/persistence/persistencetest"
	"confirmate.io/core/service"
	"confirmate.io/core/util/assert"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestService_GetAuditScopeStatistics(t *testing.T) {
	const (
		catalogId = "00000000-0000-0000-0009-000000000010"
		toeId     = "00000000-0000-0000-0000-000000000098"
		scopeId   = "00000000-0000-0000-0009-000000000011"
		ctrl1Id   = "00000000-0000-0000-000a-000000000011"
		ctrl1sub  = "00000000-0000-0000-000a-000000000012"
		ctrl2Id   = "00000000-0000-0000-000a-000000000013"
	)

	newDB := func(t *testing.T) persistence.DB {
		return persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
			assert.NoError(t, d.Create(&orchestrator.Catalog{Id: catalogId, Name: "Test Catalog"}))
			assert.NoError(t, d.Create(&orchestrator.Control{Id: ctrl1Id, ShortName: "C-1", Name: "Control 1", CatalogId: catalogId}))
			assert.NoError(t, d.Create(&orchestrator.Control{Id: ctrl1sub, ShortName: "C-1.1", Name: "Sub-control 1.1", CatalogId: catalogId, ParentControlId: new(ctrl1Id)}))
			assert.NoError(t, d.Create(&orchestrator.Control{Id: ctrl2Id, ShortName: "C-2", Name: "Control 2", CatalogId: catalogId}))
			assert.NoError(t, d.Create(&orchestrator.AuditScope{
				Id:                   scopeId,
				Name:                 "Test Scope",
				TargetOfEvaluationId: toeId,
				CatalogId:            catalogId,
				Status:               orchestrator.AuditScopeStatus_AUDIT_SCOPE_STATUS_SETUP,
			}))

			// Workflow status: only top-level controls (ctrl1, ctrl2) should be counted. The
			// sub-control's state (OPEN) must not leak into the aggregation.
			assert.NoError(t, d.Create(&orchestrator.ControlInScope{
				Id: "00000000-0000-0000-000b-000000000001", AuditScopeId: scopeId, TargetOfEvaluationId: toeId,
				ControlId: ctrl1Id, State: orchestrator.ControlInScopeState_CONTROL_IN_SCOPE_STATE_IMPLEMENTED,
			}))
			assert.NoError(t, d.Create(&orchestrator.ControlInScope{
				Id: "00000000-0000-0000-000b-000000000002", AuditScopeId: scopeId, TargetOfEvaluationId: toeId,
				ControlId: ctrl1sub, State: orchestrator.ControlInScopeState_CONTROL_IN_SCOPE_STATE_OPEN,
			}))
			assert.NoError(t, d.Create(&orchestrator.ControlInScope{
				Id: "00000000-0000-0000-000b-000000000003", AuditScopeId: scopeId, TargetOfEvaluationId: toeId,
				ControlId: ctrl2Id, State: orchestrator.ControlInScopeState_CONTROL_IN_SCOPE_STATE_ACCEPTED,
			}))

			// Compliance status: an evaluation on the sub-control must be counted (metrics
			// typically attach there), and only the latest of ctrl2's two results should count.
			assert.NoError(t, d.Create(&evaluation.EvaluationResult{
				Id: "00000000-0000-0000-000c-000000000001", TargetOfEvaluationId: toeId, AuditScopeId: scopeId,
				ControlId: ctrl1Id, ControlCatalogId: catalogId,
				Status:    evaluation.EvaluationStatus_EVALUATION_STATUS_COMPLIANT,
				Timestamp: timestamppb.New(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)),
			}))
			assert.NoError(t, d.Create(&evaluation.EvaluationResult{
				Id: "00000000-0000-0000-000c-000000000002", TargetOfEvaluationId: toeId, AuditScopeId: scopeId,
				ControlId: ctrl1sub, ControlCatalogId: catalogId, ParentControlId: new(ctrl1Id),
				Status:    evaluation.EvaluationStatus_EVALUATION_STATUS_NOT_COMPLIANT,
				Timestamp: timestamppb.New(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)),
			}))
			assert.NoError(t, d.Create(&evaluation.EvaluationResult{
				Id: "00000000-0000-0000-000c-000000000003", TargetOfEvaluationId: toeId, AuditScopeId: scopeId,
				ControlId: ctrl2Id, ControlCatalogId: catalogId,
				Status:    evaluation.EvaluationStatus_EVALUATION_STATUS_PENDING,
				Timestamp: timestamppb.New(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)),
			}))
			assert.NoError(t, d.Create(&evaluation.EvaluationResult{
				Id: "00000000-0000-0000-000c-000000000004", TargetOfEvaluationId: toeId, AuditScopeId: scopeId,
				ControlId: ctrl2Id, ControlCatalogId: catalogId,
				Status:    evaluation.EvaluationStatus_EVALUATION_STATUS_COMPLIANT,
				Timestamp: timestamppb.New(time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)),
			}))
		})
	}

	t.Run("happy path", func(t *testing.T) {
		svc := &Service{db: newDB(t), authz: &service.AuthorizationStrategyAllowAll{}}

		res, err := svc.GetAuditScopeStatistics(context.Background(), connect.NewRequest(&orchestrator.GetAuditScopeStatisticsRequest{
			AuditScopeId: scopeId,
		}))
		assert.NoError(t, err)
		if !assert.NotNil(t, res) {
			return
		}

		assert.Equal(t, map[string]int64{
			"CONTROL_IN_SCOPE_STATE_IMPLEMENTED": 1,
			"CONTROL_IN_SCOPE_STATE_ACCEPTED":    1,
		}, res.Msg.GetCountsByControlState())

		assert.Equal(t, map[string]int64{
			"EVALUATION_STATUS_COMPLIANT":     2,
			"EVALUATION_STATUS_NOT_COMPLIANT": 1,
		}, res.Msg.GetCountsByComplianceStatus())
	})

	t.Run("err: audit scope not found", func(t *testing.T) {
		svc := &Service{db: newDB(t), authz: &service.AuthorizationStrategyAllowAll{}}

		res, err := svc.GetAuditScopeStatistics(context.Background(), connect.NewRequest(&orchestrator.GetAuditScopeStatisticsRequest{
			AuditScopeId: "00000000-0000-0000-0000-000000000000",
		}))
		assert.Nil(t, res)
		assert.IsConnectError(t, err, connect.CodeNotFound)
	})

	t.Run("err: permission denied", func(t *testing.T) {
		svc := &Service{db: newDB(t), authz: &denyAuthorizationStrategy{}}

		res, err := svc.GetAuditScopeStatistics(context.Background(), connect.NewRequest(&orchestrator.GetAuditScopeStatisticsRequest{
			AuditScopeId: scopeId,
		}))
		assert.Nil(t, res)
		assert.IsConnectError(t, err, connect.CodePermissionDenied)
		assert.ErrorContains(t, err, service.ErrPermissionDenied.Error())
	})
}
