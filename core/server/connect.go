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
	mux.Handle("/", transcoder)
	err = http.ListenAndServe(
		addr,
		// Use h2c so we can serve HTTP/2 without TLS.
		h2c.NewHandler(mux, &http2.Server{}),
	)

	return err
}
