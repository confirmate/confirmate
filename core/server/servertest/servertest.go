package servertest

import (
	"net"
	"net/http/httptest"
	"testing"

	"confirmate.io/core/server"
	"confirmate.io/core/util/assert"
)

// testServerOptions holds configuration for creating test servers.
type testServerOptions struct {
	listener net.Listener
}

// TestServerOption is a functional option for configuring test server creation.
type TestServerOption func(*testServerOptions)

// WithListener configures the test server to use a specific listener.
// This allows restarting the server on the same address.
func WithListener(listener net.Listener) TestServerOption {
	return func(opts *testServerOptions) {
		opts.listener = listener
	}
}

// NewTestConnectServer creates a new in-memory Connect server for testing purposes. It returns the
// server instance and an [httptest.Server] that can be used to send requests to the server. The
// server is already started in the background.
//
// If WithListener option is provided, the server will use the specified listener, allowing it to
// restart on the same address. Otherwise, a new random port will be used.
//
// The caller must close the returned [httptest.Server] using testsrv.Close() when done. This will
// fail the test if the server could not be created.
func NewTestConnectServer(t *testing.T, serverOpts []server.Option, testOpts ...TestServerOption) (srv *server.Server, testsrv *httptest.Server) {
	var (
		err     error
		options testServerOptions
	)

	// Apply test server options
	for _, opt := range testOpts {
		opt(&options)
	}

	srv, err = server.NewConnectServer(serverOpts)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	if options.listener != nil {
		// Use provided listener
		testsrv = httptest.NewUnstartedServer(srv.Handler)
		testsrv.Listener = options.listener
		testsrv.Start()
	} else {
		// Create new server with random port
		testsrv = httptest.NewServer(srv.Handler)
	}

	return srv, testsrv
}
