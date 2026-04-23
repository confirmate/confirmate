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
	"crypto/elliptic"
	"crypto/rand"
	"net/http"
	"testing"

	"confirmate.io/core/auth"
	"confirmate.io/core/util/assert"

	"connectrpc.com/connect"
	"github.com/MicahParks/keyfunc/v2"
	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/protobuf/types/known/emptypb"
)

func wantError(t *testing.T, err error, msgAndArgs ...any) bool {
	t.Helper()

	return assert.Error(t, err, msgAndArgs...)
}

func TestBearerToken(t *testing.T) {
	type args struct {
		header string
	}

	tests := []struct {
		name    string
		args    args
		want    assert.Want[string]
		wantErr assert.WantErr
	}{
		{
			name: "valid bearer header",
			args: args{header: "Bearer abc.def.ghi"},
			want: func(t *testing.T, got string, _ ...any) bool {
				return assert.Equal(t, "abc.def.ghi", got)
			},
			wantErr: assert.NoError,
		},
		{
			name:    "missing header",
			args:    args{header: ""},
			want:    assert.AnyValue[string],
			wantErr: wantError,
		},
		{
			name:    "wrong scheme",
			args:    args{header: "Basic abc"},
			want:    assert.AnyValue[string],
			wantErr: wantError,
		},
		{
			name:    "malformed bearer header",
			args:    args{header: "Bearer"},
			want:    assert.AnyValue[string],
			wantErr: wantError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := bearerToken(tt.args.header)

			assert.True(t, tt.wantErr(t, err))
			assert.True(t, tt.want(t, got))
		})
	}
}

func TestAuthInterceptorWrapUnary(t *testing.T) {
	type args struct {
		authHeader string
	}
	type fields struct {
		interceptor *AuthInterceptor
	}
	type gotData struct {
		code       connect.Code
		nextCalled bool
		claims     *auth.OAuthClaims
	}

	var (
		privateKey, publicKey = mustECDSAKeyPair(t)
		validToken            = mustSignES256Token(t, privateKey, "kid-1", jwt.MapClaims{"sub": "user-1", "cfadmin": true})
		validTokenViaRoles    = mustSignES256Token(t, privateKey, "kid-1", jwt.MapClaims{"sub": "user-role-admin", "roles": []string{"ROLE_ADMIN"}})
		invalidToken          = mustSignES256Token(t, mustECDSAKeyPairPrivateOnly(t), "kid-1", jwt.MapClaims{"sub": "user-1"})
	)

	tests := []struct {
		name    string
		args    args
		fields  fields
		want    assert.Want[gotData]
		wantErr assert.WantErr
	}{
		{
			name:   "public procedure bypasses auth",
			args:   args{authHeader: ""},
			fields: fields{interceptor: NewAuthInterceptor(WithPublicProcedures(""), WithPublicKey(publicKey))},
			want: func(t *testing.T, got gotData, _ ...any) bool {
				return assert.True(t, got.nextCalled) &&
					assert.Nil(t, got.claims)
			},
			wantErr: assert.NoError,
		},
		{
			name:   "missing authorization header returns unauthenticated",
			args:   args{authHeader: ""},
			fields: fields{interceptor: NewAuthInterceptor(WithPublicKey(publicKey))},
			want: func(t *testing.T, got gotData, _ ...any) bool {
				return assert.Equal(t, connect.CodeUnauthenticated, got.code) &&
					assert.False(t, got.nextCalled)
			},
			wantErr: wantError,
		},
		{
			name:   "invalid signature returns unauthenticated",
			args:   args{authHeader: "Bearer " + invalidToken},
			fields: fields{interceptor: NewAuthInterceptor(WithPublicKey(publicKey))},
			want: func(t *testing.T, got gotData, _ ...any) bool {
				return assert.Equal(t, connect.CodeUnauthenticated, got.code) &&
					assert.False(t, got.nextCalled)
			},
			wantErr: wantError,
		},
		{
			name:   "valid token passes, sets claims",
			args:   args{authHeader: "Bearer " + validToken},
			fields: fields{interceptor: NewAuthInterceptor(WithPublicKey(publicKey))},
			want: func(t *testing.T, got gotData, _ ...any) bool {
				// Check if claims are set correctly
				assert.True(t, got.claims.IsAdminToken)
				assert.NotEmpty(t, got.claims.Subject)

				return assert.True(t, got.nextCalled)
			},
			wantErr: assert.NoError,
		},
		{
			name:   "valid token with ROLE_ADMIN sets admin claims",
			args:   args{authHeader: "Bearer " + validTokenViaRoles},
			fields: fields{interceptor: NewAuthInterceptor(WithPublicKey(publicKey))},
			want: func(t *testing.T, got gotData, _ ...any) bool {
				assert.True(t, got.claims.IsAdmin())
				assert.Equal(t, "user-role-admin", got.claims.Subject)
				return assert.True(t, got.nextCalled)
			},
			wantErr: assert.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				nextCalled bool
				claims     *auth.OAuthClaims
			)

			wrapped := tt.fields.interceptor.WrapUnary(func(ctx context.Context, _ connect.AnyRequest) (connect.AnyResponse, error) {
				nextCalled = true
				claims, _ = auth.ClaimsFromContext(ctx)
				return connect.NewResponse(&emptypb.Empty{}), nil
			})

			req := connect.NewRequest(&emptypb.Empty{})
			if tt.args.authHeader != "" {
				req.Header().Set("Authorization", tt.args.authHeader)
			}

			_, err := wrapped(context.Background(), req)
			got := gotData{code: connect.CodeOf(err), nextCalled: nextCalled, claims: claims}

			assert.True(t, tt.wantErr(t, err))
			assert.True(t, tt.want(t, got))
		})
	}
}

func TestAuthInterceptorParseToken(t *testing.T) {
	type args struct {
		token string
	}
	type fields struct {
		interceptor *AuthInterceptor
	}

	var (
		privateKey, publicKey = mustECDSAKeyPair(t)
		kid                   = "jwks-kid-1"
		given                 = keyfunc.NewGivenECDSA(publicKey, keyfunc.GivenKeyOptions{Algorithm: jwt.SigningMethodES256.Alg()})
		jwks                  = keyfunc.NewGiven(map[string]keyfunc.GivenKey{kid: given})
		validJWKSToken        = mustSignES256Token(t, privateKey, kid, jwt.MapClaims{"sub": "jwks-user"})
		invalidJWKSToken      = mustSignES256Token(t, mustECDSAKeyPairPrivateOnly(t), kid, jwt.MapClaims{"sub": "other"})
		validPublicKeyToken   = mustSignES256Token(t, privateKey, kid, jwt.MapClaims{"sub": "pk-user"})
		roleAdminToken        = mustSignES256Token(t, privateKey, kid, jwt.MapClaims{"sub": "pk-role-admin", "roles": []string{"ROLE_ADMIN"}})
	)

	tests := []struct {
		name    string
		args    args
		fields  fields
		want    assert.Want[*auth.OAuthClaims]
		wantErr assert.WantErr
	}{
		{
			name:   "jwks validates signature",
			args:   args{token: validJWKSToken},
			fields: fields{interceptor: &AuthInterceptor{cfg: &AuthConfig{useJWKS: true, jwks: jwks}}},
			want: func(t *testing.T, got *auth.OAuthClaims, _ ...any) bool {
				return assert.Equal(t, "jwks-user", got.Subject)
			},
			wantErr: assert.NoError,
		},
		{
			name:   "jwks rejects invalid signature",
			args:   args{token: invalidJWKSToken},
			fields: fields{interceptor: &AuthInterceptor{cfg: &AuthConfig{useJWKS: true, jwks: jwks}}},
			want:   assert.AnyValue[*auth.OAuthClaims],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return wantError(t, err, msgAndArgs...)
			},
		},
		{
			name:   "public key validates signature",
			args:   args{token: validPublicKeyToken},
			fields: fields{interceptor: NewAuthInterceptor(WithPublicKey(publicKey))},
			want: func(t *testing.T, got *auth.OAuthClaims, _ ...any) bool {
				return assert.Equal(t, "pk-user", got.Subject)
			},
			wantErr: assert.NoError,
		},
		{
			name:   "public key sets admin from ROLE_ADMIN role",
			args:   args{token: roleAdminToken},
			fields: fields{interceptor: NewAuthInterceptor(WithPublicKey(publicKey))},
			want: func(t *testing.T, got *auth.OAuthClaims, _ ...any) bool {
				return assert.Equal(t, "pk-role-admin", got.Subject) &&
					assert.True(t, got.IsAdmin())
			},
			wantErr: assert.NoError,
		},
		{
			name:   "missing public key returns error",
			args:   args{token: validPublicKeyToken},
			fields: fields{interceptor: NewAuthInterceptor()},
			want:   assert.AnyValue[*auth.OAuthClaims],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.ErrorContains(t, err, "no public key configured", msgAndArgs...)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.fields.interceptor.parseToken(tt.args.token)

			assert.True(t, tt.wantErr(t, err))
			assert.True(t, tt.want(t, got))
		})
	}
}

func TestAuthInterceptorWrapStreamingHandler(t *testing.T) {
	type args struct {
		procedure string
		header    string
	}
	type fields struct {
		interceptor *AuthInterceptor
	}
	type gotData struct {
		code       connect.Code
		nextCalled bool
		claimsSet  bool
	}

	var (
		privateKey, publicKey = mustECDSAKeyPair(t)
		validToken            = mustSignES256Token(t, privateKey, "kid-1", jwt.MapClaims{"sub": "stream-user"})
	)

	tests := []struct {
		name    string
		args    args
		fields  fields
		want    assert.Want[gotData]
		wantErr assert.WantErr
	}{
		{
			name:   "public procedure bypasses auth",
			args:   args{procedure: "/svc/Public", header: ""},
			fields: fields{interceptor: NewAuthInterceptor(WithPublicProcedures("/svc/Public"), WithPublicKey(publicKey))},
			want: func(t *testing.T, got gotData, _ ...any) bool {
				return assert.True(t, got.nextCalled) &&
					assert.False(t, got.claimsSet)
			},
			wantErr: assert.NoError,
		},
		{
			name:   "missing authorization returns unauthenticated",
			args:   args{procedure: "/svc/Method", header: ""},
			fields: fields{interceptor: NewAuthInterceptor(WithPublicKey(publicKey))},
			want: func(t *testing.T, got gotData, _ ...any) bool {
				return assert.Equal(t, connect.CodeUnauthenticated, got.code) &&
					assert.False(t, got.nextCalled)
			},
			wantErr: wantError,
		},
		{
			name:   "valid token allows streaming handler and sets claims",
			args:   args{procedure: "/svc/Method", header: "Bearer " + validToken},
			fields: fields{interceptor: NewAuthInterceptor(WithPublicKey(publicKey))},
			want: func(t *testing.T, got gotData, _ ...any) bool {
				return assert.True(t, got.nextCalled) &&
					assert.True(t, got.claimsSet)
			},
			wantErr: assert.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				nextCalled bool
				claimsSet  bool
			)

			wrapped := tt.fields.interceptor.WrapStreamingHandler(func(ctx context.Context, _ connect.StreamingHandlerConn) error {
				nextCalled = true
				_, claimsSet = auth.ClaimsFromContext(ctx)
				return nil
			})

			conn := &testStreamingConn{
				spec:            connect.Spec{Procedure: tt.args.procedure},
				requestHeader:   make(http.Header),
				responseHeader:  make(http.Header),
				responseTrailer: make(http.Header),
			}
			if tt.args.header != "" {
				conn.requestHeader.Set("Authorization", tt.args.header)
			}

			err := wrapped(context.Background(), conn)
			got := gotData{code: connect.CodeOf(err), nextCalled: nextCalled, claimsSet: claimsSet}

			assert.True(t, tt.wantErr(t, err))
			assert.True(t, tt.want(t, got), context.Background())
		})
	}
}

type testStreamingConn struct {
	spec            connect.Spec
	requestHeader   http.Header
	responseHeader  http.Header
	responseTrailer http.Header
}

func (t *testStreamingConn) Spec() connect.Spec           { return t.spec }
func (t *testStreamingConn) Peer() connect.Peer           { return connect.Peer{} }
func (t *testStreamingConn) Receive(any) error            { return nil }
func (t *testStreamingConn) RequestHeader() http.Header   { return t.requestHeader }
func (t *testStreamingConn) Send(any) error               { return nil }
func (t *testStreamingConn) ResponseHeader() http.Header  { return t.responseHeader }
func (t *testStreamingConn) ResponseTrailer() http.Header { return t.responseTrailer }

func mustECDSAKeyPair(t *testing.T) (privateKey *ecdsa.PrivateKey, publicKey *ecdsa.PublicKey) {
	t.Helper()

	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("failed to generate key pair: %v", err)
	}

	publicKey = &privateKey.PublicKey
	return privateKey, publicKey
}

func mustECDSAKeyPairPrivateOnly(t *testing.T) (privateKey *ecdsa.PrivateKey) {
	t.Helper()

	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("failed to generate private key: %v", err)
	}

	return privateKey
}

func mustSignES256Token(t *testing.T, privateKey *ecdsa.PrivateKey, kid string, claims jwt.MapClaims) (token string) {
	t.Helper()

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	if kid != "" {
		jwtToken.Header["kid"] = kid
	}

	token, err := jwtToken.SignedString(privateKey)
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}

	return token
}
