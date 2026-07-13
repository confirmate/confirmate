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
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"confirmate.io/core/api/orchestrator"
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
	DefaultOAuth2UIClientID      = "ui"
	DefaultOAuth2UIRedirectURI   = "http://localhost:5173/auth/callback"
	DefaultOAuth2ServiceClientID = "confirmate"
	DefaultOAuth2ServiceSecret   = "confirmate"
)

// DemoUser is a pre-configured user for the embedded OAuth2 demo server.
type DemoUser struct {
	Username  string
	Password  string
	FirstName string
	LastName  string
}

// DefaultDemoUsers are additional users registered in the embedded OAuth2 server for demo purposes.
var DefaultDemoUsers = []DemoUser{
	{Username: "alice", Password: "alice", FirstName: "Alice", LastName: "Adams"},
	{Username: "bob", Password: "bob", FirstName: "Bob", LastName: "Baker"},
	{Username: "charlie", Password: "charlie", FirstName: "Charlie", LastName: "Chen"},
}

// DemoOrchestratorUsers converts DefaultDemoUsers into orchestrator.User records for DB seeding.
// issuer is the public OAuth2 server URL (e.g. "http://localhost:8080/v1/auth"); when non-empty the
// user ID is constructed as "md5(issuer)-username" to match what GetConfirmateUserIDFromClaims produces.
func DemoOrchestratorUsers(issuer string) []*orchestrator.User {
	id := func(username string) string {
		if issuer != "" {
			h := md5.Sum([]byte(issuer))
			return hex.EncodeToString(h[:]) + "-" + username
		}
		return username
	}
	users := make([]*orchestrator.User, 0, len(DefaultDemoUsers)+1)
	adminName := DefaultOAuth2LoginUser
	users = append(users, &orchestrator.User{
		Id:        id(adminName),
		Username:  &adminName,
		FirstName: strPtr("Confirmate"),
		LastName:  strPtr("Admin"),
		Enabled:   true,
		Roles:     []orchestrator.Role{orchestrator.Role_ROLE_ADMIN},
	})
	for _, u := range DefaultDemoUsers {
		u := u
		users = append(users, &orchestrator.User{
			Id:        id(u.Username),
			Username:  &u.Username,
			FirstName: &u.FirstName,
			LastName:  &u.LastName,
			Enabled:   true,
			Roles:     []orchestrator.Role{orchestrator.Role_ROLE_TECHNICAL_IMPLEMENTER},
		})
	}
	return users
}

func strPtr(s string) *string { return &s }

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

		oauthPublicURL = NormalizeOAuthPublicURL(publicURL, srv.cfg.Port)
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
			oauth2.WithClient(DefaultOAuth2UIClientID, "", DefaultOAuth2UIRedirectURI),
			oauth2.WithClient(DefaultOAuth2ServiceClientID, DefaultOAuth2ServiceSecret, ""),
			login.WithLoginPage(
				login.WithBaseURL("/v1/auth"),
				login.WithUser(DefaultOAuth2LoginUser, DefaultOAuth2LoginPassword),
				login.WithUser(DefaultDemoUsers[0].Username, DefaultDemoUsers[0].Password),
				login.WithUser(DefaultDemoUsers[1].Username, DefaultDemoUsers[1].Password),
				login.WithUser(DefaultDemoUsers[2].Username, DefaultDemoUsers[2].Password),
			),
			oauth2.WithSigningKeysFunc(func() map[int]*ecdsa.PrivateKey {
				return storage.LoadSigningKeys(expandedKeyPath, keyPassword, saveOnCreate)
			}),
			oauth2.WithPublicURL(oauthPublicURL),
			oauth2.WithTokenClaimsFunc(func(clientID string, userID string) map[string]any {
				// Grant ROLE_ADMIN to service clients and the main admin user only.
				if clientID == DefaultOAuth2CLIClientID || clientID == DefaultOAuth2ServiceClientID || userID == DefaultOAuth2LoginUser {
					return map[string]any{
						"roles":       []string{"ROLE_ADMIN"},
						"given_name":  "Confirmate",
						"family_name": "Admin",
					}
				}
				// Demo users: include name claims so JIT-provisioning sets proper display names.
				for _, u := range DefaultDemoUsers {
					if userID == u.Username {
						return map[string]any{
							"roles":       []string{"ROLE_TECHNICAL_IMPLEMENTER"},
							"given_name":  u.FirstName,
							"family_name": u.LastName,
						}
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
		srv.httpHandlers["/v1/auth/logout"] = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.SetCookie(w, &http.Cookie{
				Name:    "id",
				Value:   "",
				Path:    "/v1/auth",
				MaxAge:  -1,
				Expires: time.Unix(0, 0),
			})
			returnTo := r.URL.Query().Get("return_to")
			if returnTo == "" {
				returnTo = "/"
			}
			http.Redirect(w, r, returnTo, http.StatusFound)
		})
	}
}

// NormalizeOAuthPublicURL ensures that the public URL for the OAuth 2.0 server is properly
// formatted, defaulting to http://localhost:<port>/v1/auth if no URL is provided, and appending
// /v1/auth if it's missing.
func NormalizeOAuthPublicURL(publicURL string, port uint16) (normalized string) {
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
