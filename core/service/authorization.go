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
	"confirmate.io/core/auth"

	"connectrpc.com/connect"
	"github.com/golang-jwt/jwt/v5"
)

// RequestType specifies the type of request, usually CRUD.
type RequestType int

const (
	AccessCreate RequestType = iota
	AccessRead
	AccessUpdate
	AccessDelete
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
	CheckAccess(ctx context.Context, typ RequestType, req api.TargetOfEvaluationRequest) bool
	AllowedTargetOfEvaluations(ctx context.Context) (all bool, IDs []string)
}

// AuthorizationStrategyJWT expects a list of TOE IDs in a JWT claim key.
type AuthorizationStrategyJWT struct {
	TargetOfEvaluationsKey string
	AllowAllKey            string
}

// CheckAccess checks whether the request can be fulfilled using the current access strategy.
func (a *AuthorizationStrategyJWT) CheckAccess(ctx context.Context, _ RequestType, req api.TargetOfEvaluationRequest) (ok bool) {
	var (
		all  bool
		list []string
	)

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
		token   string
		claims  jwt.MapClaims
		ok      bool
		err     error
		rawList any
	)

	if ctx == nil {
		return false, nil
	}

	token, ok = auth.TokenFromContext(ctx)
	if !ok || token == "" {
		return false, nil
	}

	var parser *jwt.Parser
	parser = jwt.NewParser()
	_, _, err = parser.ParseUnverified(token, &claims)
	if err != nil {
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
func (*AuthorizationStrategyAllowAll) CheckAccess(_ context.Context, _ RequestType, _ api.TargetOfEvaluationRequest) (ok bool) {
	return true
}

// AllowedTargetOfEvaluations returns all = true for this strategy.
func (*AuthorizationStrategyAllowAll) AllowedTargetOfEvaluations(_ context.Context) (all bool, list []string) {
	return true, nil
}
