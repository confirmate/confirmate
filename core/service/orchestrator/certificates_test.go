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
	"confirmate.io/core/service"
	"confirmate.io/core/service/orchestrator/orchestratortest"
	"confirmate.io/core/util/assert"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/emptypb"
)

func TestService_CreateComplianceAttestation(t *testing.T) {
	type args struct {
		req *orchestrator.CreateComplianceAttestationRequest
	}
	type fields struct {
		db    persistence.DB
		authz service.AuthorizationStrategy
	}
	tests := []struct {
		name    string
		args    args
		fields  fields
		want    assert.Want[*connect.Response[orchestrator.ComplianceAttestation]]
		wantErr assert.WantErr
		wantDB  assert.Want[persistence.DB]
	}{
		{
			name:   "happy path",
			args:   args{req: &orchestrator.CreateComplianceAttestationRequest{ComplianceAttestation: orchestratortest.MockCertificate1}},
			fields: fields{db: persistencetest.NewInMemoryDB(t, types, joinTables)},
			want: func(t *testing.T, got *connect.Response[orchestrator.ComplianceAttestation], args ...any) bool {
				assert.NotNil(t, got.Msg)
				return assert.Equal(t, orchestratortest.MockCertificate1.Id, got.Msg.Id)
			},
			wantErr: assert.NoError,
			wantDB: func(t *testing.T, db persistence.DB, msgAndArgs ...any) bool {
				attestation := assert.InDB[orchestrator.ComplianceAttestation](t, db, orchestratortest.MockCertificate1.Id)
				assert.Equal(t, orchestratortest.MockCertificate1.Name, attestation.Name)
				return true
			},
		},
		{
			name:   "validation error - empty request",
			args:   args{req: &orchestrator.CreateComplianceAttestationRequest{}},
			fields: fields{db: persistencetest.NewInMemoryDB(t, types, joinTables)},
			want:   assert.Nil[*connect.Response[orchestrator.ComplianceAttestation]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeInvalidArgument)
			},
			wantDB: func(t *testing.T, db persistence.DB, msgAndArgs ...any) bool { return true },
		},
		{
			name:   "validation error - missing attestation",
			args:   args{req: &orchestrator.CreateComplianceAttestationRequest{ComplianceAttestation: &orchestrator.ComplianceAttestation{}}},
			fields: fields{db: persistencetest.NewInMemoryDB(t, types, joinTables)},
			want:   assert.Nil[*connect.Response[orchestrator.ComplianceAttestation]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeInvalidArgument) &&
					assert.IsValidationError(t, err, "compliance_attestation.audit_scope_id")
			},
			wantDB: func(t *testing.T, db persistence.DB, msgAndArgs ...any) bool { return true },
		},
		{
			name: "authorization failure",
			args: args{req: &orchestrator.CreateComplianceAttestationRequest{ComplianceAttestation: &orchestrator.ComplianceAttestation{
				Id: orchestratortest.MockCertificate1.Id, Name: orchestratortest.MockCertificate1.Name, AuditScopeId: orchestratortest.MockCertificate1.AuditScopeId,
			}}},
			fields: fields{db: persistencetest.NewInMemoryDB(t, types, joinTables), authz: &denyAuthorizationStrategy{}},
			want:   assert.Nil[*connect.Response[orchestrator.ComplianceAttestation]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodePermissionDenied)
			},
			wantDB: func(t *testing.T, db persistence.DB, msgAndArgs ...any) bool { return true },
		},
		{
			name:   "db error - unique constraint",
			args:   args{req: &orchestrator.CreateComplianceAttestationRequest{ComplianceAttestation: orchestratortest.MockCertificate1}},
			fields: fields{db: persistencetest.CreateErrorDB(t, persistence.ErrUniqueConstraintFailed, types, joinTables)},
			want:   assert.Nil[*connect.Response[orchestrator.ComplianceAttestation]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeAlreadyExists)
			},
			wantDB: assert.NotNil[persistence.DB],
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &Service{db: tt.fields.db, authz: tt.fields.authz}
			res, err := svc.CreateComplianceAttestation(context.Background(), connect.NewRequest(tt.args.req))
			tt.want(t, res)
			tt.wantErr(t, err)
			tt.wantDB(t, tt.fields.db)
		})
	}
}

func TestService_CreateComplianceAttestation_AuthorizationFailure(t *testing.T) {
	type args struct {
		req *orchestrator.CreateComplianceAttestationRequest
	}
	tests := []struct {
		name    string
		args    args
		wantErr assert.WantErr
	}{
		{
			name: "permission denied",
			args: args{req: &orchestrator.CreateComplianceAttestationRequest{ComplianceAttestation: &orchestrator.ComplianceAttestation{
				Id: orchestratortest.MockCertificate1.Id, Name: orchestratortest.MockCertificate1.Name, AuditScopeId: orchestratortest.MockCertificate1.AuditScopeId,
			}}},
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodePermissionDenied)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &Service{db: persistencetest.NewInMemoryDB(t, types, joinTables), authz: &denyAuthorizationStrategy{}}

			res, err := svc.CreateComplianceAttestation(context.Background(), connect.NewRequest(tt.args.req))
			assert.Nil(t, res)
			tt.wantErr(t, err)
		})
	}
}

func TestService_GetComplianceAttestation(t *testing.T) {
	type args struct {
		req *orchestrator.GetComplianceAttestationRequest
	}
	type fields struct {
		db    persistence.DB
		authz service.AuthorizationStrategy
	}
	tests := []struct {
		name    string
		args    args
		fields  fields
		want    assert.Want[*connect.Response[orchestrator.ComplianceAttestation]]
		wantErr assert.WantErr
	}{
		{
			name: "happy path",
			args: args{req: &orchestrator.GetComplianceAttestationRequest{ComplianceAttestationId: orchestratortest.MockCertificate1.Id}},
			fields: fields{db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
				err := d.Create(orchestratortest.MockCertificate1)
				assert.NoError(t, err)
			})},
			want: func(t *testing.T, got *connect.Response[orchestrator.ComplianceAttestation], args ...any) bool {
				assert.NotNil(t, got.Msg)
				return assert.Equal(t, orchestratortest.MockCertificate1.Id, got.Msg.Id)
			},
			wantErr: assert.NoError,
		},
		{
			name:   "validation error - empty request",
			args:   args{req: &orchestrator.GetComplianceAttestationRequest{}},
			fields: fields{db: persistencetest.NewInMemoryDB(t, types, joinTables)},
			want:   assert.Nil[*connect.Response[orchestrator.ComplianceAttestation]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeInvalidArgument)
			},
		},
		{
			name:   "not found",
			args:   args{req: &orchestrator.GetComplianceAttestationRequest{ComplianceAttestationId: orchestratortest.MockNonExistentId}},
			fields: fields{db: persistencetest.NewInMemoryDB(t, types, joinTables)},
			want:   assert.Nil[*connect.Response[orchestrator.ComplianceAttestation]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeNotFound)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &Service{db: tt.fields.db, authz: tt.fields.authz}
			res, err := svc.GetComplianceAttestation(context.Background(), connect.NewRequest(tt.args.req))
			tt.want(t, res)
			tt.wantErr(t, err)
		})
	}
}

func TestService_ListComplianceAttestations(t *testing.T) {
	tests := []struct {
		name    string
		req     *orchestrator.ListComplianceAttestationsRequest
		db      persistence.DB
		want    assert.Want[*connect.Response[orchestrator.ListComplianceAttestationsResponse]]
		wantErr assert.WantErr
	}{
		{
			name: "list all",
			req:  &orchestrator.ListComplianceAttestationsRequest{},
			db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
				err := d.Create(orchestratortest.MockAuditScope1)
				assert.NoError(t, err)
				err = d.Create(orchestratortest.MockAuditScope2)
				assert.NoError(t, err)
				err = d.Create(orchestratortest.MockCertificate1)
				assert.NoError(t, err)
				err = d.Create(orchestratortest.MockCertificate2)
				assert.NoError(t, err)
			}),
			want: func(t *testing.T, got *connect.Response[orchestrator.ListComplianceAttestationsResponse], args ...any) bool {
				assert.NotNil(t, got.Msg)
				return assert.Equal(t, 2, len(got.Msg.ComplianceAttestations))
			},
			wantErr: assert.NoError,
		},
		{
			name: "empty list",
			req:  &orchestrator.ListComplianceAttestationsRequest{},
			db:   persistencetest.NewInMemoryDB(t, types, joinTables),
			want: func(t *testing.T, got *connect.Response[orchestrator.ListComplianceAttestationsResponse], args ...any) bool {
				assert.NotNil(t, got.Msg)
				return assert.Equal(t, 0, len(got.Msg.ComplianceAttestations))
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &Service{db: tt.db}
			res, err := svc.ListComplianceAttestations(context.Background(), connect.NewRequest(tt.req))
			tt.want(t, res)
			tt.wantErr(t, err)
		})
	}
}

func TestService_ListPublicComplianceAttestations(t *testing.T) {
	tests := []struct {
		name    string
		req     *orchestrator.ListPublicComplianceAttestationsRequest
		db      persistence.DB
		want    assert.Want[*connect.Response[orchestrator.ListPublicComplianceAttestationsResponse]]
		wantErr assert.WantErr
	}{
		{
			name: "list all public attestations",
			req:  &orchestrator.ListPublicComplianceAttestationsRequest{},
			db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
				err := d.Create(orchestratortest.MockAuditScope1)
				assert.NoError(t, err)
				err = d.Create(orchestratortest.MockAuditScope2)
				assert.NoError(t, err)
				err = d.Create(&orchestrator.ComplianceAttestation{Id: orchestratortest.MockCertificate1.Id, Name: orchestratortest.MockCertificate1.Name, Description: orchestratortest.MockCertificate1.Description, AuditScopeId: orchestratortest.MockCertificate1.AuditScopeId, States: []*orchestrator.ComplianceAttestationState{{State: orchestrator.AttestationState_ATTESTATION_STATE_ACTIVE}}})
				assert.NoError(t, err)
				err = d.Create(&orchestrator.ComplianceAttestation{Id: orchestratortest.MockCertificate2.Id, Name: orchestratortest.MockCertificate2.Name, Description: orchestratortest.MockCertificate2.Description, AuditScopeId: orchestratortest.MockCertificate2.AuditScopeId, States: []*orchestrator.ComplianceAttestationState{{State: orchestrator.AttestationState_ATTESTATION_STATE_IN_PROGRESS}}})
				assert.NoError(t, err)
			}),
			want: func(t *testing.T, got *connect.Response[orchestrator.ListPublicComplianceAttestationsResponse], args ...any) bool {
				if !assert.NotNil(t, got.Msg) || !assert.Equal(t, 2, len(got.Msg.ComplianceAttestations)) {
					return false
				}
				for _, att := range got.Msg.ComplianceAttestations {
					if !assert.Nil(t, att.States) {
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
			svc := &Service{db: tt.db}
			res, err := svc.ListPublicComplianceAttestations(context.Background(), connect.NewRequest(tt.req))
			tt.want(t, res)
			tt.wantErr(t, err)
		})
	}
}

func TestService_UpdateComplianceAttestation(t *testing.T) {
	tests := []struct {
		name    string
		req     *orchestrator.UpdateComplianceAttestationRequest
		db      persistence.DB
		authz   service.AuthorizationStrategy
		want    assert.Want[*connect.Response[orchestrator.ComplianceAttestation]]
		wantErr assert.WantErr
	}{
		{
			name: "happy path",
			req:  &orchestrator.UpdateComplianceAttestationRequest{ComplianceAttestation: &orchestrator.ComplianceAttestation{Id: orchestratortest.MockCertificate1.Id, Name: "Updated Certificate", Description: "Updated description", AuditScopeId: orchestratortest.MockScopeId1}},
			db:   persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) { err := d.Create(orchestratortest.MockCertificate1); assert.NoError(t, err) }),
			want: func(t *testing.T, got *connect.Response[orchestrator.ComplianceAttestation], args ...any) bool {
				assert.NotNil(t, got.Msg)
				return assert.Equal(t, "Updated Certificate", got.Msg.Name)
			},
			wantErr: assert.NoError,
		},
		{
			name: "validation error - empty request",
			req:  &orchestrator.UpdateComplianceAttestationRequest{},
			db:   persistencetest.NewInMemoryDB(t, types, joinTables),
			want: assert.Nil[*connect.Response[orchestrator.ComplianceAttestation]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeInvalidArgument)
			},
		},
		{
			name: "validation error - missing id",
			req:  &orchestrator.UpdateComplianceAttestationRequest{ComplianceAttestation: &orchestrator.ComplianceAttestation{Name: "Updated Certificate"}},
			db:   persistencetest.NewInMemoryDB(t, types, joinTables),
			want: assert.Nil[*connect.Response[orchestrator.ComplianceAttestation]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeInvalidArgument) &&
					assert.IsValidationError(t, err, "compliance_attestation.id")
			},
		},
		{
			name: "not found",
			req: &orchestrator.UpdateComplianceAttestationRequest{ComplianceAttestation: &orchestrator.ComplianceAttestation{
				Id: orchestratortest.MockNonExistentId, Name: "Updated Certificate", Description: "Updated description", AuditScopeId: orchestratortest.MockScopeId1}},
			db:   persistencetest.NewInMemoryDB(t, types, joinTables),
			want: assert.Nil[*connect.Response[orchestrator.ComplianceAttestation]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeNotFound)
			},
		},
		{
			name: "authorization failure",
			req: &orchestrator.UpdateComplianceAttestationRequest{ComplianceAttestation: &orchestrator.ComplianceAttestation{
				Id: orchestratortest.MockCertificate1.Id, Name: "Updated Certificate", Description: "Updated description", AuditScopeId: orchestratortest.MockScopeId1}},
			db:    persistencetest.NewInMemoryDB(t, types, joinTables),
			authz: &denyAuthorizationStrategy{},
			want:  assert.Nil[*connect.Response[orchestrator.ComplianceAttestation]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodePermissionDenied)
			},
		},
		{
			name: "db error - constraint",
			req: &orchestrator.UpdateComplianceAttestationRequest{ComplianceAttestation: &orchestrator.ComplianceAttestation{
				Id: orchestratortest.MockCertificate1.Id, Name: "Updated Certificate", Description: "Updated description", AuditScopeId: orchestratortest.MockScopeId1}},
			db:   persistencetest.UpdateErrorDB(t, persistence.ErrConstraintFailed, types, joinTables),
			want: assert.Nil[*connect.Response[orchestrator.ComplianceAttestation]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeInvalidArgument)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &Service{db: tt.db, authz: tt.authz}
			res, err := svc.UpdateComplianceAttestation(context.Background(), connect.NewRequest(tt.req))
			tt.want(t, res)
			tt.wantErr(t, err)
		})
	}
}

func TestService_RemoveComplianceAttestation(t *testing.T) {
	tests := []struct {
		name    string
		req     *orchestrator.RemoveComplianceAttestationRequest
		db      persistence.DB
		authz   service.AuthorizationStrategy
		want    assert.Want[*connect.Response[emptypb.Empty]]
		wantErr assert.WantErr
	}{
		{
			name: "happy path",
			req:  &orchestrator.RemoveComplianceAttestationRequest{ComplianceAttestationId: orchestratortest.MockCertificate1.Id},
			db:   persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) { err := d.Create(orchestratortest.MockCertificate1); assert.NoError(t, err) }),
			want: func(t *testing.T, got *connect.Response[emptypb.Empty], args ...any) bool {
				return assert.NotNil(t, got.Msg)
			},
			wantErr: assert.NoError,
		},
		{
			name: "validation error - empty request",
			req:  &orchestrator.RemoveComplianceAttestationRequest{},
			db:   persistencetest.NewInMemoryDB(t, types, joinTables),
			want: assert.Nil[*connect.Response[emptypb.Empty]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeInvalidArgument)
			},
		},
		{
			name: "authorization failure",
			req:  &orchestrator.RemoveComplianceAttestationRequest{ComplianceAttestationId: orchestratortest.MockCertificate1.Id},
			db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
				err := d.Create(orchestratortest.MockCertificate1)
				assert.NoError(t, err)
			}),
			authz: &denyAuthorizationStrategy{},
			want:  assert.Nil[*connect.Response[emptypb.Empty]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodePermissionDenied)
			},
		},
		{
			name: "db error - not found",
			req:  &orchestrator.RemoveComplianceAttestationRequest{ComplianceAttestationId: orchestratortest.MockCertificate1.Id},
			db:   persistencetest.GetErrorDB(t, persistence.ErrRecordNotFound, types, joinTables),
			want: assert.Nil[*connect.Response[emptypb.Empty]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeNotFound)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &Service{db: tt.db, authz: tt.authz}
			res, err := svc.RemoveComplianceAttestation(context.Background(), connect.NewRequest(tt.req))
			tt.want(t, res)
			tt.wantErr(t, err)
		})
	}
}
