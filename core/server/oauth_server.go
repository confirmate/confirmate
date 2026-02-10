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
	"net/http"
	"os"
	"path/filepath"
	"strings"

	oauth2 "github.com/oxisto/oauth2go"
	"github.com/oxisto/oauth2go/storage"
)

const (
	DefaultOAuth2KeyPassword     = "changeme"
	DefaultOAuth2KeySaveOnCreate = true
	DefaultOAuth2KeyPath         = "~/.confirmate/api.key"
)

// WithEmbeddedOAuth2Server configures the server to include an embedded OAuth 2.0 authorization server.
// If publicURL is empty, it defaults to http://localhost:<port>/v1/auth.
func WithEmbeddedOAuth2Server(keyPath string, keyPassword string, saveOnCreate bool, publicURL string, opts ...oauth2.AuthorizationServerOption) Option {
	return func(svr *Server) {
		var oauthPublicURL string
		var authSrv *oauth2.AuthorizationServer

		oauthPublicURL = normalizeOAuthPublicURL(publicURL, svr.cfg.Port)

		opts = append(opts,
			oauth2.WithSigningKeysFunc(func() map[int]*ecdsa.PrivateKey {
				var path string
				path = expandPath(keyPath)
				return storage.LoadSigningKeys(path, keyPassword, saveOnCreate)
			}),
			oauth2.WithPublicURL(oauthPublicURL),
		)

		authSrv = oauth2.NewServer("", opts...)

		var authHandler func(w http.ResponseWriter, r *http.Request)
		authHandler = func(w http.ResponseWriter, r *http.Request) {
			http.StripPrefix("/v1/auth", authSrv.Handler).ServeHTTP(w, r)
		}

		svr.httpHandlers["/.well-known/openid-configuration"] = authSrv.Handler
		svr.httpHandlers["/v1/auth/certs"] = http.HandlerFunc(authHandler)
		svr.httpHandlers["/v1/auth/login"] = http.HandlerFunc(authHandler)
		svr.httpHandlers["/v1/auth/authorize"] = http.HandlerFunc(authHandler)
		svr.httpHandlers["/v1/auth/token"] = http.HandlerFunc(authHandler)
	}
}

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

func expandPath(path string) (expanded string) {
	var home string
	var clean string
	var err error

	if path == "" {
		return path
	}
	if strings.HasPrefix(path, "~") {
		home, err = os.UserHomeDir()
		if err != nil {
			return path
		}

		clean = strings.TrimPrefix(path, "~")
		clean = strings.TrimPrefix(clean, string(os.PathSeparator))
		expanded = filepath.Join(home, clean)
		return expanded
	}

	return path
}
