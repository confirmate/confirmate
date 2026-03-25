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
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"confirmate.io/core/api/orchestrator"
	"confirmate.io/core/auth"
	"confirmate.io/core/log"
	"confirmate.io/core/persistence"
	"confirmate.io/core/util"

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

	// PersistenceConfig is the configuration for the persistence layer.
	persistenceConfig persistence.Config

	publicKey *ecdsa.PublicKey

	publicProcedures map[string]struct{}
}

// AuthOption configures the auth middleware.
type AuthOption func(*AuthConfig)

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

// WithUserPersistence enables persisting the calling user (by JWT sub) on each request.
func WithUserPersistence(config persistence.Config) AuthOption {
	return func(c *AuthConfig) {
		c.persistenceConfig = config
	}
}

// AuthInterceptor authenticates incoming requests using bearer tokens.
type AuthInterceptor struct {
	cfg *AuthConfig
	db  persistence.DB
}

// NewAuthInterceptor creates a new auth interceptor.
func NewAuthInterceptor(opts ...AuthOption) (interceptor *AuthInterceptor) {
	var (
		cfg *AuthConfig
		err error
		db  persistence.DB
	)

	cfg = &AuthConfig{}
	for _, opt := range opts {
		opt(cfg)
	}

	// Initialize the database with the defined auto-migration types and join tables
	if cfg.persistenceConfig.Host == "" {
		slog.Debug("no persistence config provided for auth interceptor, use in-memory-db for user persistence")
		cfg.persistenceConfig = persistence.Config{
			InMemoryDB: true,
		}
	}
	pcfg := cfg.persistenceConfig
	pcfg.Types = []any{&orchestrator.User{}}
	pcfg.CustomJoinTables = []persistence.CustomJoinTable{}
	db, err = persistence.NewDB(persistence.WithConfig(pcfg))
	if err != nil {
		slog.Warn("could not create db", log.Err(err))
		db = nil
	}

	interceptor = &AuthInterceptor{
		cfg: cfg,
		db:  db,
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

		// Extract subject from claims and persist user if database is configured
		ctx, err = ai.persistUserFromClaims(ctx, claims)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, errors.New("failed to persist user"))
		}
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

		// Extract subject from claims and persist user if database is configured
		ctx, err = ai.persistUserFromClaims(ctx, claims)
		if err != nil {
			return connect.NewError(connect.CodeInternal, errors.New("failed to persist user"))
		}

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

func (ai *AuthInterceptor) parseToken(token string) (claims jwt.MapClaims, err error) {
	var (
		jwks     *keyfunc.JWKS
		keyFunc  jwt.Keyfunc
		parsed   *jwt.Token
		claimsOK bool
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

	parsed, err = jwt.ParseWithClaims(token, jwt.MapClaims{}, keyFunc)
	if err != nil {
		return nil, err
	}

	claims, claimsOK = parsed.Claims.(jwt.MapClaims)
	if !claimsOK {
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
}

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

// upsertUserFromClaims creates or updates a user in the database based on JWT claims.
func upsertUserFromClaims(ctx context.Context, db persistence.DB, sub string, claims jwt.MapClaims) error {
	attrs := map[string]string{
		"sub": sub,
	}

	// TODO(anatheka): Are they available and necessary?
	// extraction of common claims
	if v, ok := claims["email"].(string); ok && v != "" {
		attrs["email"] = v
	}
	if v, ok := claims["preferred_username"].(string); ok && v != "" {
		attrs["preferred_username"] = v
	}
	if v, ok := claims["given_name"].(string); ok && v != "" {
		attrs["given_name"] = v
	}
	if v, ok := claims["family_name"].(string); ok && v != "" {
		attrs["family_name"] = v
	}
	if v, ok := claims["groups"].(string); ok && v != "" {
		attrs["groups"] = v
	}
	if v, ok := claims["roles"].(string); ok && v != "" {
		attrs["roles"] = v
	}

	user := &orchestrator.User{
		Id:        sub,
		Username:  attrs["preferred_username"],
		Email:     util.Ref(attrs["email"]),
		FirstName: util.Ref(attrs["given_name"]),
		LastName:  util.Ref(attrs["family_name"]),
		// Roles:      ,
		Enabled:    true,
		Attributes: attrs,
	}

	// TODO(antheka): Should we better use create and update???
	if err := db.Save(user); err != nil {
		return fmt.Errorf("could not upsert user %q: %w", sub, err)
	}
	return nil
}

// WithSubject stores the JWT subject in the context for later use (authz, auditing, etc.).
func WithSubject(ctx context.Context, sub string) context.Context {
	return context.WithValue(ctx, "sub", sub)
}

// SubjectFromContext returns the stored JWT subject.
func SubjectFromContext(ctx context.Context) (string, error) {
	sub, ok := ctx.Value("sub").(string)
	if !ok {
		return "", errors.New("subject not found in context")
	}

	return sub, nil
}

func (ai *AuthInterceptor) persistUserFromClaims(ctx context.Context, claims jwt.MapClaims) (context context.Context, err error) {
	sub, ok := claims["sub"].(string)
	if ok && ai != nil && ai.db != nil {
		sub = strings.TrimSpace(sub)
		// Persist the calling user (identified by JWT sub) in the database
		err := upsertUserFromClaims(ctx, ai.db, sub, claims)
		if err != nil {
			return ctx, connect.NewError(connect.CodeInternal, errors.New("failed to persist user"))
		}
		ctx = WithSubject(ctx, sub)
	}

	return ctx, nil
}
