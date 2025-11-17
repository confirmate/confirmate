package servertest

import (
	"net"
	"net/http/httptest"
	"testing"

	"confirmate.io/core/server"
	"confirmate.io/core/util/assert"
)

// NewTestConnectServer creates a new in-memory Connect server for testing purposes. It returns the
// server instance and an [httptest.Server] that can be used to send requests to the server. The
// server is already started in the background.
//
// The caller must close the returned [httptest.Server] using testsrv.Close() when done. This will
// fail the test if the server could not be created.
func NewTestConnectServer(t *testing.T, opts []server.Option) (srv *server.Server, testsrv *httptest.Server) {
	return newTestConnectServerInternal(t, opts, nil)
}

// NewTestConnectServerWithListener creates a new in-memory Connect server for testing purposes,
// using the specified listener. This allows restarting the server on the same address. It returns
// the server instance and an [httptest.Server] that can be used to send requests to the server.
// The server is already started in the background.
//
// The caller must close the returned [httptest.Server] using testsrv.Close() when done. This will
// fail the test if the server could not be created.
func NewTestConnectServerWithListener(t *testing.T, opts []server.Option, listener net.Listener) (srv *server.Server, testsrv *httptest.Server) {
	return newTestConnectServerInternal(t, opts, listener)
}

// newTestConnectServerInternal is the shared implementation for creating test Connect servers.
func newTestConnectServerInternal(t *testing.T, opts []server.Option, listener net.Listener) (srv *server.Server, testsrv *httptest.Server) {
	var err error

	srv, err = server.NewConnectServer(opts)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	if listener != nil {
		// Use provided listener
		testsrv = httptest.NewUnstartedServer(srv.Handler)
		testsrv.Listener = listener
		testsrv.Start()
	} else {
		// Create new server with random port
		testsrv = httptest.NewServer(srv.Handler)
	}

	return srv, testsrv
}
