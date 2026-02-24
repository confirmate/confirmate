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

	"connectrpc.com/vanguard"

	"confirmate.io/core/log"
)

// Server represents a Connect server, with RPC and HTTP support.
type Server struct {
	*http.Server
	cfg Config
}

// Option is a functional option for configuring the [Server].
type Option func(*Server)

// WithConfig sets the server configuration, overriding the default configuration.
func WithConfig(cfg Config) Option {
	return func(svr *Server) {
		if cfg.Port != 0 {
			svr.cfg.Port = cfg.Port
		}
		if cfg.Path != "" {
			svr.cfg.Path = cfg.Path
		}
		if cfg.LogLevel != "" {
			svr.cfg.LogLevel = cfg.LogLevel
		}
		if len(cfg.CORS.AllowedOrigins) > 0 {
			svr.cfg.CORS.AllowedOrigins = cfg.CORS.AllowedOrigins
		}
		if len(cfg.CORS.AllowedMethods) > 0 {
			svr.cfg.CORS.AllowedMethods = cfg.CORS.AllowedMethods
		}
		if len(cfg.CORS.AllowedHeaders) > 0 {
			svr.cfg.CORS.AllowedHeaders = cfg.CORS.AllowedHeaders
		}
		if cfg.Handlers != nil {
			svr.cfg.Handlers = cfg.Handlers
		}
	}
}

// RunConnectServer runs a Connect server with the given options.
// It uses [http.Protocols] to serve HTTP/2 without TLS (h2c).
func RunConnectServer(opts ...Option) (err error) {
	var (
		svr *Server
	)

	svr, err = NewConnectServer(opts)
	if err != nil {
		return
	}

	err = svr.ListenAndServe()

	return err
}

// NewConnectServer creates a new Connect server with the given options.
// It uses [http.Protocols] to serve HTTP/2 without TLS (h2c).
func NewConnectServer(opts []Option) (srv *Server, err error) {
	var (
		svr        *Server
		vs         []*vanguard.Service
		transcoder http.Handler
		mux        *http.ServeMux
		p          *http.Protocols
	)

	// Setup default server config
	svr = &Server{
		cfg: DefaultConfig,
	}

	// Apply options
	for _, opt := range opts {
		opt(svr)
	}

	// Configure log level
	if err = configureLogLevel(svr.cfg.LogLevel); err != nil {
		return nil, fmt.Errorf("invalid log level %q: %w", svr.cfg.LogLevel, err)
	}

	// Create one vanguard service for each handler and add to transcoder
	for path, handler := range svr.cfg.Handlers {
		vs = append(vs, vanguard.NewService(path, handler))
	}
	transcoder, err = vanguard.NewTranscoder(vs)
	if err != nil {
		slog.Error("Failed to create vanguard transcoder", log.Err(err))
		return nil, err
	}

	// Create new mux
	mux = http.NewServeMux()
	mux.Handle("/", srv.handleCORS(transcoder))

	// Configure h2c support using standard library
	p = new(http.Protocols)
	p.SetHTTP1(true)
	p.SetUnencryptedHTTP2(true)

	// Set address, handler, and protocols
	svr.Server = &http.Server{
		Addr:      fmt.Sprintf("localhost:%d", svr.cfg.Port),
		Handler:   mux,
		Protocols: p,
	}

	slog.Info("Starting Connect server",
		slog.String("address", svr.Addr),
		slog.String("path", svr.cfg.Path),
	)

	return svr, nil
}

// configureLogLevel configures the global slog logger with the specified level.
func configureLogLevel(levelStr string) error {
	return log.Configure(levelStr)
}
