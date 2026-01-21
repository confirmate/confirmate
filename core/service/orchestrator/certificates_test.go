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
	"confirmate.io/core/persistence"
	"confirmate.io/core/persistence/persistencetest"
	"confirmate.io/core/service/orchestrator/orchestratortest"
	"confirmate.io/core/util/assert"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/emptypb"
)

func TestService_CreateCertificate(t *testing.T) {
	type args struct {
		req *orchestrator.CreateCertificateRequest
	}
	type fields struct {
		db persistence.DB
	}
	tests := []struct {
		name    string
		args    args
		fields  fields
		want    assert.Want[*connect.Response[orchestrator.Certificate]]
		wantErr assert.WantErr
		wantDB  assert.Want[persistence.DB]
	}{
		{
			name: "happy path",
			args: args{
				req: &orchestrator.CreateCertificateRequest{
					Certificate: orchestratortest.MockCertificate1,
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables),
			},
			want: func(t *testing.T, got *connect.Response[orchestrator.Certificate], args ...any) bool {
				assert.NotNil(t, got.Msg)
				return assert.Equal(t, orchestratortest.MockCertificate1.Id, got.Msg.Id)
			},
			wantErr: assert.NoError,
			wantDB: func(t *testing.T, db persistence.DB, msgAndArgs ...any) bool {
				cert := assert.InDB[orchestrator.Certificate](t, db, orchestratortest.MockCertificate1.Id)
				assert.Equal(t, orchestratortest.MockCertificate1.Name, cert.Name)
				return true
			},
		},
		{
			name: "validation error - empty request",
			args: args{
				req: &orchestrator.CreateCertificateRequest{},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables),
			},
			want: assert.Nil[*connect.Response[orchestrator.Certificate]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeInvalidArgument)
			},
			wantDB: func(t *testing.T, db persistence.DB, msgAndArgs ...any) bool {
				return true
			},
		},
		{
			name: "validation error - missing certificate",
			args: args{
				req: &orchestrator.CreateCertificateRequest{
					Certificate: &orchestrator.Certificate{},
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables),
			},
			want: assert.Nil[*connect.Response[orchestrator.Certificate]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeInvalidArgument) &&
					assert.IsValidationError(t, err, "certificate.target_of_evaluation_id")
			},
			wantDB: func(t *testing.T, db persistence.DB, msgAndArgs ...any) bool {
				return true
			},
		},
		{
			name: "db error - unique constraint",
			args: args{
				req: &orchestrator.CreateCertificateRequest{
					Certificate: orchestratortest.MockCertificate1,
				},
			},
			fields: fields{
				db: persistencetest.CreateErrorDB(t, persistence.ErrUniqueConstraintFailed, types, joinTables),
			},
			want: assert.Nil[*connect.Response[orchestrator.Certificate]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeAlreadyExists)
			},
			wantDB: assert.NotNil[persistence.DB],
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &Service{
				db: tt.fields.db,
			}
			res, err := svc.CreateCertificate(context.Background(), connect.NewRequest(tt.args.req))
			tt.want(t, res)
			tt.wantErr(t, err)
			tt.wantDB(t, tt.fields.db)
		})
	}
}

func TestService_GetCertificate(t *testing.T) {
	type args struct {
		req *orchestrator.GetCertificateRequest
	}
	type fields struct {
		db persistence.DB
	}
	tests := []struct {
		name    string
		args    args
		fields  fields
		want    assert.Want[*connect.Response[orchestrator.Certificate]]
		wantErr assert.WantErr
	}{
		{
			name: "happy path",
			args: args{
				req: &orchestrator.GetCertificateRequest{
					CertificateId: orchestratortest.MockCertificate1.Id,
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					err := d.Create(orchestratortest.MockCertificate1)
					assert.NoError(t, err)
				}),
			},
			want: func(t *testing.T, got *connect.Response[orchestrator.Certificate], args ...any) bool {
				assert.NotNil(t, got.Msg)
				return assert.Equal(t, orchestratortest.MockCertificate1.Id, got.Msg.Id)
			},
			wantErr: assert.NoError,
		},
		{
			name: "validation error - empty request",
			args: args{
				req: &orchestrator.GetCertificateRequest{},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables),
			},
			want: assert.Nil[*connect.Response[orchestrator.Certificate]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeInvalidArgument)
			},
		},
		{
			name: "not found",
			args: args{
				req: &orchestrator.GetCertificateRequest{
					CertificateId: orchestratortest.MockNonExistentId,
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables),
			},
			want: assert.Nil[*connect.Response[orchestrator.Certificate]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeNotFound)
			},
		},
		{
			name: "db error - not found",
			args: args{
				req: &orchestrator.GetCertificateRequest{
					CertificateId: orchestratortest.MockCertificate1.Id,
				},
			},
			fields: fields{
				db: persistencetest.GetErrorDB(t, persistence.ErrRecordNotFound, types, joinTables),
			},
			want: assert.Nil[*connect.Response[orchestrator.Certificate]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeNotFound)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &Service{
				db: tt.fields.db,
			}
			res, err := svc.GetCertificate(context.Background(), connect.NewRequest(tt.args.req))
			tt.want(t, res)
			tt.wantErr(t, err)
		})
	}
}

func TestService_ListCertificates(t *testing.T) {
	type args struct {
		req *orchestrator.ListCertificatesRequest
	}
	type fields struct {
		db persistence.DB
	}
	tests := []struct {
		name    string
		args    args
		fields  fields
		want    assert.Want[*connect.Response[orchestrator.ListCertificatesResponse]]
		wantErr assert.WantErr
	}{
		{
			name: "list all",
			args: args{
				req: &orchestrator.ListCertificatesRequest{},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					err := d.Create(orchestratortest.MockCertificate1)
					assert.NoError(t, err)
					err = d.Create(orchestratortest.MockCertificate2)
					assert.NoError(t, err)
				}),
			},
			want: func(t *testing.T, got *connect.Response[orchestrator.ListCertificatesResponse], args ...any) bool {
				assert.NotNil(t, got.Msg)
				return assert.Equal(t, 2, len(got.Msg.Certificates))
			},
			wantErr: assert.NoError,
		},
		{
			name: "empty list",
			args: args{
				req: &orchestrator.ListCertificatesRequest{},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables),
			},
			want: func(t *testing.T, got *connect.Response[orchestrator.ListCertificatesResponse], args ...any) bool {
				assert.NotNil(t, got.Msg)
				return assert.Equal(t, 0, len(got.Msg.Certificates))
			},
			wantErr: assert.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &Service{
				db: tt.fields.db,
			}
			res, err := svc.ListCertificates(context.Background(), connect.NewRequest(tt.args.req))
			tt.want(t, res)
			tt.wantErr(t, err)
		})
	}
}

func TestService_ListPublicCertificates(t *testing.T) {
	type args struct {
		req *orchestrator.ListPublicCertificatesRequest
	}
	type fields struct {
		db persistence.DB
	}
	tests := []struct {
		name    string
		args    args
		fields  fields
		want    assert.Want[*connect.Response[orchestrator.ListPublicCertificatesResponse]]
		wantErr assert.WantErr
	}{
		{
			name: "list all public certificates",
			args: args{
				req: &orchestrator.ListPublicCertificatesRequest{},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					err := d.Create(&orchestrator.Certificate{
						Id:          orchestratortest.MockCertificate1.Id,
						Name:        orchestratortest.MockCertificate1.Name,
						Description: orchestratortest.MockCertificate1.Description,
						States:      []*orchestrator.State{{State: "active"}},
					})
					assert.NoError(t, err)
					err = d.Create(&orchestrator.Certificate{
						Id:          orchestratortest.MockCertificate2.Id,
						Name:        orchestratortest.MockCertificate2.Name,
						Description: orchestratortest.MockCertificate2.Description,
						States:      []*orchestrator.State{{State: "pending"}},
					})
					assert.NoError(t, err)
				}),
			},
			want: func(t *testing.T, got *connect.Response[orchestrator.ListPublicCertificatesResponse], args ...any) bool {
				if !assert.NotNil(t, got.Msg) || !assert.Equal(t, 2, len(got.Msg.Certificates)) {
					return false
				}
				// Verify that states are removed
				for _, cert := range got.Msg.Certificates {
					if !assert.Nil(t, cert.States) {
						return false
					}
				}
				return true
			},
			wantErr: assert.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &Service{
				db: tt.fields.db,
			}
			res, err := svc.ListPublicCertificates(context.Background(), connect.NewRequest(tt.args.req))
			tt.want(t, res)
			tt.wantErr(t, err)
		})
	}
}

func TestService_UpdateCertificate(t *testing.T) {
	type args struct {
		req *orchestrator.UpdateCertificateRequest
	}
	type fields struct {
		db persistence.DB
	}
	tests := []struct {
		name    string
		args    args
		fields  fields
		want    assert.Want[*connect.Response[orchestrator.Certificate]]
		wantErr assert.WantErr
	}{
		{
			name: "happy path",
			args: args{
				req: &orchestrator.UpdateCertificateRequest{
					Certificate: &orchestrator.Certificate{
						Id:                   orchestratortest.MockCertificate1.Id,
						Name:                 "Updated Certificate",
						Description:          "Updated description",
						TargetOfEvaluationId: orchestratortest.MockToeId1,
					},
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					err := d.Create(orchestratortest.MockCertificate1)
					assert.NoError(t, err)
				}),
			},
			want: func(t *testing.T, got *connect.Response[orchestrator.Certificate], args ...any) bool {
				assert.NotNil(t, got.Msg)
				return assert.Equal(t, "Updated Certificate", got.Msg.Name)
			},
			wantErr: assert.NoError,
		},
		{
			name: "validation error - empty request",
			args: args{
				req: &orchestrator.UpdateCertificateRequest{},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables),
			},
			want: assert.Nil[*connect.Response[orchestrator.Certificate]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeInvalidArgument)
			},
		},
		{
			name: "validation error - missing id",
			args: args{
				req: &orchestrator.UpdateCertificateRequest{
					Certificate: &orchestrator.Certificate{
						Name: "Updated Certificate",
					},
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables),
			},
			want: assert.Nil[*connect.Response[orchestrator.Certificate]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeInvalidArgument) &&
					assert.IsValidationError(t, err, "certificate.id")
			},
		},
		{
			name: "not found",
			args: args{
				req: &orchestrator.UpdateCertificateRequest{
					Certificate: &orchestrator.Certificate{
						Id:                   orchestratortest.MockNonExistentId,
						Name:                 "Updated Certificate",
						Description:          "Updated description",
						TargetOfEvaluationId: orchestratortest.MockToeId1,
					},
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables),
			},
			want: assert.Nil[*connect.Response[orchestrator.Certificate]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeNotFound)
			},
		},
		{
			name: "db error - constraint",
			args: args{
				req: &orchestrator.UpdateCertificateRequest{
					Certificate: &orchestrator.Certificate{
						Id:                   orchestratortest.MockCertificate1.Id,
						Name:                 "Updated Certificate",
						Description:          "Updated description",
						TargetOfEvaluationId: orchestratortest.MockToeId1,
					},
				},
			},
			fields: fields{
				db: persistencetest.UpdateErrorDB(t, persistence.ErrConstraintFailed, types, joinTables),
			},
			want: assert.Nil[*connect.Response[orchestrator.Certificate]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeInvalidArgument)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &Service{
				db: tt.fields.db,
			}
			res, err := svc.UpdateCertificate(context.Background(), connect.NewRequest(tt.args.req))
			tt.want(t, res)
			tt.wantErr(t, err)
		})
	}
}

func TestService_RemoveCertificate(t *testing.T) {
	type args struct {
		req *orchestrator.RemoveCertificateRequest
	}
	type fields struct {
		db persistence.DB
	}
	tests := []struct {
		name    string
		args    args
		fields  fields
		want    assert.Want[*connect.Response[emptypb.Empty]]
		wantErr assert.WantErr
	}{
		{
			name: "happy path",
			args: args{
				req: &orchestrator.RemoveCertificateRequest{
					CertificateId: orchestratortest.MockCertificate1.Id,
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					err := d.Create(orchestratortest.MockCertificate1)
					assert.NoError(t, err)
				}),
			},
			want: func(t *testing.T, got *connect.Response[emptypb.Empty], args ...any) bool {
				return assert.NotNil(t, got.Msg)
			},
			wantErr: assert.NoError,
		},
		{
			name: "validation error - empty request",
			args: args{
				req: &orchestrator.RemoveCertificateRequest{},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables),
			},
			want: assert.Nil[*connect.Response[emptypb.Empty]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeInvalidArgument)
			},
		},
		{
			name: "db error - not found",
			args: args{
				req: &orchestrator.RemoveCertificateRequest{
					CertificateId: orchestratortest.MockCertificate1.Id,
				},
			},
			fields: fields{
				db: persistencetest.GetErrorDB(t, persistence.ErrRecordNotFound, types, joinTables),
			},
			want: assert.Nil[*connect.Response[emptypb.Empty]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeNotFound)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &Service{
				db: tt.fields.db,
			}
			res, err := svc.RemoveCertificate(context.Background(), connect.NewRequest(tt.args.req))
			tt.want(t, res)
			tt.wantErr(t, err)
		})
	}
}
