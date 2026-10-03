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
	"net/url"
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
// If publicURL is empty, it defaults to http://localhost:<port>/v1/auth. The path of publicURL is
// also used as the browser-facing base path of the login page, so the login flow keeps working
// when the API is served behind a path-stripping reverse proxy (e.g. /proxy/5173/v1/auth). In that
// case, redirects are made absolute against publicURL, because some proxies (e.g. code-server)
// prefix root-relative Location headers with their own path a second time.
// If uiRedirectURI is empty, the UI client is registered with [DefaultOAuth2UIRedirectURI].
func WithEmbeddedOAuth2Server(keyPath string, keyPassword string, saveOnCreate bool, publicURL string, uiRedirectURI string, opts ...oauth2.AuthorizationServerOption) Option {
	return func(srv *Server) {
		var (
			oauthPublicURL  string
			loginBaseURL    string
			proxyPrefix     string
			redirectOrigin  string
			expandedKeyPath string
			authSrv         *oauth2.AuthorizationServer
			authHandler     func(w http.ResponseWriter, r *http.Request)
		)

		oauthPublicURL = NormalizeOAuthPublicURL(publicURL, srv.cfg.Port)
		loginBaseURL = OAuthPublicPath(oauthPublicURL)
		proxyPrefix = strings.TrimSuffix(loginBaseURL, "/v1/auth")
		if proxyPrefix != "" {
			redirectOrigin = strings.TrimSuffix(oauthPublicURL, loginBaseURL)
		}
		expandedKeyPath = util.ExpandPath(keyPath)
		if uiRedirectURI == "" {
			uiRedirectURI = DefaultOAuth2UIRedirectURI
		}

		slog.Info("Configuring embedded OAuth 2.0 server",
			slog.String("public_url", oauthPublicURL),
			slog.String("login_base_url", loginBaseURL),
			slog.String("key_path", expandedKeyPath),
			slog.Bool("key_save_on_create", saveOnCreate),
			slog.String("login_user", DefaultOAuth2LoginUser),
			slog.String("cli_client_id", DefaultOAuth2CLIClientID),
			slog.String("cli_redirect_uri", DefaultOAuth2CLIRedirectURI),
			slog.String("ui_redirect_uri", uiRedirectURI),
			slog.String("service_client_id", DefaultOAuth2ServiceClientID),
		)

		// Build the login page's user options dynamically: DefaultDemoUsers may be overridden via
		// --demo-seed-file with fewer or more than the built-in 3 demo users. The element type is
		// unexported by the login package, so it cannot be named in a var (...) block above.
		var loginPageOpts = sliceOf(
			login.WithBaseURL(loginBaseURL),
			login.WithUser(DefaultOAuth2LoginUser, DefaultOAuth2LoginPassword),
		)
		for _, u := range DefaultDemoUsers {
			loginPageOpts = append(loginPageOpts, login.WithUser(u.Username, u.Password))
		}

		opts = append(opts,
			oauth2.WithClient(DefaultOAuth2CLIClientID, "", DefaultOAuth2CLIRedirectURI),
			oauth2.WithClient(DefaultOAuth2UIClientID, "", uiRedirectURI),
			oauth2.WithClient(DefaultOAuth2ServiceClientID, DefaultOAuth2ServiceSecret, ""),
			login.WithLoginPage(loginPageOpts...),
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
			// A path-stripping reverse proxy hands us /v1/auth/..., but the browser sits at
			// <prefix>/v1/auth/.... The login page sends the user back to the request URI of the
			// authorize call after a successful login, so it must carry the browser-facing prefix.
			if proxyPrefix != "" {
				r = r.Clone(r.Context())
				r.RequestURI = proxyPrefix + r.RequestURI
				w = &absoluteRedirectWriter{ResponseWriter: w, origin: redirectOrigin}
			}
			http.StripPrefix("/v1/auth", authSrv.Handler).ServeHTTP(w, r)
		}

		srv.httpHandlers["/.well-known/openid-configuration"] = authSrv.Handler
		srv.httpHandlers["/v1/auth/certs"] = http.HandlerFunc(authHandler)
		srv.httpHandlers["/v1/auth/login"] = http.HandlerFunc(authHandler)
		srv.httpHandlers["/v1/auth/authorize"] = http.HandlerFunc(authHandler)
		srv.httpHandlers["/v1/auth/token"] = http.HandlerFunc(authHandler)
		srv.httpHandlers["/v1/auth/logout"] = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if proxyPrefix != "" {
				w = &absoluteRedirectWriter{ResponseWriter: w, origin: redirectOrigin}
			}
			http.SetCookie(w, &http.Cookie{
				Name:    "id",
				Value:   "",
				Path:    loginBaseURL,
				MaxAge:  -1,
				Expires: time.Unix(0, 0),
			})
			returnTo := r.URL.Query().Get("return_to")
			if !isSafeRedirectPath(returnTo) {
				returnTo = "/"
			}
			http.Redirect(w, r, returnTo, http.StatusFound)
		})
	}
}

// absoluteRedirectWriter turns root-relative Location headers into absolute URLs below origin
// before the response header is written.
type absoluteRedirectWriter struct {
	http.ResponseWriter
	origin string
}

func (w *absoluteRedirectWriter) WriteHeader(statusCode int) {
	var location = w.Header().Get("Location")

	if strings.HasPrefix(location, "/") && !strings.HasPrefix(location, "//") {
		w.Header().Set("Location", w.origin+location)
	}

	w.ResponseWriter.WriteHeader(statusCode)
}

// sliceOf collects its variadic arguments into a slice. This lets us build a dynamically-sized
// slice of the login package's option type, which is unexported and so cannot be spelled here.
func sliceOf[T any](items ...T) []T {
	return items
}

// isSafeRedirectPath reports whether path is safe to redirect to after logout: a same-origin,
// relative path. This rejects absolute URLs and protocol-relative URLs (e.g. "//evil.com", which
// browsers treat as a same-scheme redirect to a different host), which would otherwise allow an
// open redirect via the return_to query parameter.
func isSafeRedirectPath(path string) (safe bool) {
	var (
		u   *url.URL
		err error
	)

	if path == "" || path[0] != '/' || strings.HasPrefix(path, "//") || strings.HasPrefix(path, "/\\") {
		return false
	}

	u, err = url.Parse(path)
	if err != nil {
		return false
	}

	safe = u.Scheme == "" && u.Host == ""
	return safe
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

// OAuthPublicPath returns the path component of a normalized OAuth 2.0 public URL, which is the
// browser-facing base path of the embedded login page. It falls back to /v1/auth if the URL cannot
// be parsed or has no path.
func OAuthPublicPath(publicURL string) (path string) {
	var (
		u   *url.URL
		err error
	)

	u, err = url.Parse(publicURL)
	if err != nil || u.Path == "" {
		return "/v1/auth"
	}

	path = u.Path
	return path
}
