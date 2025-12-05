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
	"errors"
	"fmt"

	"confirmate.io/core/api/orchestrator"
	"confirmate.io/core/persistence"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/emptypb"
)

// CreateCertificate creates a new certificate.
func (svc *service) CreateCertificate(
	ctx context.Context,
	req *connect.Request[orchestrator.CreateCertificateRequest],
) (res *connect.Response[orchestrator.Certificate], err error) {
	var (
		cert = req.Msg.Certificate
	)

	// Persist the new certificate in the database
	err = svc.db.Create(cert)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("could not add certificate to the database: %w", err))
	}

	res = connect.NewResponse(cert)
	return
}

// GetCertificate retrieves a certificate by ID.
func (svc *service) GetCertificate(
	ctx context.Context,
	req *connect.Request[orchestrator.GetCertificateRequest],
) (res *connect.Response[orchestrator.Certificate], err error) {
	var (
		cert orchestrator.Certificate
	)

	err = svc.db.Get(&cert, "id = ?", req.Msg.CertificateId)
	if errors.Is(err, persistence.ErrRecordNotFound) {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("certificate not found"))
	} else if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("database error: %w", err))
	}

	res = connect.NewResponse(&cert)
	return
}

// ListCertificates lists all certificates.
func (svc *service) ListCertificates(
	ctx context.Context,
	req *connect.Request[orchestrator.ListCertificatesRequest],
) (res *connect.Response[orchestrator.ListCertificatesResponse], err error) {
	var (
		certificates []*orchestrator.Certificate
	)

	err = svc.db.List(&certificates, "id", true, 0, -1, nil)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("could not list certificates: %w", err))
	}

	res = connect.NewResponse(&orchestrator.ListCertificatesResponse{
		Certificates: certificates,
	})
	return
}

// ListPublicCertificates lists all certificates without state history.
func (svc *service) ListPublicCertificates(
	ctx context.Context,
	req *connect.Request[orchestrator.ListPublicCertificatesRequest],
) (res *connect.Response[orchestrator.ListPublicCertificatesResponse], err error) {
	var (
		certificates []*orchestrator.Certificate
	)

	err = svc.db.List(&certificates, "id", true, 0, -1, nil)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("could not list certificates: %w", err))
	}

	// Remove state history from certificates
	for i := range certificates {
		certificates[i].States = nil
	}

	res = connect.NewResponse(&orchestrator.ListPublicCertificatesResponse{
		Certificates: certificates,
	})
	return
}

// UpdateCertificate updates an existing certificate.
func (svc *service) UpdateCertificate(
	ctx context.Context,
	req *connect.Request[orchestrator.UpdateCertificateRequest],
) (res *connect.Response[orchestrator.Certificate], err error) {
	var (
		count int64
		cert  = req.Msg.Certificate
	)

	// Check if the certificate exists
	count, err = svc.db.Count(cert, "id = ?", cert.Id)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("database error: %w", err))
	}

	if count == 0 {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("certificate not found"))
	}

	// Save the updated certificate
	err = svc.db.Save(cert, "id = ?", cert.Id)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("database error: %w", err))
	}

	res = connect.NewResponse(cert)
	return
}

// RemoveCertificate removes a certificate by ID.
func (svc *service) RemoveCertificate(
	ctx context.Context,
	req *connect.Request[orchestrator.RemoveCertificateRequest],
) (res *connect.Response[emptypb.Empty], err error) {
	var (
		cert orchestrator.Certificate
	)

	// Delete the certificate
	err = svc.db.Delete(&cert, "id = ?", req.Msg.CertificateId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("database error: %w", err))
	}

	res = connect.NewResponse(&emptypb.Empty{})
	return
}
