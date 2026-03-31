package service

import (
	"context"

	"confirmate.io/core/api/orchestrator"
)

// PermissionStore is an interface that defines methods for checking user permissions and retrieving resources based on permissions. It is used by the AuthorizationStrategyJWT to determine if a user has the necessary permissions to access a resource and to fetch the resources that a user has permissions for.
type PermissionStore interface {
	HasPermission(ctx context.Context, userId string, resourceType orchestrator.UserPermission_ResourceType, resourceID string, permission orchestrator.UserPermission_Permission) (bool, error)
	PermissionForResources(ctx context.Context, userId string, resourceType orchestrator.UserPermission_ResourceType, permission orchestrator.UserPermission_Permission) ([]string, error)
}
