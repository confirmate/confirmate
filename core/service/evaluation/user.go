package evaluation

import (
	"context"

	"confirmate.io/core/api/orchestrator"
	"confirmate.io/core/auth"
	"confirmate.io/core/service"
)

// checkAccess is a convenience wrapper around [service.AuthorizationStrategy.CheckAccess] that
// extracts the user ID from the request context claims. Unlike the orchestrator's equivalent
// function, it does not perform JIT user provisioning since the evaluation service has no database.
// A nil authz defaults to allow-all (same behaviour as [service.CheckAccess]).
func checkAccess(ctx context.Context, authz service.AuthorizationStrategy, reqType orchestrator.RequestType, resourceId string, objectType orchestrator.ObjectType) (bool, []string, error) {
	if authz == nil {
		return true, nil, nil
	}

	claims, _ := auth.ClaimsFromContext(ctx)
	userId := auth.GetConfirmateUserIDFromClaims(claims)

	allowed, resourceIDs := authz.CheckAccess(ctx, userId, reqType, orchestrator.UserPermission_PERMISSION_READER, resourceId, objectType)

	return allowed, resourceIDs, nil
}
