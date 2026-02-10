// Copyright 2016-2026 Fraunhofer AISEC
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

package api

import (
	"context"
	"net/http"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

// Authorizer provides OAuth 2.0 tokens for authenticating client requests.
type Authorizer interface {
	oauth2.TokenSource
}

// UsesAuthorizer denotes a struct that can accept and use an Authorizer.
type UsesAuthorizer interface {
	SetAuthorizer(auth Authorizer)
	Authorizer() Authorizer
}

type oauthAuthorizer struct {
	oauth2.TokenSource
}

// NewOAuthAuthorizerFromClientCredentials creates a new authorizer based on OAuth 2.0 client credentials.
func NewOAuthAuthorizerFromClientCredentials(config *clientcredentials.Config) (authorizer Authorizer) {
	if config == nil {
		return nil
	}

	authorizer = &oauthAuthorizer{
		TokenSource: oauth2.ReuseTokenSource(nil, config.TokenSource(context.Background())),
	}

	return authorizer
}

// NewOAuthAuthorizerFromConfig creates a new authorizer based on an OAuth 2.0 config.
func NewOAuthAuthorizerFromConfig(config *oauth2.Config, token *oauth2.Token) (authorizer Authorizer) {
	if config == nil || token == nil {
		return nil
	}

	authorizer = &oauthAuthorizer{
		TokenSource: config.TokenSource(context.Background(), token),
	}

	return authorizer
}

// NewOAuthHTTPClient returns a copy of base client that injects OAuth 2.0 bearer tokens.
// If authorizer is nil, base is returned as-is (or http.DefaultClient if base is nil).
func NewOAuthHTTPClient(base *http.Client, authorizer Authorizer) (client *http.Client) {
	var transport http.RoundTripper
	var clientCopy http.Client

	if base == nil {
		base = http.DefaultClient
	}
	if authorizer == nil {
		return base
	}

	transport = base.Transport
	if transport == nil {
		transport = http.DefaultTransport
	}

	clientCopy = *base
	clientCopy.Transport = &oauth2.Transport{
		Source: authorizer,
		Base:   transport,
	}

	return &clientCopy
}
