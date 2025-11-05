package servertest

import (
	"net/http/httptest"

	"confirmate.io/core/server"
)

// NewInMemoryConnectServer creates a new in-memory Connect server for testing purposes. It returns
// the server instance and an [httptest.Server] that can be used to send requests to the server. The
// server is already started in the background.
//
// The server needs to stopped by the caller using the [server.Server]'s methods.
func NewTestConnectServer(opts ...server.Option) (
	srv *server.Server,
	testsrv *httptest.Server,
	err error,
) {
	srv = server.NewConnectServer(opts)
	testsrv = httptest.NewServer(srv.Handler)

	return srv, testsrv, nil
}
