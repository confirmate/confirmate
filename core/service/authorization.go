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

	"confirmate.io/core/api"
	"confirmate.io/core/api/orchestrator"
	"confirmate.io/core/auth"

	"connectrpc.com/connect"
)

const (
	// DefaultTargetOfEvaluationsClaim is the default claim key containing allowed TOE IDs.
	DefaultTargetOfEvaluationsClaim = "TargetOfEvaluationid"
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
		req api.HasTargetOfEvaluationId,
		resourceId string) bool
	AllowedTargetOfEvaluations(ctx context.Context) (all bool, IDs []string)
}

// CheckAccess checks access via the configured strategy.
//
// If no strategy is configured, access is allowed by default. It uses [resolveTargetOfEvaluationID]
// to extract the target_of_evaluation_id from either the request itself or its payload.
func CheckAccess[T any](authz AuthorizationStrategy,
	ctx context.Context,
	userId string,
	reqType orchestrator.RequestType,
	resource_type orchestrator.UserPermission_ResourceType,
	userPermission orchestrator.UserPermission_Permission,
	req *connect.Request[T],
	resourceId string) bool {
	if authz == nil {
		return true
	}

	if req == nil {
		return authz.CheckAccess(ctx,
			userId,
			reqType,
			resource_type,
			orchestrator.UserPermission_Permission(0),
			nil,
			resourceId)
	}

	return authz.CheckAccess(ctx,
		userId,
		reqType,
		resource_type,
		orchestrator.UserPermission_Permission(0),
		nil, //esolveTargetOfEvaluationID(req.Any()), TODO(all): Do we need that anymore?
		resourceId)
}

// resolveTargetOfEvaluationID attempts to extract a target_of_evaluation_id from the request (if it
// implements [api.HasTargetOfEvaluationId]) or its payload (if it is an [api.PayloadRequest]).
func resolveTargetOfEvaluationID(req any) api.HasTargetOfEvaluationId {
	if req == nil {
		return nil
	}

	withTargetOfEvaluationID, ok := req.(api.HasTargetOfEvaluationId)
	if ok {
		return withTargetOfEvaluationID
	}

	payloadReq, ok := req.(api.PayloadRequest)
	if !ok {
		return nil
	}

	payload := payloadReq.GetPayload()
	if payload == nil {
		return nil
	}

	withTargetOfEvaluationID, ok = payload.(api.HasTargetOfEvaluationId)
	if ok {
		return withTargetOfEvaluationID
	}

	return nil
}

// AuthorizationStrategyJWT expects a list of TOE IDs in a JWT claim key.
type AuthorizationStrategyJWT struct {
	TargetOfEvaluationsKey string
	AllowAllKey            string

	// Permissions is optional. If set, the strategy can check fine-grained permissions in Orchestrator DB in addition to JWT claims.
	Permissions PermissionStore
}

// CheckAccess checks whether the request can be fulfilled using the current access strategy.
func (a *AuthorizationStrategyJWT) CheckAccess(ctx context.Context,
	userId string,
	reqType orchestrator.RequestType,
	resourceType orchestrator.UserPermission_ResourceType,
	userPermission orchestrator.UserPermission_Permission,
	req api.HasTargetOfEvaluationId,
	resourceId string) (allowed bool) {
	var (
		err error
	)

	if a == nil || resourceId == "" || userId == "" {
		return false
	}

	// TODO(@anatheka): Here we have to check the permissions, but are not able to store/get the information from the DB

	// Check AllowAllKey claim to allow access to all (e.g., {"cladmin": true}).
	if claims, ok := auth.ClaimsFromContext(ctx); ok && a.AllowAllKey != "" {
		if b, ok := claims[a.AllowAllKey].(bool); ok && b {
			return true
		}
	}

	// Check permission based on the permissions stored in the Orchestrator DB.
	// Check permissions stored in the Orchestrator DB
	if a.Permissions == nil {
		return false
	}

	// TODO(@anatheka): Update
	allowed, err = a.Permissions.HasPermission(ctx,
		userId,
		resourceType,
		resourceId,
		orchestrator.UserPermission_PERMISSION_READER,
	)
	if err != nil {
		return false
	}

	return allowed
}

func (a *AuthorizationStrategyJWT) AllowedTargetOfEvaluations(ctx context.Context) (all bool, list []string) {
	return true, nil
}

// TODO(anatheka): Deprecated?
// // CheckAccess checks whether the request can be fulfilled using the current access strategy.
// func (a *AuthorizationStrategyJWT) CheckAccess(ctx context.Context, _ orchestrator.RequestType, req api.HasTargetOfEvaluationId) (ok bool) {
// 	var (
// 		all  bool
// 		list []string
// 	)

// 	if a == nil {
// 		return false
// 	}

// 	all, list = a.AllowedTargetOfEvaluations(ctx)
// 	if all {
// 		return true
// 	}

// 	if req == nil {
// 		return false
// 	}

// 	ok = slices.Contains(list, req.GetTargetOfEvaluationId())
// 	return ok
// }

// TODO(anatheka): Deprecated?
// // AllowedTargetOfEvaluations retrieves a list of allowed TOE IDs according to the current access strategy.
// func (a *AuthorizationStrategyJWT) AllowedTargetOfEvaluations(ctx context.Context) (all bool, list []string) {
// 	var (
// 		claims  jwt.MapClaims
// 		ok      bool
// 		rawList any
// 	)

// 	if a == nil {
// 		return false, nil
// 	}

// 	if ctx == nil {
// 		return false, nil
// 	}

// 	claims, ok = auth.ClaimsFromContext(ctx)
// 	if !ok {
// 		return false, nil
// 	}

// 	// if b, ok := claims[a.AllowAllKey].(bool); ok && b {
// 	// 	return true, nil
// 	// }
// 	if b, ok := claims["sub"].(string); ok && b == a.AllowAllKey {
// 		return true, nil
// 	}

// 	rawList, ok = claims[a.TargetOfEvaluationsKey]
// 	if !ok {
// 		return false, nil
// 	}

// 	switch v := rawList.(type) {
// 	case []interface{}:
// 		for _, item := range v {
// 			if s, ok := item.(string); ok {
// 				list = append(list, s)
// 			}
// 		}
// 	case []string:
// 		list = append(list, v...)
// 	}

// 	return false, list
// }

// AuthorizationStrategyAllowAll allows all requests.
type AuthorizationStrategyAllowAll struct{}

// CheckAccess returns true for all requests.
func (*AuthorizationStrategyAllowAll) CheckAccess(_ context.Context,
	_ string,
	_ orchestrator.RequestType,
	_ orchestrator.UserPermission_ResourceType,
	_ orchestrator.UserPermission_Permission,
	_ api.HasTargetOfEvaluationId,
	_ string) (ok bool) {
	// Keep this strategy permissive by design.
	return true
}

// AllowedTargetOfEvaluations returns all = true for this strategy.
func (*AuthorizationStrategyAllowAll) AllowedTargetOfEvaluations(_ context.Context) (all bool, list []string) {
	return true, nil
}

// getClaim is a helper function to extract a specific claim from the claims map, returning an empty string if the claim is not present or not a string.
func GetClaim(claims map[string]any, key string) string {
	if val, ok := claims[key]; ok && val != nil {
		return val.(string)
	}
	return ""
}

// getClaimInt64 extracts a numeric claim as int64, returning 0 if missing or not numeric.
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
