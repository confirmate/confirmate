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
	"crypto/ecdsa"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"confirmate.io/core/util"
	oauth2 "github.com/oxisto/oauth2go"
	"github.com/oxisto/oauth2go/login"
	"github.com/oxisto/oauth2go/storage"
)

const (
	DefaultOAuth2KeyPassword     = "changeme"
	DefaultOAuth2KeySaveOnCreate = true
	DefaultOAuth2KeyPath         = "~/.confirmate/api.key"
	DefaultOAuth2LoginUser       = "confirmate"
	DefaultOAuth2LoginPassword   = "confirmate"
	DefaultOAuth2CLIClientID     = "cli"
	DefaultOAuth2CLIRedirectURI  = "http://localhost:10000/callback"
	DefaultOAuth2ServiceClientID = "confirmate"
	DefaultOAuth2ServiceSecret   = "confirmate"
)

// WithEmbeddedOAuth2Server configures the server to include an embedded OAuth 2.0 authorization server.
// If publicURL is empty, it defaults to http://localhost:<port>/v1/auth.
func WithEmbeddedOAuth2Server(keyPath string, keyPassword string, saveOnCreate bool, publicURL string, opts ...oauth2.AuthorizationServerOption) Option {
	return func(srv *Server) {
		var (
			oauthPublicURL  string
			expandedKeyPath string
			authSrv         *oauth2.AuthorizationServer
			authHandler     func(w http.ResponseWriter, r *http.Request)
		)

		oauthPublicURL = normalizeOAuthPublicURL(publicURL, srv.cfg.Port)
		expandedKeyPath = util.ExpandPath(keyPath)

		slog.Info("Configuring embedded OAuth 2.0 server",
			slog.String("public_url", oauthPublicURL),
			slog.String("key_path", expandedKeyPath),
			slog.Bool("key_save_on_create", saveOnCreate),
			slog.String("login_user", DefaultOAuth2LoginUser),
			slog.String("cli_client_id", DefaultOAuth2CLIClientID),
			slog.String("cli_redirect_uri", DefaultOAuth2CLIRedirectURI),
			slog.String("service_client_id", DefaultOAuth2ServiceClientID),
		)

		opts = append(opts,
			oauth2.WithClient(DefaultOAuth2CLIClientID, "", DefaultOAuth2CLIRedirectURI),
			oauth2.WithClient(DefaultOAuth2ServiceClientID, DefaultOAuth2ServiceSecret, ""),
			login.WithLoginPage(
				login.WithBaseURL("/v1/auth"),
				login.WithUser(DefaultOAuth2LoginUser, DefaultOAuth2LoginPassword),
			),
			oauth2.WithSigningKeysFunc(func() map[int]*ecdsa.PrivateKey {
				return storage.LoadSigningKeys(expandedKeyPath, keyPassword, saveOnCreate)
			}),
			oauth2.WithPublicURL(oauthPublicURL),
			oauth2.WithTokenClaimsFunc(func(clientID string, userID string) map[string]any {
				if clientID == DefaultOAuth2CLIClientID || userID == DefaultOAuth2LoginUser {
					// For now we assign the "admin" role to any user authenticated via the CLI
					// client or the default login user. In a real implementation, you would want to
					// have a more robust way of assigning roles and permissions.
					return map[string]any{
						"roles": []string{"ROLE_ADMIN"},
					}
				}

				return map[string]any{}
			}),
		)

		authSrv = oauth2.NewServer("", opts...)

		authHandler = func(w http.ResponseWriter, r *http.Request) {
			http.StripPrefix("/v1/auth", authSrv.Handler).ServeHTTP(w, r)
		}

		srv.httpHandlers["/.well-known/openid-configuration"] = authSrv.Handler
		srv.httpHandlers["/v1/auth/certs"] = http.HandlerFunc(authHandler)
		srv.httpHandlers["/v1/auth/login"] = http.HandlerFunc(authHandler)
		srv.httpHandlers["/v1/auth/authorize"] = http.HandlerFunc(authHandler)
		srv.httpHandlers["/v1/auth/token"] = http.HandlerFunc(authHandler)
	}
}

// normalizeOAuthPublicURL ensures that the public URL for the OAuth 2.0 server is properly
// formatted, defaulting to http://localhost:<port>/v1/auth if no URL is provided, and appending
// /v1/auth if it's missing.
func normalizeOAuthPublicURL(publicURL string, port uint16) (normalized string) {
	if publicURL == "" {
		normalized = fmt.Sprintf("http://localhost:%d/v1/auth", port)
		return normalized
	}

	publicURL = strings.TrimSuffix(publicURL, "/")
	if !strings.HasSuffix(publicURL, "/v1/auth") {
		normalized = publicURL + "/v1/auth"
		return normalized
	}

	return publicURL
}
