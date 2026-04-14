package service

import (
	"context"

	"confirmate.io/core/api/orchestrator"
)

// PermissionStore is an interface that defines methods for checking user permissions and retrieving resources based on permissions. It is used by the AuthorizationStrategyPermissionStore to determine if a user has the necessary permissions to access a resource and to fetch the resources that a user has permissions for.
type PermissionStore interface {
	HasPermission(ctx context.Context, userId string, resourceID string, permission orchestrator.UserPermission_Permission, reqType orchestrator.RequestType, objectType orchestrator.ObjectType) (bool, error)
	PermissionForResources(ctx context.Context, userId string, permission orchestrator.UserPermission_Permission, reqType orchestrator.RequestType, objectType orchestrator.ObjectType) ([]string, error)
}
