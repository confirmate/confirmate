// Copyright 2016-2026 Fraunhofer AISEC
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

package service

import (
	"context"
	"errors"
	"log/slog"

	"confirmate.io/core/api/orchestrator"
	"confirmate.io/core/auth"

	"connectrpc.com/connect"
)

// ErrPermissionDenied represents an error, where permission to fulfill the request is denied.
var ErrPermissionDenied = connect.NewError(connect.CodePermissionDenied, errors.New("access denied"))

// AuthorizationStrategy implements access checks based on the request and context.
type AuthorizationStrategy interface {
	CheckAccess(ctx context.Context,
		userId string,
		reqType orchestrator.RequestType,
		userPermission orchestrator.UserPermission_Permission,
		resourceId string,
		objectType orchestrator.ObjectType,
	) (bool, []string)
}

// CheckAccess checks access via the configured strategy.
//
// If no authorization strategy is configured (i.e., authz is nil), it defaults to allowing all access, returning true and nil for resource IDs. This design choice ensures that in the absence of an explicit strategy, the system remains permissive by default, which can be useful during development or in scenarios where access control is not a concern. However, in production environments, it is recommended to configure an appropriate authorization strategy to enforce access control policies effectively.
func CheckAccess[T any](authz AuthorizationStrategy,
	ctx context.Context,
	userId string,
	reqType orchestrator.RequestType,
	userPermission orchestrator.UserPermission_Permission,
	resourceId string,
	objectType orchestrator.ObjectType) (bool, []string) {
	if authz == nil {
		return true, nil
	}

	return authz.CheckAccess(ctx,
		userId,
		reqType,
		userPermission,
		resourceId,
		objectType)
}

// AuthorizationStrategyPermissionStore implements access checks based on user permissions stored in
// a [PermissionStore]. It checks permissions for the user making the request and the requested
// resource, returning whether access is allowed and, for list requests, the IDs of resources the
// user has access to.
type AuthorizationStrategyPermissionStore struct {
	// Permissions is an interface to check user permissions stored in the Orchestrator DB. It is used as part of the JWT-based authorization strategy to determine access rights based on the user's permissions.
	Permissions PermissionStore
}

// CheckAccess checks whether the request can be fulfilled using the current access strategy.
func (a *AuthorizationStrategyPermissionStore) CheckAccess(ctx context.Context,
	userId string,
	reqType orchestrator.RequestType,
	userPermission orchestrator.UserPermission_Permission,
	resourceId string,
	objectType orchestrator.ObjectType) (allowed bool, resourceIDs []string) {
	var (
		err            error
		objectTypeUsed orchestrator.ObjectType
	)

	if a == nil || userId == "" {
		return false, nil
	}

	// Check IsAdminToken claim to allow access to all.
	if claims, ok := auth.ClaimsFromContext(ctx); ok && claims.IsAdminToken {
		return true, nil
	}

	if a.Permissions == nil {
		slog.Error("Permission store is not configured for JWT authorization strategy")
		return false, nil
	}

	// Check if ToE ID or Audit Scope ID is necessary for the permission check; return false if not provided.
	switch objectType {
	case orchestrator.ObjectType_OBJECT_TYPE_TARGET_OF_EVALUATION,
		orchestrator.ObjectType_OBJECT_TYPE_EVIDENCE,
		orchestrator.ObjectType_OBJECT_TYPE_ASSESSMENT_RESULT,
		orchestrator.ObjectType_OBJECT_TYPE_METRIC_CONFIGURATION:
		objectTypeUsed = orchestrator.ObjectType_OBJECT_TYPE_TARGET_OF_EVALUATION
	case orchestrator.ObjectType_OBJECT_TYPE_AUDIT_SCOPE,
		orchestrator.ObjectType_OBJECT_TYPE_CERTIFICATE:
		objectTypeUsed = orchestrator.ObjectType_OBJECT_TYPE_TARGET_OF_EVALUATION
	case orchestrator.ObjectType_OBJECT_TYPE_EVALUATION_RESULT:
		objectTypeUsed = orchestrator.ObjectType_OBJECT_TYPE_AUDIT_SCOPE
	default:
		slog.Debug("Unsupported object type for permission check", "objectType", objectType)
		return false, nil
	}

	// For list requests, we check if the user has reader permissions for any resources of the given type and return the list of resource IDs they have access to.
	if reqType == orchestrator.RequestType_REQUEST_TYPE_LIST {
		resourceIDs, err = a.Permissions.PermissionForResources(ctx,
			userId,
			orchestrator.UserPermission_PERMISSION_READER,
			reqType,
			objectTypeUsed,
		)
		if err != nil {
			return false, nil
		}

		return false, resourceIDs
	}

	// For update and delete requests, we check if the user has the required permissions for the specific resource ID.
	if resourceId == "" {
		return false, nil
	}
	switch reqType {
	case orchestrator.RequestType_REQUEST_TYPE_UPDATED:
		userPermission = orchestrator.UserPermission_PERMISSION_CONTRIBUTOR
	case orchestrator.RequestType_REQUEST_TYPE_DELETED:
		userPermission = orchestrator.UserPermission_PERMISSION_ADMIN
	}

	allowed, err = a.Permissions.HasPermission(ctx,
		userId,
		resourceId,
		userPermission,
		reqType,
		objectTypeUsed,
	)
	if err != nil {
		return false, nil
	}

	return allowed, nil
}

// AuthorizationStrategyAllowAll allows all requests.
type AuthorizationStrategyAllowAll struct{}

// CheckAccess returns true for all requests.
func (*AuthorizationStrategyAllowAll) CheckAccess(_ context.Context,
	_ string,
	_ orchestrator.RequestType,
	_ orchestrator.UserPermission_Permission,
	_ string,
	_ orchestrator.ObjectType) (ok bool, resourceIDs []string,
) {
	// Keep this strategy permissive by design.
	return true, nil
}
