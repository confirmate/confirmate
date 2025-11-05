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

// Server represents a Connect server, with RPC and HTTP support.
type Server struct {
	*http.Server
	cfg      Config
	handlers map[string]http.Handler
}

// Option is a functional option for configuring the [Server].
type Option func(*Server)

// WithConfig sets the server configuration.
func WithConfig(cfg Config) Option {
	return func(svr *Server) {
		svr.cfg = cfg
	}
}

func WithHandler(path string, handler http.Handler) Option {
	return func(svr *Server) {
		svr.handlers[path] = handler
	}
}

// RunConnectServer runs a Connect server with the given options.
// It uses [http.Protocols] to serve HTTP/2 without TLS (h2c).
func RunConnectServer(opts ...Option) (err error) {
	var (
		svr *Server
		svc *vanguard.Service
		mux *http.ServeMux
		p   *http.Protocols
	)

	// Setup default server config
	svr = &Server{
		cfg:      DefaultConfig,
		handlers: make(map[string]http.Handler),
	}

	// Apply options
	for _, opt := range opts {
		opt(svr)
	}

	// Create one vanguard service for each handler and add to transcoder
	for path, handler := range svr.handlers {
		svc = vanguard.NewService(path, handler)
	}
	transcoder, _ := vanguard.NewTranscoder([]*vanguard.Service{
		svc,
	})

	// Create new mux
	mux = http.NewServeMux()
	mux.Handle("/", corsMiddleware(transcoder))

	// Configure h2c support using standard library
	p = new(http.Protocols)
	p.SetHTTP1(true)
	p.SetUnencryptedHTTP2(true)

	// Set address, handler, and protocols
	svr.Addr = fmt.Sprintf("localhost:%d", svr.cfg.Port)
	svr.Handler = mux
	svr.Protocols = p

	slog.Info("Starting Connect server",
		slog.String("address", svr.Addr),
		slog.String("path", svr.cfg.Path),
	)

	err = svr.ListenAndServe()

	return err
}
