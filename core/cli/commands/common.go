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

package commands

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"confirmate.io/core/api/orchestrator/orchestratorconnect"
	confcli "confirmate.io/core/cli"
	"github.com/hokaccha/go-prettyjson"
	"github.com/urfave/cli/v3"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type httpClientKey struct{}

// WithHTTPClient returns a new context carrying an HTTP client override.
// If client is nil, ctx is returned unchanged.
func WithHTTPClient(ctx context.Context, client *http.Client) (out context.Context) {
	if client == nil {
		return ctx
	}

	out = context.WithValue(ctx, httpClientKey{}, client)
	return out
}

// httpClientFromContext extracts an HTTP client from the context.
// If no client is found, [http.DefaultClient] is returned.
func httpClientFromContext(ctx context.Context) (*http.Client, bool) {
	if ctx != nil {
		if client, ok := ctx.Value(httpClientKey{}).(*http.Client); ok && client != nil {
			return client, true
		}
	}

	return http.DefaultClient, false
}

// OrchestratorClient returns an orchestrator client. It is configured by the
// "addr" flag and its HTTP client can be overriden by setting an
// [httpClientKey] in the ctx.
func OrchestratorClient(ctx context.Context, c *cli.Command) (client orchestratorconnect.OrchestratorClient) {
	var httpClient *http.Client
	var overridden bool
	var session *confcli.Session
	var err error

	httpClient, overridden = httpClientFromContext(ctx)
	if !overridden {
		session, err = confcli.LoadSession(c.Root().String(confcli.SessionFolderFlag))
		if err == nil && session != nil {
			httpClient = session.HTTPClient(httpClient)
		}
	}

	client = orchestratorconnect.NewOrchestratorClient(httpClient, c.Root().String("addr"))
	return client
}

// ExpandCommaSeparated flattens values that may contain comma-separated items.
func ExpandCommaSeparated(values []string) (out []string) {
	if len(values) == 0 {
		return nil
	}

	for _, value := range values {
		for _, part := range strings.Split(value, ",") {
			var item string
			item = strings.TrimSpace(part)
			if item != "" {
				out = append(out, item)
			}
		}
	}
	return out
}

// PrettyPrint prints a proto message as pretty-printed JSON to stdout.
func PrettyPrint(msg proto.Message) (err error) {
	var m protojson.MarshalOptions
	var b []byte
	var out []byte

	m = protojson.MarshalOptions{EmitUnpopulated: false}

	b, err = m.Marshal(msg)
	if err != nil {
		return err
	}

	out, err = prettyjson.Format(b)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintln(os.Stdout, string(out))
	return err
}

// PaginationFlags returns a slice of common pagination flags.
func PaginationFlags() []cli.Flag {
	return []cli.Flag{
		&cli.IntFlag{
			Name:    "page-size",
			Aliases: []string{"n"},
			Usage:   "Number of items to return",
			Value:   10,
		},
		&cli.StringFlag{
			Name:    "page-token",
			Aliases: []string{"p"},
			Usage:   "Page token for the next page",
		},
	}
}
