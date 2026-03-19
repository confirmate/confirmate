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

	"confirmate.io/core/api/orchestrator"
	"confirmate.io/core/service"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/emptypb"
)

// CreateComplianceAttestation creates a new compliance attestation.
func (svc *Service) CreateComplianceAttestation(
	ctx context.Context,
	req *connect.Request[orchestrator.CreateComplianceAttestationRequest],
) (res *connect.Response[orchestrator.ComplianceAttestation], err error) {
	var (
		attestation *orchestrator.ComplianceAttestation
	)

	// Validate the request
	if err = service.Validate(req); err != nil {
		return nil, err
	}

	attestation = req.Msg.ComplianceAttestation
	if !service.CheckAccess(svc.authz, ctx, orchestrator.RequestType_REQUEST_TYPE_CREATED, req) {
		return nil, service.ErrPermissionDenied
	}

	// Persist the new compliance attestation in the database
	err = svc.db.Create(attestation)
	if err = service.HandleDatabaseError(err); err != nil {
		return nil, err
	}

	res = connect.NewResponse(attestation)
	return
}

// GetComplianceAttestation retrieves a compliance attestation by ID.
func (svc *Service) GetComplianceAttestation(
	ctx context.Context,
	req *connect.Request[orchestrator.GetComplianceAttestationRequest],
) (res *connect.Response[orchestrator.ComplianceAttestation], err error) {
	var (
		attestation orchestrator.ComplianceAttestation
	)

	// Validate the request
	if err = service.Validate(req); err != nil {
		return nil, err
	}

	err = svc.db.Get(&attestation, "id = ?", req.Msg.ComplianceAttestationId)
	if err = service.HandleDatabaseError(err, service.ErrNotFound("compliance attestation")); err != nil {
		return nil, err
	}

	if !service.CheckAccess(svc.authz, ctx, orchestrator.RequestType_REQUEST_TYPE_UNSPECIFIED, connect.NewRequest(&attestation)) {
		return nil, service.ErrPermissionDenied
	}

	res = connect.NewResponse(&attestation)
	return
}

// ListComplianceAttestations lists all compliance attestations.
func (svc *Service) ListComplianceAttestations(
	ctx context.Context,
	req *connect.Request[orchestrator.ListComplianceAttestationsRequest],
) (res *connect.Response[orchestrator.ListComplianceAttestationsResponse], err error) {
	var (
		attestations  []*orchestrator.ComplianceAttestation
		npt           string
		all           bool
		allowed       []string
		auditScopes   []*orchestrator.AuditScope
		auditScopeIDs []string
	)

	// Validate the request
	if err = service.Validate(req); err != nil {
		return nil, err
	}

	// Set default ordering
	if req.Msg.OrderBy == "" {
		req.Msg.OrderBy = "id"
		req.Msg.Asc = true
	}

	all, allowed = svc.allowedTargetOfEvaluations(ctx)
	if !all {
		err = svc.db.List(&auditScopes, "id", true, 0, -1, "target_of_evaluation_id IN ?", allowed)
		if err = service.HandleDatabaseError(err); err != nil {
			return nil, err
		}

		for _, auditScope := range auditScopes {
			auditScopeIDs = append(auditScopeIDs, auditScope.Id)
		}

		if len(auditScopeIDs) == 0 {
			res = connect.NewResponse(&orchestrator.ListComplianceAttestationsResponse{ComplianceAttestations: []*orchestrator.ComplianceAttestation{}})
			return
		}

		attestations, npt, err = service.PaginateStorage[*orchestrator.ComplianceAttestation](req.Msg, svc.db, service.DefaultPaginationOpts, "audit_scope_id IN ?", auditScopeIDs)
	} else {
		attestations, npt, err = service.PaginateStorage[*orchestrator.ComplianceAttestation](req.Msg, svc.db, service.DefaultPaginationOpts)
	}
	if err = service.HandleDatabaseError(err); err != nil {
		return nil, err
	}

	res = connect.NewResponse(&orchestrator.ListComplianceAttestationsResponse{
		ComplianceAttestations: attestations,
		NextPageToken:          npt,
	})
	return
}

// ListPublicComplianceAttestations lists all compliance attestations without state history.
func (svc *Service) ListPublicComplianceAttestations(
	ctx context.Context,
	req *connect.Request[orchestrator.ListPublicComplianceAttestationsRequest],
) (res *connect.Response[orchestrator.ListPublicComplianceAttestationsResponse], err error) {
	var (
		attestations  []*orchestrator.ComplianceAttestation
		npt           string
		all           bool
		allowed       []string
		auditScopes   []*orchestrator.AuditScope
		auditScopeIDs []string
	)

	// Validate the request
	if err = service.Validate(req); err != nil {
		return nil, err
	}

	// Set default ordering
	if req.Msg.OrderBy == "" {
		req.Msg.OrderBy = "id"
		req.Msg.Asc = true
	}

	all, allowed = svc.allowedTargetOfEvaluations(ctx)
	if !all {
		err = svc.db.List(&auditScopes, "id", true, 0, -1, "target_of_evaluation_id IN ?", allowed)
		if err = service.HandleDatabaseError(err); err != nil {
			return nil, err
		}

		for _, auditScope := range auditScopes {
			auditScopeIDs = append(auditScopeIDs, auditScope.Id)
		}

		if len(auditScopeIDs) == 0 {
			res = connect.NewResponse(&orchestrator.ListPublicComplianceAttestationsResponse{ComplianceAttestations: []*orchestrator.ComplianceAttestation{}})
			return
		}

		attestations, npt, err = service.PaginateStorage[*orchestrator.ComplianceAttestation](req.Msg, svc.db, service.DefaultPaginationOpts, "audit_scope_id IN ?", auditScopeIDs)
	} else {
		attestations, npt, err = service.PaginateStorage[*orchestrator.ComplianceAttestation](req.Msg, svc.db, service.DefaultPaginationOpts)
	}
	if err = service.HandleDatabaseError(err); err != nil {
		return nil, err
	}

	// Remove state history from compliance attestations
	for i := range attestations {
		attestations[i].States = nil
	}

	res = connect.NewResponse(&orchestrator.ListPublicComplianceAttestationsResponse{
		ComplianceAttestations: attestations,
		NextPageToken:          npt,
	})
	return
}

// UpdateComplianceAttestation updates an existing compliance attestation.
func (svc *Service) UpdateComplianceAttestation(
	ctx context.Context,
	req *connect.Request[orchestrator.UpdateComplianceAttestationRequest],
) (res *connect.Response[orchestrator.ComplianceAttestation], err error) {
	var attestation *orchestrator.ComplianceAttestation

	// Validate the request
	if err = service.Validate(req); err != nil {
		return nil, err
	}

	attestation = req.Msg.ComplianceAttestation
	if attestation == nil || !service.CheckAccess(svc.authz, ctx, orchestrator.RequestType_REQUEST_TYPE_UPDATED, req) {
		return nil, service.ErrPermissionDenied
	}

	// Update the compliance attestation
	err = svc.db.Update(attestation, "id = ?", attestation.Id)
	if err = service.HandleDatabaseError(err, service.ErrNotFound("compliance attestation")); err != nil {
		return nil, err
	}

	res = connect.NewResponse(attestation)
	return
}

// RemoveComplianceAttestation removes a compliance attestation by ID.
func (svc *Service) RemoveComplianceAttestation(
	ctx context.Context,
	req *connect.Request[orchestrator.RemoveComplianceAttestationRequest],
) (res *connect.Response[emptypb.Empty], err error) {
	// Validate the request
	if err = service.Validate(req); err != nil {
		return nil, err
	}

	var attestation orchestrator.ComplianceAttestation

	err = svc.db.Get(&attestation, "id = ?", req.Msg.ComplianceAttestationId)
	if err = service.HandleDatabaseError(err, service.ErrNotFound("compliance attestation")); err != nil {
		return nil, err
	}

	if !service.CheckAccess(svc.authz, ctx, orchestrator.RequestType_REQUEST_TYPE_DELETED, connect.NewRequest(&attestation)) {
		return nil, service.ErrPermissionDenied
	}

	// Delete the compliance attestation
	err = svc.db.Delete(&attestation, "id = ?", req.Msg.ComplianceAttestationId)
	if err = service.HandleDatabaseError(err); err != nil {
		return nil, err
	}

	res = connect.NewResponse(&emptypb.Empty{})
	return
}
