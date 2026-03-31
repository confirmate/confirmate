package service

import (
	"context"

	"confirmate.io/core/api/orchestrator"
)

// PermissionStore
type PermissionStore interface {
	HasPermission(ctx context.Context, userId string, resourceType orchestrator.UserPermission_ResourceType, resourceID string, permission orchestrator.UserPermission_Permission) (bool, error)
	PermissionForResources(ctx context.Context, userId string, resourceType orchestrator.UserPermission_ResourceType, permission orchestrator.UserPermission_Permission) ([]string, error)
}
