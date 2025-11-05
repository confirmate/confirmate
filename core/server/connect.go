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
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"connectrpc.com/vanguard"
	"github.com/lmittmann/tint"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

var (
	logger *slog.Logger
)

func init() {
	logger = slog.New(tint.NewHandler(os.Stdout, nil))

	slog.SetDefault(logger)
}

// corsMiddleware adds CORS headers to all requests
func corsMiddleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Connect-Protocol-Version, Connect-Timeout-Ms")
		w.Header().Set("Access-Control-Expose-Headers", "Connect-Protocol-Version")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		handler.ServeHTTP(w, r)
	})
}

// RunConnectServer runs a Connect server with the given [net/http.Handler] at the given path.
// It uses [golang.org/x/net/http2/h2c] to serve HTTP/2 without TLS.
func RunConnectServer(path string, handler http.Handler) (err error) {
	var (
		svc  *vanguard.Service
		mux  *http.ServeMux
		port = "8080"
		addr = fmt.Sprintf("localhost:%s", port)
	)

	svc = vanguard.NewService(path, handler)
	transcoder, _ := vanguard.NewTranscoder([]*vanguard.Service{
		svc,
	})

	slog.Info("Starting Connect server",
		slog.String("address", addr),
		slog.String("path", path),
	)

	mux = http.NewServeMux()
	mux.Handle("/", corsMiddleware(transcoder))
	err = http.ListenAndServe(
		addr,
		// Use h2c so we can serve HTTP/2 without TLS.
		h2c.NewHandler(mux, &http2.Server{}),
	)

	return err
}
