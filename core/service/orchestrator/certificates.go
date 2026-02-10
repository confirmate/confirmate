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

	cert = req.Msg.Certificate
	if !svc.hasTargetAccess(ctx, cert.GetTargetOfEvaluationId()) {
		return nil, service.ErrPermissionDenied
	}

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
		cert orchestrator.Certificate
	)

	// Validate the request
	if err = service.Validate(req); err != nil {
		return nil, err
	}

	err = svc.db.Get(&cert, "id = ?", req.Msg.CertificateId)
	if err = service.HandleDatabaseError(err, service.ErrNotFound("certificate")); err != nil {
		return nil, err
	}

	if !svc.hasTargetAccess(ctx, cert.TargetOfEvaluationId) {
		return nil, service.ErrPermissionDenied
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

	all, allowed := svc.allowedTargetOfEvaluations(ctx)
	if !all {
		certificates, npt, err = service.PaginateStorage[*orchestrator.Certificate](req.Msg, svc.db, service.DefaultPaginationOpts, "target_of_evaluation_id IN ?", allowed)
	} else {
		certificates, npt, err = service.PaginateStorage[*orchestrator.Certificate](req.Msg, svc.db, service.DefaultPaginationOpts)
	}
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

	all, allowed := svc.allowedTargetOfEvaluations(ctx)
	if !all {
		certificates, npt, err = service.PaginateStorage[*orchestrator.Certificate](req.Msg, svc.db, service.DefaultPaginationOpts, "target_of_evaluation_id IN ?", allowed)
	} else {
		certificates, npt, err = service.PaginateStorage[*orchestrator.Certificate](req.Msg, svc.db, service.DefaultPaginationOpts)
	}
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
	var cert *orchestrator.Certificate

	// Validate the request
	if err = service.Validate(req); err != nil {
		return nil, err
	}

	cert = req.Msg.Certificate
	if cert == nil || !svc.hasTargetAccess(ctx, cert.GetTargetOfEvaluationId()) {
		return nil, service.ErrPermissionDenied
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
	// Validate the request
	if err = service.Validate(req); err != nil {
		return nil, err
	}

	var cert orchestrator.Certificate

	err = svc.db.Get(&cert, "id = ?", req.Msg.CertificateId)
	if err = service.HandleDatabaseError(err, service.ErrNotFound("certificate")); err != nil {
		return nil, err
	}

	if !svc.hasTargetAccess(ctx, cert.TargetOfEvaluationId) {
		return nil, service.ErrPermissionDenied
	}

	// Delete the certificate
	err = svc.db.Delete(&cert, "id = ?", req.Msg.CertificateId)
	if err = service.HandleDatabaseError(err); err != nil {
		return nil, err
	}

	res = connect.NewResponse(&emptypb.Empty{})
	return
}
