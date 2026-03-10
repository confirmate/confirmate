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
	"slices"

	"confirmate.io/core/api"
	"confirmate.io/core/api/orchestrator"
	"confirmate.io/core/auth"

	"connectrpc.com/connect"
	"github.com/golang-jwt/jwt/v5"
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
	CheckAccess(ctx context.Context, typ orchestrator.RequestType, req api.HasTargetOfEvaluationId) bool
	AllowedTargetOfEvaluations(ctx context.Context) (all bool, IDs []string)
}

// AuthorizationStrategyJWT expects a list of TOE IDs in a JWT claim key.
type AuthorizationStrategyJWT struct {
	TargetOfEvaluationsKey string
	AllowAllKey            string
}

// CheckAccess checks whether the request can be fulfilled using the current access strategy.
func (a *AuthorizationStrategyJWT) CheckAccess(ctx context.Context, _ orchestrator.RequestType, req api.HasTargetOfEvaluationId) (ok bool) {
	var (
		all  bool
		list []string
	)

	if a == nil {
		return false
	}

	all, list = a.AllowedTargetOfEvaluations(ctx)
	if all {
		return true
	}

	if req == nil {
		return false
	}

	ok = slices.Contains(list, req.GetTargetOfEvaluationId())
	return ok
}

// AllowedTargetOfEvaluations retrieves a list of allowed TOE IDs according to the current access strategy.
func (a *AuthorizationStrategyJWT) AllowedTargetOfEvaluations(ctx context.Context) (all bool, list []string) {
	var (
		claims  jwt.MapClaims
		ok      bool
		rawList any
	)

	if a == nil {
		return false, nil
	}

	if ctx == nil {
		return false, nil
	}

	claims, ok = auth.ClaimsFromContext(ctx)
	if !ok {
		return false, nil
	}

	if b, ok := claims[a.AllowAllKey].(bool); ok && b {
		return true, nil
	}

	rawList, ok = claims[a.TargetOfEvaluationsKey]
	if !ok {
		return false, nil
	}

	switch v := rawList.(type) {
	case []interface{}:
		for _, item := range v {
			if s, ok := item.(string); ok {
				list = append(list, s)
			}
		}
	case []string:
		list = append(list, v...)
	}

	return false, list
}

// AuthorizationStrategyAllowAll allows all requests.
type AuthorizationStrategyAllowAll struct{}

// CheckAccess returns true for all requests.
func (*AuthorizationStrategyAllowAll) CheckAccess(_ context.Context, _ orchestrator.RequestType, _ api.HasTargetOfEvaluationId) (ok bool) {
	// Keep this strategy permissive by design.
	return true
}

// AllowedTargetOfEvaluations returns all = true for this strategy.
func (*AuthorizationStrategyAllowAll) AllowedTargetOfEvaluations(_ context.Context) (all bool, list []string) {
	return true, nil
}
