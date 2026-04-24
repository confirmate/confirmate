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

package servertest

import (
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"confirmate.io/core/server"
	"confirmate.io/core/util/assert"
)

// NewTestConnectServer creates a new in-memory Connect server for testing purposes. It returns the
// server instance and an [httptest.Server] that can be used to send requests to the server. The
// server is already started in the background with HTTP/2 (h2c) support enabled for streaming.
//
// The caller must close the returned [httptest.Server] using testsrv.Close() when done. This will
// fail the test if the server could not be created.
func NewTestConnectServer(t *testing.T, opts ...server.Option) (srv *server.Server, testsrv *httptest.Server) {
	return NewTestConnectServerWithListener(t, nil, opts...)
}

// NewTestConnectServerWithListener creates a new in-memory Connect server for testing purposes,
// using the specified listener. This allows restarting the server on the same address. It returns
// the server instance and an [httptest.Server] that can be used to send requests to the server.
// The server is already started in the background.
//
// The caller must close the returned [httptest.Server] using testsrv.Close() when done. This will
// fail the test if the server could not be created.
func NewTestConnectServerWithListener(t *testing.T, listener net.Listener, opts ...server.Option) (srv *server.Server, testsrv *httptest.Server) {
	var err error

	srv, err = server.NewConnectServer(opts)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	if listener != nil {
		// Use provided listener
		testsrv = &httptest.Server{
			Listener: listener,
			Config:   &http.Server{Handler: srv.Handler},
		}
	} else {
		// Create new server with random port
		testsrv = httptest.NewUnstartedServer(srv.Handler)
	}

	testsrv.EnableHTTP2 = true
	testsrv.StartTLS()

	return srv, testsrv
}
