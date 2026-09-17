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

	"confirmate.io/core/api/evaluation"
	"confirmate.io/core/api/orchestrator"
	"confirmate.io/core/persistence"
	"confirmate.io/core/service"

	"connectrpc.com/connect"
)

// GetAuditScopeStatistics retrieves aggregated statistics for an audit scope: the number of
// top-level controls in scope grouped by workflow (implementation) state, and the number of
// controls grouped by their latest compliance (evaluation) status.
func (svc *Service) GetAuditScopeStatistics(
	ctx context.Context,
	req *connect.Request[orchestrator.GetAuditScopeStatisticsRequest],
) (res *connect.Response[orchestrator.GetAuditScopeStatisticsResponse], err error) {
	var (
		scope   orchestrator.AuditScope
		allowed bool
	)

	// Validate the request
	if err = service.Validate(req); err != nil {
		return nil, err
	}

	// Check access via the configured auth strategy
	allowed, _, err = CheckAccess(ctx, svc.authz, svc, orchestrator.RequestType_REQUEST_TYPE_GET, req.Msg.GetAuditScopeId(), orchestrator.ObjectType_OBJECT_TYPE_AUDIT_SCOPE)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	if !allowed {
		return nil, service.ErrPermissionDenied
	}

	err = svc.db.Get(&scope, persistence.WithoutPreload(), "id = ?", req.Msg.GetAuditScopeId())
	if err = service.HandleDatabaseError(err, service.ErrNotFound("audit scope")); err != nil {
		return nil, err
	}

	res = connect.NewResponse(&orchestrator.GetAuditScopeStatisticsResponse{
		CountsByControlState:     map[string]int64{},
		CountsByComplianceStatus: map[string]int64{},
	})

	// Workflow status: count ControlInScope rows by state, restricted to top-level controls.
	// ControlInScope records exist for every control including sub-controls (see
	// autoCreateControlsInScope), so counting all of them would inflate the numbers beyond what
	// a single row in the UI represents. Aggregation is done in Go rather than via SQL GROUP BY
	// so this works against every supported database, including the in-memory one used in tests.
	var topLevelControls []*orchestrator.Control
	if err = svc.db.List(&topLevelControls, "", true, 0, -1, persistence.WithoutPreload(),
		"catalog_id = ? AND parent_control_id IS NULL", scope.GetCatalogId()); err != nil {
		return nil, service.HandleDatabaseError(err)
	}
	topLevelControlIds := make(map[string]bool, len(topLevelControls))
	for _, c := range topLevelControls {
		topLevelControlIds[c.GetId()] = true
	}

	var controlsInScope []*orchestrator.ControlInScope
	if err = svc.db.List(&controlsInScope, "", true, 0, -1, persistence.WithoutPreload(),
		"audit_scope_id = ?", req.Msg.GetAuditScopeId()); err != nil {
		return nil, service.HandleDatabaseError(err)
	}
	for _, cis := range controlsInScope {
		if topLevelControlIds[cis.GetControlId()] {
			res.Msg.CountsByControlState[cis.GetState().String()]++
		}
	}

	// Compliance status: latest evaluation result per control (regardless of depth: metrics are
	// typically attached to sub-controls rather than their top-level parent), grouped by status.
	// Uses PostgreSQL's DISTINCT ON, mirroring ListEvaluationResults' latest_by_control_id path.
	var latestResults []*evaluation.EvaluationResult
	if err = svc.db.Raw(&latestResults, `
		SELECT DISTINCT ON (control_id) *
		FROM evaluation_results
		WHERE audit_scope_id = ?
		ORDER BY control_id, timestamp DESC, id DESC
	`, req.Msg.GetAuditScopeId()); err != nil {
		return nil, service.HandleDatabaseError(err)
	}
	for _, r := range latestResults {
		res.Msg.CountsByComplianceStatus[r.GetStatus().String()]++
	}

	return
}
