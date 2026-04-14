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
	"slices"
	"strings"

	"connectrpc.com/grpcreflect"
	"connectrpc.com/vanguard"

	"confirmate.io/core/log"
)

// Server represents a Connect server, with RPC and HTTP support.
type Server struct {
	*http.Server
	cfg          Config
	handlers     map[string]http.Handler
	httpHandlers map[string]http.Handler
}

// Option is a functional option for configuring the [Server].
type Option func(*Server)

// WithConfig sets the server configuration, overriding the default configuration.
func WithConfig(cfg Config) Option {
	return func(srv *Server) {
		srv.cfg = cfg
	}
}

// WithHandler adds an [http.Handler] at the specified path to the server.
// Multiple handlers can be registered by calling WithHandler multiple times.
func WithHandler(path string, handler http.Handler) Option {
	return func(srv *Server) {
		srv.handlers[path] = handler
	}
}

// WithReflection adds gRPC reflection support to the server, which allows clients to query the
// server for its supported services and methods.
func WithReflection() Option {
	return func(srv *Server) {
		srv.cfg.UseGRPCReflection = true
	}
}

func registerReflectionHandlers(srv *Server) {
	var (
		reflector         *grpcreflect.Reflector
		reflectionV1Path  string
		reflectionV1      http.Handler
		reflectionV1A     http.Handler
		reflectionV1APath string
	)

	reflector = grpcreflect.NewReflector(srv)
	reflectionV1Path, reflectionV1 = grpcreflect.NewHandlerV1(reflector)
	reflectionV1APath, reflectionV1A = grpcreflect.NewHandlerV1Alpha(reflector)

	srv.httpHandlers[reflectionV1Path] = reflectionV1
	srv.httpHandlers[reflectionV1APath] = reflectionV1A
}

// RunConnectServer runs a Connect server with the given options.
// It uses [http.Protocols] to serve HTTP/2 without TLS (h2c).
func RunConnectServer(opts ...Option) (err error) {
	var (
		srv *Server
	)

	srv, err = NewConnectServer(opts)
	if err != nil {
		return
	}

	err = srv.ListenAndServe()

	return err
}

// NewConnectServer creates a new Connect server with the given options.
// It uses [http.Protocols] to serve HTTP/2 without TLS (h2c).
func NewConnectServer(opts []Option) (srv *Server, err error) {
	var (
		vs         []*vanguard.Service
		transcoder http.Handler
		mux        *http.ServeMux
		p          *http.Protocols
	)

	// Setup default server config
	srv = &Server{
		cfg:          DefaultConfig,
		handlers:     make(map[string]http.Handler),
		httpHandlers: make(map[string]http.Handler),
	}

	// Apply options
	for _, opt := range opts {
		opt(srv)
	}

	// Configure log level
	if err = configureLogLevel(srv.cfg.LogLevel); err != nil {
		return nil, fmt.Errorf("invalid log level %q: %w", srv.cfg.LogLevel, err)
	}

	if srv.cfg.UseGRPCReflection {
		registerReflectionHandlers(srv)
	}

	// Create one vanguard service for each handler and add to transcoder
	for path, handler := range srv.handlers {
		vs = append(vs, vanguard.NewService(path, handler))
	}
	transcoder, err = vanguard.NewTranscoder(vs)
	if err != nil {
		slog.Error("Failed to create vanguard transcoder", log.Err(err))
		return nil, err
	}

	// Create new mux
	mux = http.NewServeMux()
	for path, handler := range srv.httpHandlers {
		mux.Handle(path, handler)
	}
	mux.Handle("/", srv.handleCORS(transcoder))

	// Configure h2c support using standard library
	p = new(http.Protocols)
	p.SetHTTP1(true)
	p.SetUnencryptedHTTP2(true)

	// Set address, handler, and protocols
	srv.Server = &http.Server{
		Addr:      fmt.Sprintf("0.0.0.0:%d", srv.cfg.Port),
		Handler:   mux,
		Protocols: p,
	}

	slog.Info("Starting Connect server",
		slog.String("address", srv.Addr),
		slog.String("path", srv.cfg.Path),
	)

	return srv, nil
}

// Names implements the [grpcreflect.Namer] interface, returning the names of the services supported
// by the server.
func (srv *Server) Names() []string {
	return serviceNamesFromHandlerPaths(srv.handlers)
}

// serviceNamesFromHandlerPaths extracts service names from the given map of handler paths to
// handlers.
func serviceNamesFromHandlerPaths(handlers map[string]http.Handler) (services []string) {
	var (
		path    string
		trimmed string
	)

	// Extract service names from handler paths by trimming leading and trailing slashes. For
	// example, a handler registered at path "/my.service.Name/" would yield the service name
	// "my.service.Name".
	for path = range handlers {
		trimmed = strings.Trim(path, "/")
		if trimmed == "" {
			continue
		}

		services = append(services, trimmed)
	}

	// Sort service names for consistent ordering
	slices.Sort(services)

	return services
}

// configureLogLevel configures the global slog logger with the specified level.
func configureLogLevel(levelStr string) error {
	return log.Configure(levelStr)
}
