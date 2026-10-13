package service

import (
	"context"
	"slices"

	"confirmate.io/core/api"
	"confirmate.io/core/api/orchestrator"
	"confirmate.io/core/api/orchestrator/orchestratorconnect"

	"connectrpc.com/connect"
)

// OrchestratorPermissionStore implements [PermissionStore] by querying the orchestrator via its
// API. It is intended for services that do not have direct database access (e.g. the evaluation
// service) and need to delegate permission checks to the orchestrator over a service-to-service
// connection.
type OrchestratorPermissionStore struct {
	Client orchestratorconnect.OrchestratorClient
}

// HasPermission checks whether the given user has at least the required permission for the given object.
func (ps *OrchestratorPermissionStore) HasPermission(ctx context.Context, userId string, objectId string, permission orchestrator.UserPermission_Permission, _ orchestrator.RequestType, objectType orchestrator.ObjectType) (bool, error) {
	var (
		permissions []*orchestrator.UserPermission
		err         error
	)

	permissions, err = ps.listPermissions(ctx, userId)
	if err != nil {
		return false, err
	}

	for _, p := range permissions {
		if p.GetObjectId() == objectId && p.GetObjectType() == objectType && p.GetPermission() >= permission {
			return true, nil
		}
	}

	// Permissions on a target of evaluation are inherited by its audit scopes.
	if objectType != orchestrator.ObjectType_OBJECT_TYPE_AUDIT_SCOPE {
		return false, nil
	}

	scope, err := ps.Client.GetAuditScope(ctx, connect.NewRequest(&orchestrator.GetAuditScopeRequest{
		AuditScopeId: objectId,
	}))
	if connect.CodeOf(err) == connect.CodeNotFound {
		return false, nil
	} else if err != nil {
		return false, err
	}

	for _, p := range permissions {
		if p.GetObjectId() == scope.Msg.GetTargetOfEvaluationId() &&
			p.GetObjectType() == orchestrator.ObjectType_OBJECT_TYPE_TARGET_OF_EVALUATION &&
			p.GetPermission() >= permission {
			return true, nil
		}
	}

	return false, nil
}

// PermissionForObjects returns the object IDs for which the given user has at least the required permission.
func (ps *OrchestratorPermissionStore) PermissionForObjects(ctx context.Context, userId string, permission orchestrator.UserPermission_Permission, _ orchestrator.RequestType, objectType orchestrator.ObjectType) ([]string, error) {
	var (
		permissions []*orchestrator.UserPermission
		ids         []string
		err         error
	)

	permissions, err = ps.listPermissions(ctx, userId)
	if err != nil {
		return nil, err
	}

	for _, p := range permissions {
		if p.GetObjectType() == objectType && p.GetPermission() >= permission {
			ids = append(ids, p.GetObjectId())
		}
	}

	if objectType != orchestrator.ObjectType_OBJECT_TYPE_AUDIT_SCOPE {
		return ids, nil
	}

	// Permissions on a target of evaluation are inherited by its audit scopes, so we add all audit
	// scopes of the targets of evaluation the user has (at least) the requested permission for.
	for _, p := range permissions {
		if p.GetObjectType() != orchestrator.ObjectType_OBJECT_TYPE_TARGET_OF_EVALUATION || p.GetPermission() < permission {
			continue
		}

		toeId := p.GetObjectId()
		scopes, err := api.ListAllPaginated(ctx, &orchestrator.ListAuditScopesRequest{
			Filter: &orchestrator.ListAuditScopesRequest_Filter{TargetOfEvaluationId: &toeId},
		}, func(ctx context.Context, req *orchestrator.ListAuditScopesRequest) (*orchestrator.ListAuditScopesResponse, error) {
			res, err := ps.Client.ListAuditScopes(ctx, connect.NewRequest(req))
			if err != nil {
				return nil, err
			}
			return res.Msg, nil
		}, func(res *orchestrator.ListAuditScopesResponse) []*orchestrator.AuditScope {
			return res.GetAuditScopes()
		})
		if err != nil {
			return nil, err
		}

		for _, scope := range scopes {
			if !slices.Contains(ids, scope.GetId()) {
				ids = append(ids, scope.GetId())
			}
		}
	}

	return ids, nil
}

// listPermissions fetches all permissions for the given user from the orchestrator. This call goes
// through the service-to-service HTTP client, which injects the service OAuth2 token, so the
// orchestrator treats it as an admin request.
func (ps *OrchestratorPermissionStore) listPermissions(ctx context.Context, userId string) ([]*orchestrator.UserPermission, error) {
	req := &orchestrator.ListUserPermissionsRequest{
		Filter: &orchestrator.ListUserPermissionsRequest_Filter{},
	}
	if userId != "" {
		req.Filter.UserId = &userId
	}
	return api.ListAllPaginated(ctx, req, func(ctx context.Context, req *orchestrator.ListUserPermissionsRequest) (*orchestrator.ListUserPermissionsResponse, error) {
		var (
			res *connect.Response[orchestrator.ListUserPermissionsResponse]
			err error
		)

		res, err = ps.Client.ListUserPermissions(ctx, connect.NewRequest(req))
		if err != nil {
			return nil, err
		}

		return res.Msg, nil
	}, func(res *orchestrator.ListUserPermissionsResponse) []*orchestrator.UserPermission {
		return res.GetUserPermissions()
	})
}
