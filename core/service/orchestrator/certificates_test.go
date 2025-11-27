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

func TestService_CreateCertificate(t *testing.T) {
	var (
		tests = []struct {
			name    string
			req     *orchestrator.CreateCertificateRequest
			wantErr bool
		}{
			{
				name: "happy path",
				req: &orchestrator.CreateCertificateRequest{
					Certificate: &orchestrator.Certificate{
						Id:          "cert-1",
						Name:        "Test Certificate",
						Description: "A test certificate",
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

			res, err := svc.CreateCertificate(context.Background(), req)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, res)
			assert.Equal(t, tt.req.Certificate.Id, res.Msg.Id)
		})
	}
}

func TestService_GetCertificate(t *testing.T) {
	var (
		tests = []struct {
			name    string
			id      string
			setup   func(*service)
			wantErr bool
		}{
			{
				name: "happy path",
				id:   "cert-1",
				setup: func(svc *service) {
					err := svc.db.Create(&orchestrator.Certificate{
						Id:          "cert-1",
						Name:        "Test Certificate",
						Description: "A test certificate",
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
				req = connect.NewRequest(&orchestrator.GetCertificateRequest{
					CertificateId: tt.id,
				})
			)

			tt.setup(svc)

			res, err := svc.GetCertificate(context.Background(), req)

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

func TestService_ListCertificates(t *testing.T) {
	var (
		tests = []struct {
			name      string
			setup     func(*service)
			wantCount int
		}{
			{
				name: "list all",
				setup: func(svc *service) {
					err := svc.db.Create(&orchestrator.Certificate{
						Id:          "cert-1",
						Name:        "Certificate 1",
						Description: "First certificate",
					})
					assert.NoError(t, err)

					err = svc.db.Create(&orchestrator.Certificate{
						Id:          "cert-2",
						Name:        "Certificate 2",
						Description: "Second certificate",
					})
					assert.NoError(t, err)
				},
				wantCount: 2,
			},
			{
				name:      "empty list",
				setup:     func(svc *service) {},
				wantCount: 0,
			},
		}
	)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				svc = &service{
					db: persistencetest.NewInMemoryDB(t, types, joinTables),
				}
				req = connect.NewRequest(&orchestrator.ListCertificatesRequest{})
			)

			tt.setup(svc)

			res, err := svc.ListCertificates(context.Background(), req)

			assert.NoError(t, err)
			assert.NotNil(t, res)
			assert.Equal(t, tt.wantCount, len(res.Msg.Certificates))
		})
	}
}

func TestService_ListPublicCertificates(t *testing.T) {
	var (
		tests = []struct {
			name      string
			setup     func(*service)
			wantCount int
		}{
			{
				name: "list all public certificates",
				setup: func(svc *service) {
					err := svc.db.Create(&orchestrator.Certificate{
						Id:          "cert-1",
						Name:        "Certificate 1",
						Description: "First certificate",
						States:      []*orchestrator.State{{State: "active"}},
					})
					assert.NoError(t, err)

					err = svc.db.Create(&orchestrator.Certificate{
						Id:          "cert-2",
						Name:        "Certificate 2",
						Description: "Second certificate",
						States:      []*orchestrator.State{{State: "pending"}},
					})
					assert.NoError(t, err)
				},
				wantCount: 2,
			},
		}
	)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				svc = &service{
					db: persistencetest.NewInMemoryDB(t, types, joinTables),
				}
				req = connect.NewRequest(&orchestrator.ListPublicCertificatesRequest{})
			)

			tt.setup(svc)

			res, err := svc.ListPublicCertificates(context.Background(), req)

			assert.NoError(t, err)
			assert.NotNil(t, res)
			assert.Equal(t, tt.wantCount, len(res.Msg.Certificates))

			// Verify that states are removed
			for _, cert := range res.Msg.Certificates {
				assert.Nil(t, cert.States)
			}
		})
	}
}

func TestService_UpdateCertificate(t *testing.T) {
	var (
		tests = []struct {
			name    string
			req     *orchestrator.UpdateCertificateRequest
			setup   func(*service)
			wantErr bool
		}{
			{
				name: "happy path",
				req: &orchestrator.UpdateCertificateRequest{
					Certificate: &orchestrator.Certificate{
						Id:          "cert-1",
						Name:        "Updated Certificate",
						Description: "Updated description",
					},
				},
				setup: func(svc *service) {
					err := svc.db.Create(&orchestrator.Certificate{
						Id:          "cert-1",
						Name:        "Test Certificate",
						Description: "Original description",
					})
					assert.NoError(t, err)
				},
				wantErr: false,
			},
			{
				name: "not found",
				req: &orchestrator.UpdateCertificateRequest{
					Certificate: &orchestrator.Certificate{
						Id:          "non-existent",
						Name:        "Updated Certificate",
						Description: "Updated description",
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

			res, err := svc.UpdateCertificate(context.Background(), req)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, res)
			assert.Equal(t, tt.req.Certificate.Name, res.Msg.Name)
		})
	}
}

func TestService_RemoveCertificate(t *testing.T) {
	var (
		tests = []struct {
			name    string
			id      string
			setup   func(*service)
			wantErr bool
		}{
			{
				name: "happy path",
				id:   "cert-1",
				setup: func(svc *service) {
					err := svc.db.Create(&orchestrator.Certificate{
						Id:          "cert-1",
						Name:        "Test Certificate",
						Description: "A test certificate",
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
				req = connect.NewRequest(&orchestrator.RemoveCertificateRequest{
					CertificateId: tt.id,
				})
			)

			tt.setup(svc)

			res, err := svc.RemoveCertificate(context.Background(), req)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, res)
		})
	}
}
