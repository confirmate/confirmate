package servertest

import (
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
	var (
		err error
	)

	srv, err = server.NewConnectServer(opts)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	// Use NewUnstartedServer to configure HTTP/2 with TLS
	testsrv = httptest.NewUnstartedServer(srv.Handler)
	testsrv.EnableHTTP2 = true

	// Start the server
	testsrv.StartTLS()

	return srv, testsrv
}
