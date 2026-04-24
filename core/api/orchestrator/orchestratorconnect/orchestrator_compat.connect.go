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

package orchestratorconnect

import (
	"net/http"
	"strings"

	"connectrpc.com/connect"
)

const (
	// OrchestratorCompatName is a backwards-compatibility alias for clients that still
	// call the historic clouditor service name.
	OrchestratorCompatName = "clouditor.orchestrator.v1.Orchestrator"
)

// NewOrchestratorCompatHandler builds an HTTP handler exposing the Orchestrator API
// under the clouditor compatibility service name.
func NewOrchestratorCompatHandler(svc OrchestratorHandler, opts ...connect.HandlerOption) (string, http.Handler) {
	canonicalPath, canonicalHandler := NewOrchestratorHandler(svc, opts...)
	compatPathPrefix := "/" + OrchestratorCompatName + "/"

	return compatPathPrefix, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, compatPathPrefix) {
			http.NotFound(w, r)
			return
		}

		rewrittenRequest := r.Clone(r.Context())
		rewrittenURL := *r.URL
		rewrittenURL.Path = canonicalPath + strings.TrimPrefix(r.URL.Path, compatPathPrefix)
		rewrittenRequest.URL = &rewrittenURL

		canonicalHandler.ServeHTTP(w, rewrittenRequest)
	})
}
