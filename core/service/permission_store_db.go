package service

import (
	"context"
	"fmt"

	"confirmate.io/core/api/orchestrator"
	"confirmate.io/core/persistence"
)

// DBPermissionStore implements [PermissionStore] by querying the database directly. It is intended
// for services that have direct database access (e.g. the orchestrator service).
type DBPermissionStore struct {
	DB persistence.DB
}

// HasPermission checks if the given user has the specified permission for the object.
func (ps DBPermissionStore) HasPermission(_ context.Context, userId string, objectId string, permission orchestrator.UserPermission_Permission, _ orchestrator.RequestType, objectType orchestrator.ObjectType) (bool, error) {
	var (
		count          int64
		err            error
		userPermission orchestrator.UserPermission
	)

	count, err = ps.DB.Count(
		&userPermission,
		"user_id = ? AND object_type = ? AND object_id = ? AND permission IN (?)",
		userId, objectType, objectId, allowedPermissions(permission),
	)
	if err != nil {
		return false, fmt.Errorf("failed to check permissions: %w", err)
	}

	return count > 0, nil
}

// PermissionForObjects returns a list of object IDs for which the given user has at least the specified permission.
func (ps DBPermissionStore) PermissionForObjects(_ context.Context, userID string, permission orchestrator.UserPermission_Permission, _ orchestrator.RequestType, objectType orchestrator.ObjectType) ([]string, error) {
	var (
		userPermissions []orchestrator.UserPermission
		err             error
	)

	types := []orchestrator.ObjectType{orchestrator.ObjectType_OBJECT_TYPE_USER_PERMISSION}

	if objectType == orchestrator.ObjectType_OBJECT_TYPE_USER_PERMISSION {
		if objectType == orchestrator.ObjectType_OBJECT_TYPE_USER_PERMISSION {
			types = []orchestrator.ObjectType{orchestrator.ObjectType_OBJECT_TYPE_TARGET_OF_EVALUATION, orchestrator.ObjectType_OBJECT_TYPE_AUDIT_SCOPE}
		}
	}

	err = ps.DB.List(
		&userPermissions,
		"object_id",
		true,
		0,
		-1,
		"user_id = ? AND object_type IN (?) AND permission IN (?)",
		userID,
		types,
		allowedPermissions(permission),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve permissions: %w", err)
	}

	objectIds := make([]string, len(userPermissions))
	for i := range userPermissions {
		objectIds[i] = userPermissions[i].ObjectId
	}

	return objectIds, nil
}

// allowedPermissions returns the set of permission levels that satisfy the required permission,
// including higher levels (ADMIN > CONTRIBUTOR > READER).
func allowedPermissions(required orchestrator.UserPermission_Permission) []orchestrator.UserPermission_Permission {
	switch required {
	case orchestrator.UserPermission_PERMISSION_READER:
		return []orchestrator.UserPermission_Permission{
			orchestrator.UserPermission_PERMISSION_READER,
			orchestrator.UserPermission_PERMISSION_CONTRIBUTOR,
			orchestrator.UserPermission_PERMISSION_ADMIN,
		}
	case orchestrator.UserPermission_PERMISSION_CONTRIBUTOR:
		return []orchestrator.UserPermission_Permission{
			orchestrator.UserPermission_PERMISSION_CONTRIBUTOR,
			orchestrator.UserPermission_PERMISSION_ADMIN,
		}
	default:
		return []orchestrator.UserPermission_Permission{
			orchestrator.UserPermission_PERMISSION_ADMIN,
		}
	}
}
