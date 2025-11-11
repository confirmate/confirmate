// Copyright 2016-2025 Fraunhofer AISEC
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
	"net/http"
	"strings"
)

// OriginAllowed checks if the supplied origin is allowed according to our global CORS
// configuration.
func (srv *Server) OriginAllowed(origin string) bool {
	// If no origin is specified, we are running in a non-browser environment and this means, that
	// all origins are allowed
	if origin == "" {
		return true
	}

	for _, v := range srv.cfg.CORS.AllowedOrigins {
		if origin == v {
			return true
		}
	}

	return false
}

// handleCORS wraps an existing [http.Handler] into an appropriate [http.HandlerFunc] to configure
// Cross-Origin Resource Sharing (CORS) according to our global configuration.
func (srv *Server) handleCORS(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check, if we allow this specific origin
		origin := r.Header.Get("Origin")
		if srv.OriginAllowed(origin) {
			// Set the appropriate access control header
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Add("Vary", "Origin")

			// Additionally, we need to handle preflight (OPTIONS) requests to specify allowed
			// headers and methods
			if r.Method == "OPTIONS" && r.Header.Get("Access-Control-Request-Method") != "" {
				w.Header().Set("Access-Control-Allow-Headers", strings.Join(srv.cfg.CORS.AllowedHeaders, ","))
				w.Header().Set("Access-Control-Allow-Methods", strings.Join(srv.cfg.CORS.AllowedMethods, ","))
				return
			}
		}

		h.ServeHTTP(w, r)
	})
}
