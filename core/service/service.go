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

// Package service implements common service functionality.

package service

import "net/http"

// DefaultHTTPClient is an [http.Client] configured for HTTP/2 (TLS and unencrypted), which is
// required for Connect bidi streaming to services. Use this instead of [http.DefaultClient] as the
// default in service configs.
var DefaultHTTPClient = NewHTTPClient()

// NewHTTPClient returns a new [http.Client] configured for HTTP/2.
func NewHTTPClient() *http.Client {
	var (
		p *http.Protocols
	)

	p = new(http.Protocols)
	p.SetUnencryptedHTTP2(true)
	p.SetHTTP2(true)

	return &http.Client{Transport: &http.Transport{Protocols: p}}
}
