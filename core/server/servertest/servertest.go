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
func NewTestConnectServer(t *testing.T, opts ...server.Option) (srv *server.Server, testsrv *httptest.Server) {
	var (
		err error
	)

	srv, err = server.NewConnectServer(opts)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	testsrv = httptest.NewServer(srv.Handler)

	return srv, testsrv
}

// NewTestConnectServerWithListener creates a new in-memory Connect server with a provided listener.
// This allows restarting the server on the same address. It returns the server instance and an
// [httptest.Server] that can be used to send requests to the server. The server is already started.
//
// The caller must close the returned [httptest.Server] using testsrv.Close() when done. This will
// fail the test if the server could not be created.
func NewTestConnectServerWithListener(t *testing.T, listener net.Listener, opts ...server.Option) (srv *server.Server, testsrv *httptest.Server) {
	var (
		err error
	)

	srv, err = server.NewConnectServer(opts)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	testsrv = httptest.NewUnstartedServer(srv.Handler)
	testsrv.Listener = listener
	testsrv.Start()

	return srv, testsrv
}
