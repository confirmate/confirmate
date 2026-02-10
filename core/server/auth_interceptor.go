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

// AuthInterceptor authenticates incoming requests using bearer tokens.
type AuthInterceptor struct {
	cfg *AuthConfig
}

// NewAuthInterceptor creates a new auth interceptor.
func NewAuthInterceptor(opts ...AuthOption) (interceptor *AuthInterceptor) {
	var cfg *AuthConfig

	cfg = &AuthConfig{}
	for _, opt := range opts {
		opt(cfg)
	}

	interceptor = &AuthInterceptor{cfg: cfg}
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

		if err = ai.parseToken(token); err != nil {
			return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("invalid auth token"))
		}

		ctx = auth.WithToken(ctx, token)
		res, err = next(ctx, req)
		return res, err
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

		if err = ai.parseToken(token); err != nil {
			return connect.NewError(connect.CodeUnauthenticated, errors.New("invalid auth token"))
		}

		ctx = auth.WithToken(ctx, token)
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

func (ai *AuthInterceptor) parseToken(token string) (err error) {
	var jwks *keyfunc.JWKS
	var parseErr error

	if ai.cfg == nil {
		return errors.New("auth config not set")
	}

	if ai.cfg.useJWKS {
		if ai.cfg.jwks == nil {
			jwks, err = keyfunc.Get(ai.cfg.jwksURL, keyfunc.Options{RefreshInterval: time.Hour})
			if err != nil {
				return err
			}
			ai.cfg.jwks = jwks
		}
		_, parseErr = jwt.ParseWithClaims(token, jwt.MapClaims{}, ai.cfg.jwks.Keyfunc)
		return parseErr
	}

	if ai.cfg.publicKey == nil {
		return errors.New("no public key configured")
	}

	_, parseErr = jwt.ParseWithClaims(token, jwt.MapClaims{}, func(_ *jwt.Token) (any, error) {
		return ai.cfg.publicKey, nil
	})
	return parseErr
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
