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
	"errors"
	"fmt"
	"log/slog"
	"time"

	"confirmate.io/core/api/evaluation"
	"confirmate.io/core/api/orchestrator"
	"confirmate.io/core/persistence"
	"confirmate.io/core/service"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/emptypb"
)

const (
	// CertificateStateNew indicates a valid, compliant certificate.
	CertificateStateNew = "new"
	// CertificateStateSuspended indicates the certificate is suspended due to
	// one or more non-compliant evaluation results.
	CertificateStateSuspended = "suspended"
	// CertificateStateWithdrawn indicates the certificate has been withdrawn.
	// This state is set manually and is never overridden by the lifecycle manager.
	CertificateStateWithdrawn = "withdrawn"
)

// UpdateCertificateLifecycle re-evaluates the certificate lifecycle state for
// the given audit scope. It is called by the evaluation component once a full
// catalog evaluation run has finished, so that the certificate state reflects
// the results of that run as a whole rather than of a single control.
func (svc *Service) UpdateCertificateLifecycle(
	ctx context.Context,
	req *connect.Request[orchestrator.UpdateCertificateLifecycleRequest],
) (res *connect.Response[emptypb.Empty], err error) {
	var allowed bool

	// Validate the request
	if err = service.Validate(req); err != nil {
		return nil, err
	}

	// Check access via the configured auth strategy
	allowed, _, err = CheckAccess(ctx, svc.authz, svc, orchestrator.RequestType_REQUEST_TYPE_UPDATED, req.Msg.GetAuditScopeId(), orchestrator.ObjectType_OBJECT_TYPE_AUDIT_SCOPE)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	if !allowed {
		return nil, service.ErrPermissionDenied
	}

	if err = svc.updateCertificateLifecycle(ctx, req.Msg.GetAuditScopeId()); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&emptypb.Empty{}), nil
}

// updateCertificateLifecycle re-evaluates the certificate state for the given
// audit scope and appends a new [orchestrator.State] record if the compliance
// posture has changed.
func (svc *Service) updateCertificateLifecycle(ctx context.Context, auditScopeId string) error {
	// Find the certificate linked to this audit scope (with States preloaded).
	var cert orchestrator.Certificate
	err := svc.db.Get(&cert, "audit_scope_id = ?", auditScopeId)
	if errors.Is(err, persistence.ErrRecordNotFound) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("lifecycle: get certificate: %w", err)
	}

	// Fetch the latest parent-level evaluation result per control for this scope.
	listRes, err := svc.ListEvaluationResults(ctx, connect.NewRequest(&orchestrator.ListEvaluationResultsRequest{
		Filter: &orchestrator.ListEvaluationResultsRequest_Filter{
			AuditScopeId: &auditScopeId,
			ParentsOnly:  new(true),
		},
		LatestByControlId: new(true),
	}))
	if err != nil {
		return fmt.Errorf("lifecycle: list evaluation results: %w", err)
	}
	results := listRes.Msg.GetResults()
	if len(results) == 0 {
		return nil
	}

	// Determine the target certificate state from the current results.
	target := targetCertificateState(results)
	if target == "" {
		// Only PENDING results — not enough information to change state yet.
		return nil
	}

	// Never override a manually set "withdrawn" state.
	if latestCertificateState(cert.States) == CertificateStateWithdrawn {
		return nil
	}

	// No-op if the state hasn't changed.
	if latestCertificateState(cert.States) == target {
		return nil
	}

	state := &orchestrator.State{
		Id:            uuid.NewString(),
		State:         target,
		Timestamp:     time.Now().UTC().Format(time.RFC3339),
		CertificateId: cert.Id,
	}
	if err = svc.db.Create(state); err != nil {
		return fmt.Errorf("lifecycle: create state: %w", err)
	}

	slog.Info("Certificate lifecycle state changed",
		slog.String("certificate_id", cert.Id),
		slog.String("audit_scope_id", auditScopeId),
		slog.String("state", target))

	return nil
}

// targetCertificateState derives the certificate state from a set of evaluation
// results. It returns CertificateStateSuspended if any result is non-compliant,
// CertificateStateNew if all results are compliant, and "" if the results are
// inconclusive (only PENDING).
func targetCertificateState(results []*evaluation.EvaluationResult) string {
	for _, r := range results {
		if r.GetStatus() == evaluation.EvaluationStatus_EVALUATION_STATUS_NOT_COMPLIANT ||
			r.GetStatus() == evaluation.EvaluationStatus_EVALUATION_STATUS_NOT_COMPLIANT_MANUALLY {
			return CertificateStateSuspended
		}
	}
	for _, r := range results {
		if r.GetStatus() == evaluation.EvaluationStatus_EVALUATION_STATUS_PENDING {
			return ""
		}
	}
	return CertificateStateNew
}

// latestCertificateState returns the state value of the most recently
// timestamped entry in states, or "" if the slice is empty.
func latestCertificateState(states []*orchestrator.State) string {
	var latest *orchestrator.State
	for _, s := range states {
		if latest == nil || s.Timestamp > latest.Timestamp {
			latest = s
		}
	}
	if latest == nil {
		return ""
	}
	return latest.State
}
