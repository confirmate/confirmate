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

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/emptypb"
	"gorm.io/gorm"
)

// CreateCertificate creates a new certificate.
func (svc *service) CreateCertificate(
	ctx context.Context,
	req *connect.Request[orchestrator.CreateCertificateRequest],
) (*connect.Response[orchestrator.Certificate], error) {
	// Persist the new certificate in the database
	err := svc.db.Create(req.Msg.Certificate)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("could not add certificate to the database: %w", err))
	}

	return connect.NewResponse(req.Msg.Certificate), nil
}

// GetCertificate retrieves a certificate by ID.
func (svc *service) GetCertificate(
	ctx context.Context,
	req *connect.Request[orchestrator.GetCertificateRequest],
) (*connect.Response[orchestrator.Certificate], error) {
	var res orchestrator.Certificate

	err := svc.db.Get(&res, "id = ?", req.Msg.CertificateId)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("certificate not found"))
	} else if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("database error: %w", err))
	}

	return connect.NewResponse(&res), nil
}

// ListCertificates lists all certificates.
func (svc *service) ListCertificates(
	ctx context.Context,
	req *connect.Request[orchestrator.ListCertificatesRequest],
) (*connect.Response[orchestrator.ListCertificatesResponse], error) {
	var certificates []*orchestrator.Certificate

	err := svc.db.List(&certificates, "id", true, 0, -1, nil)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("could not list certificates: %w", err))
	}

	return connect.NewResponse(&orchestrator.ListCertificatesResponse{
		Certificates: certificates,
	}), nil
}

// ListPublicCertificates lists all certificates without state history.
func (svc *service) ListPublicCertificates(
	ctx context.Context,
	req *connect.Request[orchestrator.ListPublicCertificatesRequest],
) (*connect.Response[orchestrator.ListPublicCertificatesResponse], error) {
	var certificates []*orchestrator.Certificate

	err := svc.db.List(&certificates, "id", true, 0, -1, nil)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("could not list certificates: %w", err))
	}

	// Remove state history from certificates
	for i := range certificates {
		certificates[i].States = nil
	}

	return connect.NewResponse(&orchestrator.ListPublicCertificatesResponse{
		Certificates: certificates,
	}), nil
}

// UpdateCertificate updates an existing certificate.
func (svc *service) UpdateCertificate(
	ctx context.Context,
	req *connect.Request[orchestrator.UpdateCertificateRequest],
) (*connect.Response[orchestrator.Certificate], error) {
	// Check if the certificate exists
	count, err := svc.db.Count(req.Msg.Certificate, "id = ?", req.Msg.Certificate.Id)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("database error: %w", err))
	}

	if count == 0 {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("certificate not found"))
	}

	// Save the updated certificate
	err = svc.db.Save(req.Msg.Certificate, "id = ?", req.Msg.Certificate.Id)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("database error: %w", err))
	}

	return connect.NewResponse(req.Msg.Certificate), nil
}

// RemoveCertificate removes a certificate by ID.
func (svc *service) RemoveCertificate(
	ctx context.Context,
	req *connect.Request[orchestrator.RemoveCertificateRequest],
) (*connect.Response[emptypb.Empty], error) {
	var cert orchestrator.Certificate

	// Delete the certificate
	err := svc.db.Delete(&cert, "id = ?", req.Msg.CertificateId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("database error: %w", err))
	}

	return connect.NewResponse(&emptypb.Empty{}), nil
}
