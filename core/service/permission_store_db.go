package service

import (
	"context"
	"errors"
	"fmt"
	"slices"

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

	// Note: We do not need the object type here. If the endpoint UpsertUserPermission is used, the object type will be set to OBJECT_TYPE_USER_PERMISSION, but the object ID will be the ID of the target object. Therefore, we can ignore the object type in this query. We assume that the combination of object id (UUID) and user id (UUID) is unique across all object types.
	count, err = ps.DB.Count(
		&userPermission,
		"user_id = ? AND object_id = ? AND permission IN ?",
		userId, objectId, allowedPermissions(permission),
	)
	if err != nil {
		return false, fmt.Errorf("failed to check permissions: %w", err)
	}
	if count > 0 {
		return true, nil
	}

	// Permissions on a target of evaluation are inherited by its audit scopes. The object ID might
	// refer to an audit scope (either directly or via OBJECT_TYPE_USER_PERMISSION), so we check the
	// permission of the user on the audit scope's target of evaluation as well.
	if objectType != orchestrator.ObjectType_OBJECT_TYPE_AUDIT_SCOPE &&
		objectType != orchestrator.ObjectType_OBJECT_TYPE_USER_PERMISSION {
		return false, nil
	}

	var scope orchestrator.AuditScope
	err = ps.DB.Get(&scope, persistence.WithoutPreload(), "id = ?", objectId)
	if errors.Is(err, persistence.ErrRecordNotFound) {
		return false, nil
	} else if err != nil {
		return false, fmt.Errorf("failed to retrieve audit scope: %w", err)
	}

	count, err = ps.DB.Count(
		&userPermission,
		"user_id = ? AND object_id = ? AND object_type = ? AND permission IN ?",
		userId, scope.GetTargetOfEvaluationId(), orchestrator.ObjectType_OBJECT_TYPE_TARGET_OF_EVALUATION, allowedPermissions(permission),
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

	types := []orchestrator.ObjectType{objectType}

	if objectType == orchestrator.ObjectType_OBJECT_TYPE_USER_PERMISSION {
		types = []orchestrator.ObjectType{orchestrator.ObjectType_OBJECT_TYPE_TARGET_OF_EVALUATION, orchestrator.ObjectType_OBJECT_TYPE_AUDIT_SCOPE}
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

	objectIds := make([]string, 0, len(userPermissions))
	toeIds := make([]string, 0)
	for i := range userPermissions {
		objectIds = append(objectIds, userPermissions[i].ObjectId)
		if userPermissions[i].ObjectType == orchestrator.ObjectType_OBJECT_TYPE_TARGET_OF_EVALUATION {
			toeIds = append(toeIds, userPermissions[i].ObjectId)
		}
	}

	if !slices.Contains(types, orchestrator.ObjectType_OBJECT_TYPE_AUDIT_SCOPE) {
		return objectIds, nil
	}

	// Permissions on a target of evaluation are inherited by its audit scopes, so we add all audit
	// scopes of the targets of evaluation the user has (at least) the requested permission for.
	if objectType == orchestrator.ObjectType_OBJECT_TYPE_AUDIT_SCOPE {
		err = ps.DB.List(
			&userPermissions,
			"object_id",
			true,
			0,
			-1,
			"user_id = ? AND object_type = ? AND permission IN (?)",
			userID,
			orchestrator.ObjectType_OBJECT_TYPE_TARGET_OF_EVALUATION,
			allowedPermissions(permission),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve permissions: %w", err)
		}

		for i := range userPermissions {
			toeIds = append(toeIds, userPermissions[i].ObjectId)
		}
	}

	if len(toeIds) == 0 {
		return objectIds, nil
	}

	var scopes []orchestrator.AuditScope
	err = ps.DB.List(&scopes, "id", true, 0, -1, persistence.WithoutPreload(), "target_of_evaluation_id IN (?)", toeIds)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve audit scopes: %w", err)
	}

	for i := range scopes {
		if !slices.Contains(objectIds, scopes[i].Id) {
			objectIds = append(objectIds, scopes[i].Id)
		}
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
