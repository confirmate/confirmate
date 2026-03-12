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

package auth

import (
	"context"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const (
	claimsContextKey contextKey = "auth-claims"
)

// WithClaims stores verified JWT claims in the context.
func WithClaims(ctx context.Context, claims jwt.MapClaims) (out context.Context) {
	if ctx == nil || claims == nil {
		return ctx
	}

	out = context.WithValue(ctx, claimsContextKey, claims)
	return out
}

// ClaimsFromContext returns verified JWT claims from the context, if present.
func ClaimsFromContext(ctx context.Context) (claims jwt.MapClaims, ok bool) {
	if ctx == nil {
		return nil, false
	}

	claims, ok = ctx.Value(claimsContextKey).(jwt.MapClaims)
	return claims, ok
}
