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

const (
	// DefaultAllowAllClaim is the default claim key granting access to all TOEs.
	DefaultAllowAllClaim = "cladmin"
)

// ErrPermissionDenied represents an error, where permission to fulfill the request is denied.
var ErrPermissionDenied = connect.NewError(connect.CodePermissionDenied, errors.New("access denied"))

// AuthorizationStrategy implements access checks based on the request and context.
type AuthorizationStrategy interface {
	CheckAccess(ctx context.Context,
		userId string,
		reqType orchestrator.RequestType,
		resourceType orchestrator.UserPermission_ResourceType,
		userPermission orchestrator.UserPermission_Permission,
		resourceId string) (bool, []string)
}

// CheckAccess checks access via the configured strategy.
//
// If no authorization strategy is configured (i.e., authz is nil), it defaults to allowing all access, returning true and nil for resource IDs. This design choice ensures that in the absence of an explicit strategy, the system remains permissive by default, which can be useful during development or in scenarios where access control is not a concern. However, in production environments, it is recommended to configure an appropriate authorization strategy to enforce access control policies effectively.
func CheckAccess[T any](authz AuthorizationStrategy,
	ctx context.Context,
	userId string,
	reqType orchestrator.RequestType,
	resource_type orchestrator.UserPermission_ResourceType,
	userPermission orchestrator.UserPermission_Permission,
	resourceId string) (bool, []string) {
	if authz == nil {
		return true, nil
	}

	return authz.CheckAccess(ctx,
		userId,
		reqType,
		resource_type,
		userPermission,
		resourceId)
}

type AuthorizationStrategyJWT struct {
	AllowAllKey string

	// Permissions is an interface to check user permissions stored in the Orchestrator DB. It is used as part of the JWT-based authorization strategy to determine access rights based on the user's permissions.
	Permissions PermissionStore
}

// CheckAccess checks whether the request can be fulfilled using the current access strategy.
func (a *AuthorizationStrategyJWT) CheckAccess(ctx context.Context,
	userId string,
	reqType orchestrator.RequestType,
	resourceType orchestrator.UserPermission_ResourceType,
	userPermission orchestrator.UserPermission_Permission,
	resourceId string) (allowed bool, resourceIDs []string) {
	var (
		err error
	)

	if a == nil || userId == "" {
		return false, nil
	}

	// Check AllowAllKey claim to allow access to all (e.g., {"cladmin": true}).
	if claims, ok := auth.ClaimsFromContext(ctx); ok && a.AllowAllKey != "" {
		if b, ok := claims["sub"].(string); ok && b == a.AllowAllKey {
			return true, nil
		}
	}

	if a.Permissions == nil {
		slog.Error("Permission store is not configured for JWT authorization strategy")
		return false, nil
	}

	// For list requests, we retrieve the list of resource IDs the user has reader permissions for and return it.
	if reqType == orchestrator.RequestType_REQUEST_TYPE_LIST {
		resourceIDs, err = a.Permissions.PermissionForResources(ctx,
			userId,
			resourceType,
			orchestrator.UserPermission_PERMISSION_READER,
		)
		if err != nil {
			return false, nil
		}

		return false, resourceIDs
	}

	// For non-list requests, we check if the user has reader permissions for the specific resource ID.
	if resourceId == "" {
		return false, nil
	}

	allowed, err = a.Permissions.HasPermission(ctx,
		userId,
		resourceType,
		resourceId,
		orchestrator.UserPermission_PERMISSION_READER,
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
	_ orchestrator.UserPermission_ResourceType,
	_ orchestrator.UserPermission_Permission,
	_ string) (ok bool, resourceIDs []string) {
	// Keep this strategy permissive by design.
	return true, nil
}

// GetClaim is a helper function to extract a specific claim from the claims map, returning an empty string if the claim is not present or not a string.
func GetClaim(claims map[string]any, key string) string {
	if val, ok := claims[key]; ok && val != nil {
		return val.(string)
	}
	return ""
}

// GetClaimInt64 extracts a numeric claim as int64, returning 0 if missing or not numeric.
func GetClaimInt64(claims map[string]any, key string) int64 {
	val, ok := claims[key]
	if !ok || val == nil {
		return 0
	}

	switch v := val.(type) {
	case int64:
		return v
	case int:
		return int64(v)
	case int32:
		return int64(v)
	case float64:
		return int64(v)
	case float32:
		return int64(v)
	case uint64:
		return int64(v)
	case uint:
		return int64(v)
	case uint32:
		return int64(v)
	default:
		return 0
	}
}
