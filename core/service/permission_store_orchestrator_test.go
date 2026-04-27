package service

import (
	"context"
	"testing"

	"confirmate.io/core/api/orchestrator"
	"confirmate.io/core/api/orchestrator/orchestratorconnect"
	"confirmate.io/core/server"
	"confirmate.io/core/server/servertest"
	"confirmate.io/core/service/orchestrator/orchestratortest"
	"confirmate.io/core/util/assert"
	"connectrpc.com/connect"
)

// mockPermissionHandler is a minimal orchestrator handler for testing OrchestratorPermissionStore.
type mockPermissionHandler struct {
	orchestratorconnect.UnimplementedOrchestratorHandler
	permissions []*orchestrator.UserPermission
	listErr     error
}

func (h *mockPermissionHandler) ListUserPermissions(_ context.Context, req *connect.Request[orchestrator.ListUserPermissionsRequest]) (*connect.Response[orchestrator.ListUserPermissionsResponse], error) {
	if h.listErr != nil {
		return nil, h.listErr
	}

	var filtered []*orchestrator.UserPermission
	for _, p := range h.permissions {
		if req.Msg.GetUserId() == "" || p.GetUserId() == req.Msg.GetUserId() {
			filtered = append(filtered, p)
		}
	}

	return connect.NewResponse(&orchestrator.ListUserPermissionsResponse{
		UserPermissions: filtered,
	}), nil
}

func newPermissionStoreForTest(t *testing.T, permissions []*orchestrator.UserPermission, listErr error) *OrchestratorPermissionStore {
	t.Helper()

	handler := &mockPermissionHandler{permissions: permissions, listErr: listErr}
	_, testSrv := servertest.NewTestConnectServer(t,
		server.WithHandler(orchestratorconnect.NewOrchestratorHandler(handler)),
	)
	t.Cleanup(testSrv.Close)

	client := orchestratorconnect.NewOrchestratorClient(testSrv.Client(), testSrv.URL)
	return &OrchestratorPermissionStore{Client: client}
}

func TestOrchestratorPermissionStore_HasPermission(t *testing.T) {
	type args struct {
		userId     string
		resourceId string
		permission orchestrator.UserPermission_Permission
		reqType    orchestrator.RequestType
		objectType orchestrator.ObjectType
	}
	tests := []struct {
		name        string
		permissions []*orchestrator.UserPermission
		listErr     error
		args        args
		want        bool
		wantErr     assert.WantErr
	}{
		{
			name: "err: list permissions fails",
			listErr: connect.NewError(connect.CodeInternal, nil),
			args: args{
				userId:     orchestratortest.MockUserId1,
				resourceId: orchestratortest.MockTargetOfEvaluation1.Id,
				permission: orchestrator.UserPermission_PERMISSION_READER,
				objectType: orchestrator.ObjectType_OBJECT_TYPE_TARGET_OF_EVALUATION,
			},
			want:    false,
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeInternal)
			},
		},
		{
			name: "false: user has no matching permission",
			permissions: []*orchestrator.UserPermission{
				{
					UserId:       orchestratortest.MockUserId1,
					ResourceId:   orchestratortest.MockTargetOfEvaluation1.Id,
					ResourceType: orchestrator.ObjectType_OBJECT_TYPE_TARGET_OF_EVALUATION,
					Permission:   orchestrator.UserPermission_PERMISSION_READER,
				},
			},
			args: args{
				userId:     orchestratortest.MockUserId1,
				resourceId: "other-resource",
				permission: orchestrator.UserPermission_PERMISSION_READER,
				objectType: orchestrator.ObjectType_OBJECT_TYPE_TARGET_OF_EVALUATION,
			},
			want:    false,
			wantErr: assert.NoError,
		},
		{
			name: "false: insufficient permission level",
			permissions: []*orchestrator.UserPermission{
				{
					UserId:       orchestratortest.MockUserId1,
					ResourceId:   orchestratortest.MockTargetOfEvaluation1.Id,
					ResourceType: orchestrator.ObjectType_OBJECT_TYPE_TARGET_OF_EVALUATION,
					Permission:   orchestrator.UserPermission_PERMISSION_READER,
				},
			},
			args: args{
				userId:     orchestratortest.MockUserId1,
				resourceId: orchestratortest.MockTargetOfEvaluation1.Id,
				permission: orchestrator.UserPermission_PERMISSION_ADMIN,
				objectType: orchestrator.ObjectType_OBJECT_TYPE_TARGET_OF_EVALUATION,
			},
			want:    false,
			wantErr: assert.NoError,
		},
		{
			name: "true: exact permission match",
			permissions: []*orchestrator.UserPermission{
				{
					UserId:       orchestratortest.MockUserId1,
					ResourceId:   orchestratortest.MockTargetOfEvaluation1.Id,
					ResourceType: orchestrator.ObjectType_OBJECT_TYPE_TARGET_OF_EVALUATION,
					Permission:   orchestrator.UserPermission_PERMISSION_CONTRIBUTOR,
				},
			},
			args: args{
				userId:     orchestratortest.MockUserId1,
				resourceId: orchestratortest.MockTargetOfEvaluation1.Id,
				permission: orchestrator.UserPermission_PERMISSION_READER,
				objectType: orchestrator.ObjectType_OBJECT_TYPE_TARGET_OF_EVALUATION,
			},
			want:    true,
			wantErr: assert.NoError,
		},
		{
			name: "false: other user's permission is not returned",
			permissions: []*orchestrator.UserPermission{
				{
					UserId:       "other-user",
					ResourceId:   orchestratortest.MockTargetOfEvaluation1.Id,
					ResourceType: orchestrator.ObjectType_OBJECT_TYPE_TARGET_OF_EVALUATION,
					Permission:   orchestrator.UserPermission_PERMISSION_ADMIN,
				},
			},
			args: args{
				userId:     orchestratortest.MockUserId1,
				resourceId: orchestratortest.MockTargetOfEvaluation1.Id,
				permission: orchestrator.UserPermission_PERMISSION_READER,
				objectType: orchestrator.ObjectType_OBJECT_TYPE_TARGET_OF_EVALUATION,
			},
			want:    false,
			wantErr: assert.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ps := newPermissionStoreForTest(t, tt.permissions, tt.listErr)
			got, err := ps.HasPermission(context.Background(), tt.args.userId, tt.args.resourceId, tt.args.permission, tt.args.reqType, tt.args.objectType)
			assert.Equal(t, tt.want, got)
			tt.wantErr(t, err)
		})
	}
}

func TestOrchestratorPermissionStore_PermissionForResources(t *testing.T) {
	type args struct {
		userId     string
		permission orchestrator.UserPermission_Permission
		reqType    orchestrator.RequestType
		objectType orchestrator.ObjectType
	}
	tests := []struct {
		name        string
		permissions []*orchestrator.UserPermission
		listErr     error
		args        args
		want        []string
		wantErr     assert.WantErr
	}{
		{
			name:    "err: list permissions fails",
			listErr: connect.NewError(connect.CodeInternal, nil),
			args: args{
				userId:     orchestratortest.MockUserId1,
				permission: orchestrator.UserPermission_PERMISSION_READER,
				objectType: orchestrator.ObjectType_OBJECT_TYPE_TARGET_OF_EVALUATION,
			},
			want:    nil,
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeInternal)
			},
		},
		{
			name: "returns matching resource IDs for the user",
			permissions: []*orchestrator.UserPermission{
				{
					UserId:       orchestratortest.MockUserId1,
					ResourceId:   orchestratortest.MockTargetOfEvaluation1.Id,
					ResourceType: orchestrator.ObjectType_OBJECT_TYPE_TARGET_OF_EVALUATION,
					Permission:   orchestrator.UserPermission_PERMISSION_READER,
				},
				{
					UserId:       orchestratortest.MockUserId1,
					ResourceId:   "toe-2",
					ResourceType: orchestrator.ObjectType_OBJECT_TYPE_TARGET_OF_EVALUATION,
					Permission:   orchestrator.UserPermission_PERMISSION_ADMIN,
				},
			},
			args: args{
				userId:     orchestratortest.MockUserId1,
				permission: orchestrator.UserPermission_PERMISSION_READER,
				objectType: orchestrator.ObjectType_OBJECT_TYPE_TARGET_OF_EVALUATION,
			},
			want:    []string{orchestratortest.MockTargetOfEvaluation1.Id, "toe-2"},
			wantErr: assert.NoError,
		},
		{
			name: "excludes resources with insufficient permission",
			permissions: []*orchestrator.UserPermission{
				{
					UserId:       orchestratortest.MockUserId1,
					ResourceId:   orchestratortest.MockTargetOfEvaluation1.Id,
					ResourceType: orchestrator.ObjectType_OBJECT_TYPE_TARGET_OF_EVALUATION,
					Permission:   orchestrator.UserPermission_PERMISSION_READER,
				},
			},
			args: args{
				userId:     orchestratortest.MockUserId1,
				permission: orchestrator.UserPermission_PERMISSION_ADMIN,
				objectType: orchestrator.ObjectType_OBJECT_TYPE_TARGET_OF_EVALUATION,
			},
			want:    nil,
			wantErr: assert.NoError,
		},
		{
			name: "excludes other users' permissions",
			permissions: []*orchestrator.UserPermission{
				{
					UserId:       "other-user",
					ResourceId:   orchestratortest.MockTargetOfEvaluation1.Id,
					ResourceType: orchestrator.ObjectType_OBJECT_TYPE_TARGET_OF_EVALUATION,
					Permission:   orchestrator.UserPermission_PERMISSION_ADMIN,
				},
			},
			args: args{
				userId:     orchestratortest.MockUserId1,
				permission: orchestrator.UserPermission_PERMISSION_READER,
				objectType: orchestrator.ObjectType_OBJECT_TYPE_TARGET_OF_EVALUATION,
			},
			want:    nil,
			wantErr: assert.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ps := newPermissionStoreForTest(t, tt.permissions, tt.listErr)
			got, err := ps.PermissionForResources(context.Background(), tt.args.userId, tt.args.permission, tt.args.reqType, tt.args.objectType)
			assert.Equal(t, tt.want, got)
			tt.wantErr(t, err)
		})
	}
}
