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

package server

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"confirmate.io/core/auth"

	"connectrpc.com/connect"
	"github.com/MicahParks/keyfunc/v2"
	"github.com/golang-jwt/jwt/v5"
)

const DefaultJWKSURL = "http://localhost:8080/v1/auth/certs"

// AuthConfig contains parameters needed to configure authentication.
type AuthConfig struct {
	jwksURL string
	useJWKS bool
	jwks    *keyfunc.JWKS

	publicKey *ecdsa.PublicKey

	publicProcedures map[string]struct{}

	// Role extraction configuration
	roleClaimPaths []string

	// roleMapper, when non-nil, is applied to each raw role string after
	// extraction. Returning the empty string drops the role.
	roleMapper RoleMapper
}

// RoleMapper translates a raw role string from the JWT into a canonical role
// string used by the rest of the system (e.g. the orchestrator enum names like
// "ROLE_ADMIN"). Returning "" drops the role.
type RoleMapper func(rawRole string) string

// AuthOption configures the auth middleware.
type AuthOption func(*AuthConfig)

// WithRoleClaimPaths configures where roles are found in the JWT claims.
// Examples:
// - "roles"
// - "realm_access.roles" (Keycloak realm roles)
func WithRoleClaimPaths(paths ...string) AuthOption {
	return func(c *AuthConfig) {
		for _, p := range paths {
			p = strings.TrimSpace(p)
			if p == "" {
				continue
			}
			c.roleClaimPaths = append(c.roleClaimPaths, p)
		}
	}
}

// WithRoleMapper installs a custom translator from raw JWT role strings to the
// canonical role strings used by the rest of the system. See [RoleMapper].
func WithRoleMapper(m RoleMapper) AuthOption {
	return func(c *AuthConfig) {
		c.roleMapper = m
	}
}

// WithJWKS enables JWKS support for token verification.
func WithJWKS(url string) AuthOption {
	return func(c *AuthConfig) {
		c.jwksURL = url
		c.useJWKS = true
	}
}

// WithPublicKey configures a static public key for token verification.
func WithPublicKey(publicKey *ecdsa.PublicKey) AuthOption {
	return func(c *AuthConfig) {
		c.publicKey = publicKey
	}
}

// WithPublicProcedures marks RPC procedures as public (no auth required).
func WithPublicProcedures(procedures ...string) AuthOption {
	return func(c *AuthConfig) {
		if c.publicProcedures == nil {
			c.publicProcedures = make(map[string]struct{})
		}
		for _, p := range procedures {
			c.publicProcedures[p] = struct{}{}
		}
	}
}

// AuthInterceptor authenticates incoming requests using bearer tokens.
type AuthInterceptor struct {
	cfg *AuthConfig
}

// NewAuthInterceptor creates a new auth interceptor.
func NewAuthInterceptor(opts ...AuthOption) (interceptor *AuthInterceptor) {
	var (
		cfg *AuthConfig
	)

	cfg = &AuthConfig{}
	for _, opt := range opts {
		opt(cfg)
	}

	interceptor = &AuthInterceptor{
		cfg: cfg,
	}

	return interceptor
}

// WrapUnary implements the connect interceptor for unary calls.
func (ai *AuthInterceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, req connect.AnyRequest) (res connect.AnyResponse, err error) {
		var token string

		if ai.isPublic(req.Spec().Procedure) {
			return next(ctx, req)
		}

		token, err = bearerToken(req.Header().Get("Authorization"))
		if err != nil {
			return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("invalid auth token"))
		}

		claims, err := ai.parseToken(token)
		if err != nil {
			return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("invalid auth token"))
		}

		// Store claims in ctx
		ctx = auth.WithClaims(ctx, claims)

		return next(ctx, req)
	}
}

// WrapStreamingClient implements the connect interceptor for streaming client calls.
func (ai *AuthInterceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return next
}

// WrapStreamingHandler implements the connect interceptor for streaming handler calls.
func (ai *AuthInterceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return func(ctx context.Context, conn connect.StreamingHandlerConn) (err error) {
		var token string

		if ai.isPublic(conn.Spec().Procedure) {
			return next(ctx, conn)
		}

		token, err = bearerToken(conn.RequestHeader().Get("Authorization"))
		if err != nil {
			return connect.NewError(connect.CodeUnauthenticated, errors.New("invalid auth token"))
		}

		claims, err := ai.parseToken(token)
		if err != nil {
			return connect.NewError(connect.CodeUnauthenticated, errors.New("invalid auth token"))
		}

		// Store claims in ctx
		ctx = auth.WithClaims(ctx, claims)

		return next(ctx, conn)
	}
}

func (ai *AuthInterceptor) isPublic(procedure string) (ok bool) {
	if ai == nil || ai.cfg == nil {
		return false
	}
	if len(ai.cfg.publicProcedures) == 0 {
		return false
	}

	_, ok = ai.cfg.publicProcedures[procedure]
	return ok
}

func (ai *AuthInterceptor) parseToken(token string) (claims *auth.OAuthClaims, err error) {
	var (
		jwks    *keyfunc.JWKS
		keyFunc jwt.Keyfunc
		raw     jwt.MapClaims
	)

	if ai.cfg == nil {
		return nil, errors.New("auth config not set")
	}

	if ai.cfg.useJWKS {
		if ai.cfg.jwks == nil {
			jwks, err = keyfunc.Get(ai.cfg.jwksURL, keyfunc.Options{RefreshInterval: time.Hour})
			if err != nil {
				return nil, err
			}
			ai.cfg.jwks = jwks
		}
		keyFunc = ai.cfg.jwks.Keyfunc
	} else {
		if ai.cfg.publicKey == nil {
			return nil, errors.New("no public key configured")
		}

		keyFunc = func(_ *jwt.Token) (any, error) {
			return ai.cfg.publicKey, nil
		}
	}

	// Parse and verify the JWT into the raw map representation so we can drive
	// path-based role extraction off the full claim set (including nested objects
	// like Keycloak's realm_access). Signature, exp, nbf, and iat are all checked
	// by the default validator.
	parsed, err := jwt.Parse(token, keyFunc)
	if err != nil {
		return nil, err
	}
	mapClaims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}
	raw = mapClaims

	// Re-hydrate the typed OAuthClaims view from the verified map. Errors here
	// are non-fatal: the structured fields are best-effort convenience accessors
	// and authorization decisions read from claims.Roles / claims.Raw.
	claims = &auth.OAuthClaims{Raw: raw}
	if b, mErr := json.Marshal(raw); mErr == nil {
		_ = json.Unmarshal(b, claims)
	}

	// Normalize roles from configured claim paths into claims.Roles.
	ai.applyRoleMapping(claims)

	return claims, nil
}

// applyRoleMapping extracts roles from configured claim paths, optionally
// translates them via the configured [RoleMapper], deduplicates, and stores
// the result in claims.Roles.
func (ai *AuthInterceptor) applyRoleMapping(claims *auth.OAuthClaims) {
	if ai == nil || ai.cfg == nil || claims == nil {
		return
	}
	if len(ai.cfg.roleClaimPaths) == 0 {
		return
	}

	var out []string
	seen := map[string]struct{}{}

	for _, path := range ai.cfg.roleClaimPaths {
		for _, r := range extractStringListAtPath(claims.Raw, path) {
			r = strings.TrimSpace(r)
			if r == "" {
				continue
			}
			if ai.cfg.roleMapper != nil {
				r = ai.cfg.roleMapper(r)
				if r == "" {
					continue
				}
			}
			if _, ok := seen[r]; ok {
				continue
			}
			seen[r] = struct{}{}
			out = append(out, r)
		}
	}

	if len(out) > 0 {
		claims.Roles = out
	}
}

// extractStringListAtPath reads a list of strings from a dotted path inside JWT MapClaims.
// Supported leaf formats:
// - []any / []string
// - string (space- or comma-separated)
func extractStringListAtPath(m jwt.MapClaims, dottedPath string) []string {
	if m == nil || dottedPath == "" {
		return nil
	}

	var cur any = map[string]any(m)
	for _, key := range strings.Split(dottedPath, ".") {
		obj, ok := cur.(map[string]any)
		if !ok {
			return nil
		}
		cur, ok = obj[key]
		if !ok {
			return nil
		}
	}

	switch v := cur.(type) {
	case []string:
		return v
	case []any:
		res := make([]string, 0, len(v))
		for _, it := range v {
			if s, ok := it.(string); ok {
				res = append(res, s)
			}
		}
		return res
	case string:
		// Accept "a b c" or "a,b,c"
		parts := strings.FieldsFunc(v, func(r rune) bool { return r == ' ' || r == ',' })
		return parts
	default:
		return nil
	}
}

// bearerToken extracts the token from the Authorization header. It expects the header to be in the
// format "Bearer <token>". If the header is missing, malformed, or the token is empty, it returns
// an error.
func bearerToken(header string) (token string, err error) {
	var parts []string

	if header == "" {
		return "", errors.New("missing authorization header")
	}

	parts = strings.Fields(header)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
		return "", errors.New("invalid authorization header")
	}

	if parts[1] == "" {
		return "", errors.New("empty token")
	}

	token = parts[1]
	return token, nil
}
