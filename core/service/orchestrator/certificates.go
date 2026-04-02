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

// CreateCertificate creates a new certificate.
func (svc *Service) CreateCertificate(
	ctx context.Context,
	req *connect.Request[orchestrator.CreateCertificateRequest],
) (res *connect.Response[orchestrator.Certificate], err error) {
	var (
		cert *orchestrator.Certificate
	)

	// Validate the request
	if err = service.Validate(req); err != nil {
		return nil, err
	}

	// TODO(all): Generate new ID?
	cert = req.Msg.Certificate

	// TODO(all): Do we want to check that here or is it enough, that the user has a valid token?
	// if !service.CheckAccess(svc.authz, ctx, orchestrator.RequestType_REQUEST_TYPE_CREATED, req) {
	// 	return nil, service.ErrPermissionDenied
	// }

	// Persist the new certificate in the database
	err = svc.db.Create(cert)
	if err = service.HandleDatabaseError(err); err != nil {
		return nil, err
	}

	res = connect.NewResponse(cert)
	return
}

// GetCertificate retrieves a certificate by ID.
func (svc *Service) GetCertificate(
	ctx context.Context,
	req *connect.Request[orchestrator.GetCertificateRequest],
) (res *connect.Response[orchestrator.Certificate], err error) {
	var (
		cert    orchestrator.Certificate
		allowed bool
	)

	// Validate the request
	if err = service.Validate(req); err != nil {
		return nil, err
	}

	err = svc.db.Get(&cert, "id = ?", req.Msg.CertificateId)
	if err = service.HandleDatabaseError(err, service.ErrNotFound("certificate")); err != nil {
		return nil, err
	}

	// Check access via the configured auth strategy
	allowed, _, err = CheckAccess(ctx, svc.authz, svc, orchestrator.RequestType_REQUEST_TYPE_GET, req.Msg.GetCertificateId(), orchestrator.ObjectType_OBJECT_TYPE_CERTIFICATE)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	if !allowed {
		return connect.NewResponse(&orchestrator.Certificate{}), nil
	}

	res = connect.NewResponse(&cert)
	return
}

// ListCertificates lists all certificates.
func (svc *Service) ListCertificates(
	ctx context.Context,
	req *connect.Request[orchestrator.ListCertificatesRequest],
) (res *connect.Response[orchestrator.ListCertificatesResponse], err error) {
	var (
		certificates []*orchestrator.Certificate
		conds        []any
		npt          string
		resourceList []string
		allowed      bool
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

	// TODO(all): Should we check if the user has access to the TargetOfEvaluation and only return certificates for that TargetOfEvaluation? Or do we want to check access for each certificate individually?
	// Check access via the configured auth strategy
	allowed, resourceList, err = CheckAccess(ctx, svc.authz, svc, orchestrator.RequestType_REQUEST_TYPE_LIST, "", orchestrator.ObjectType_OBJECT_TYPE_CERTIFICATE)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// If access is not allowed to all resources and the resource list is empty, return an empty response
	if len(resourceList) == 0 && !allowed {
		return connect.NewResponse(&orchestrator.ListCertificatesResponse{
			Certificates:  []*orchestrator.Certificate{},
			NextPageToken: "",
		}), nil
	}

	// If access is not allowed to all resources, add a condition to filter by the allowed resource IDs
	if !allowed {
		conds = append(conds, "id IN ?", resourceList)
	}

	// Query the database with pagination and the constructed conditions
	certificates, npt, err = service.PaginateStorage[*orchestrator.Certificate](req.Msg, svc.db, service.DefaultPaginationOpts, conds...)
	if err = service.HandleDatabaseError(err); err != nil {
		return nil, err
	}

	res = connect.NewResponse(&orchestrator.ListCertificatesResponse{
		Certificates:  certificates,
		NextPageToken: npt,
	})
	return
}

// ListPublicCertificates lists all certificates without state history.
func (svc *Service) ListPublicCertificates(
	ctx context.Context,
	req *connect.Request[orchestrator.ListPublicCertificatesRequest],
) (res *connect.Response[orchestrator.ListPublicCertificatesResponse], err error) {
	var (
		certificates []*orchestrator.Certificate
		npt          string
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

	// Query the database with pagination
	certificates, npt, err = service.PaginateStorage[*orchestrator.Certificate](req.Msg, svc.db, service.DefaultPaginationOpts)
	if err = service.HandleDatabaseError(err); err != nil {
		return nil, err
	}

	// Remove state history from certificates
	for i := range certificates {
		certificates[i].States = nil
	}

	res = connect.NewResponse(&orchestrator.ListPublicCertificatesResponse{
		Certificates:  certificates,
		NextPageToken: npt,
	})
	return
}

// UpdateCertificate updates an existing certificate.
func (svc *Service) UpdateCertificate(
	ctx context.Context,
	req *connect.Request[orchestrator.UpdateCertificateRequest],
) (res *connect.Response[orchestrator.Certificate], err error) {
	var (
		cert    *orchestrator.Certificate
		allowed bool
	)

	// Validate the request
	if err = service.Validate(req); err != nil {
		return nil, err
	}

	// cert = req.Msg.Certificate
	// if cert == nil || !service.CheckAccess(svc.authz, ctx, orchestrator.RequestType_REQUEST_TYPE_UPDATED, req) {
	// 	return nil, service.ErrPermissionDenied
	// }

	// Check access via the configured auth strategy
	allowed, _, err = CheckAccess(ctx, svc.authz, svc, orchestrator.RequestType_REQUEST_TYPE_UPDATED, req.Msg.Certificate.GetId(), orchestrator.ObjectType_OBJECT_TYPE_CERTIFICATE)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	if !allowed {
		return nil, connect.NewError(connect.CodePermissionDenied, service.ErrPermissionDenied)
	}

	// Update the certificate
	err = svc.db.Update(cert, "id = ?", cert.Id)
	if err = service.HandleDatabaseError(err, service.ErrNotFound("certificate")); err != nil {
		return nil, err
	}

	res = connect.NewResponse(cert)
	return
}

// RemoveCertificate removes a certificate by ID.
func (svc *Service) RemoveCertificate(
	ctx context.Context,
	req *connect.Request[orchestrator.RemoveCertificateRequest],
) (res *connect.Response[emptypb.Empty], err error) {
	var (
		cert    orchestrator.Certificate
		allowed bool
	)
	// Validate the request
	if err = service.Validate(req); err != nil {
		return nil, err
	}

	err = svc.db.Get(&cert, "id = ?", req.Msg.CertificateId)
	if err = service.HandleDatabaseError(err, service.ErrNotFound("certificate")); err != nil {
		return nil, err
	}

	// Check access via the configured auth strategy
	allowed, _, err = CheckAccess(ctx, svc.authz, svc, orchestrator.RequestType_REQUEST_TYPE_DELETED, req.Msg.GetCertificateId(), orchestrator.ObjectType_OBJECT_TYPE_CERTIFICATE)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	if !allowed {
		return nil, connect.NewError(connect.CodePermissionDenied, service.ErrPermissionDenied)
	}

	// Delete the certificate
	err = svc.db.Delete(&cert, "id = ?", req.Msg.CertificateId)
	if err = service.HandleDatabaseError(err); err != nil {
		return nil, err
	}

	res = connect.NewResponse(&emptypb.Empty{})
	return
}
