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
)

type contextKey string

const (
	claimsContextKey contextKey = "auth-claims"
)

// WithClaims stores verified JWT claims in the context.
func WithClaims(ctx context.Context, claims *OAuthClaims) (out context.Context) {
	if ctx == nil || claims == nil {
		return ctx
	}

	out = context.WithValue(ctx, claimsContextKey, claims)
	return out
}

// ClaimsFromContext returns verified JWT claims from the context, if present.
func ClaimsFromContext(ctx context.Context) (claims *OAuthClaims, ok bool) {
	if ctx == nil {
		return nil, false
	}

	claims, ok = ctx.Value(claimsContextKey).(*OAuthClaims)
	return claims, ok
}
