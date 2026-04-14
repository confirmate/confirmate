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

package commands

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	confcli "confirmate.io/core/cli"
	"confirmate.io/core/util/assert"

	"github.com/urfave/cli/v3"
)

func TestLoginCommand_Action_Flow_Integration(t *testing.T) {
	var (
		originalVerifierGenerator = VerifierGenerator
		originalStateGenerator    = StateGenerator
		verifier                  = "verifier-test"
		state                     = "state-test"
		tokenCode                 string
		tokenVerifier             string
		mu                        sync.Mutex
	)

	VerifierGenerator = func() string { return verifier }
	StateGenerator = func() string { return state }
	t.Cleanup(func() {
		VerifierGenerator = originalVerifierGenerator
		StateGenerator = originalStateGenerator
	})

	tokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		if err := r.ParseForm(); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		mu.Lock()
		tokenCode = r.FormValue("code")
		tokenVerifier = r.FormValue("code_verifier")
		mu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token":  "access-token-1",
			"token_type":    "Bearer",
			"refresh_token": "refresh-token-1",
			"expires_in":    3600,
		})
	}))
	defer tokenServer.Close()

	folder := t.TempDir()
	root := &cli.Command{
		Name: "cf",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "addr", Value: "http://localhost:8080"},
			&cli.StringFlag{Name: confcli.SessionFolderFlag, Value: folder},
		},
		Commands: []*cli.Command{LoginCommand()},
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- root.Run(context.Background(), []string{
			"cf", "login",
			"--" + OAuth2AuthURLFlag, tokenServer.URL + "/authorize",
			"--" + OAuth2TokenURLFlag, tokenServer.URL + "/token",
			"--" + OAuth2ClientIDFlag, "cli-test",
		})
	}()

	callbackURL := "http://" + DefaultCallbackServerAddress + "/callback?state=" + url.QueryEscape(state) + "&code=auth-code-1"

	var callbackErr error
	for range 30 {
		_, callbackErr = http.Get(callbackURL)
		if callbackErr == nil {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	if callbackErr != nil {
		t.Fatalf("failed to call callback endpoint: %v", callbackErr)
	}

	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("login command failed: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("login command timed out")
	}

	mu.Lock()
	assert.Equal(t, "auth-code-1", tokenCode)
	assert.Equal(t, verifier, tokenVerifier)
	mu.Unlock()

	_, err := os.Stat(filepath.Join(folder, "session.json"))
	assert.NoError(t, err)

	session, err := confcli.LoadSession(folder)
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, "http://localhost:8080", session.URL)
}
