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

package cli

import (
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"confirmate.io/core/util/assert"
	"golang.org/x/oauth2"
)

// helper that creates a dummy oauth2 configuration and token.
func makeDummyOAuth() (*oauth2.Config, *oauth2.Token) {
	cfg := &oauth2.Config{
		ClientID:     "id",
		ClientSecret: "secret",
		Endpoint: oauth2.Endpoint{
			TokenURL: "https://example.com/token",
		},
	}
	tok := &oauth2.Token{AccessToken: "abc-token"}
	return cfg, tok
}

func TestNewSession(t *testing.T) {
	cfg, tok := makeDummyOAuth()
	sess, err := NewSession("https://api.local", cfg, tok, "/tmp/folder")
	assert.NoError(t, err)
	assert.NotNil(t, sess)
	assert.Equal(t, "https://api.local", sess.URL)
	assert.Same(t, cfg, sess.Config)
	assert.Equal(t, "/tmp/folder", sess.Folder)
	// authorizer should be non-nil and return our token
	assert.NotNil(t, sess.Authorizer())
	gotTok, err := sess.Authorizer().Token()
	assert.NoError(t, err)
	assert.Equal(t, tok.AccessToken, gotTok.AccessToken)
}

func TestSaveLoadCycle(t *testing.T) {
	tmp := t.TempDir()
	cfg, tok := makeDummyOAuth()
	sess, _ := NewSession("https://api.local", cfg, tok, tmp)
	// Save to disk
	assert.NoError(t, sess.Save())

	// ensure file exists
	data, err := os.ReadFile(filepath.Join(tmp, "session.json"))
	assert.NoError(t, err)
	// file should be valid json and contain our URL
	assert.Contains(t, string(data), "https://api.local")

	// load again
	loaded, err := LoadSession(tmp)
	assert.NoError(t, err)
	assert.Equal(t, sess.URL, loaded.URL)
	// authorizer should still give us token
	gotTok, err := loaded.Authorizer().Token()
	assert.NoError(t, err)
	assert.Equal(t, tok.AccessToken, gotTok.AccessToken)
}

func TestLoadSessionNotFound(t *testing.T) {
	tmp := t.TempDir()
	_, err := LoadSession(filepath.Join(tmp, "nonexistent"))
	assert.ErrorIs(t, err, ErrSessionNotFound)
}

func TestMarshalUnmarshalRoundtrip(t *testing.T) {
	cfg, tok := makeDummyOAuth()
	sess, _ := NewSession("https://foo", cfg, tok, "/somewhere")
	bytes, err := sess.MarshalJSON()
	assert.NoError(t, err)

	var copy Session
	assert.NoError(t, copy.UnmarshalJSON(bytes))
	assert.Equal(t, "https://foo", copy.URL)
	// config pointer may differ after unmarshal; compare exported fields
	assert.Equal(t, cfg.ClientID, copy.Config.ClientID)
	assert.Equal(t, cfg.ClientSecret, copy.Config.ClientSecret)
	assert.Equal(t, cfg.Endpoint.TokenURL, copy.Config.Endpoint.TokenURL)
	assert.NotNil(t, copy.Authorizer())
	gotTok, err := copy.Authorizer().Token()
	assert.NoError(t, err)
	assert.Equal(t, tok.AccessToken, gotTok.AccessToken)
}

func TestUnmarshalSetsDirtyWhenMissingData(t *testing.T) {
	// Build JSON that lacks oauth2 config/token
	data := []byte(`{"url":"u","token":null,"oauth2":null}`)
	var sess Session
	assert.NoError(t, sess.UnmarshalJSON(data))
	// authorizer should be nil
	assert.Nil(t, sess.authorizer)
	// dirty flag should be true (unexported, inspect via reflect)
	v := reflect.ValueOf(&sess).Elem().FieldByName("dirty")
	assert.True(t, v.Bool())
}

func TestAuthorizerSetterGetter(t *testing.T) {
	sess := &Session{}
	assert.Nil(t, sess.Authorizer())
	src := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: "x"})
	sess.SetAuthorizer(src)
	// interfaces from different packages are distinct types to generics; compare directly
	assert.True(t, src == sess.Authorizer())
}


func TestHTTPClient(t *testing.T) {
	sess := &Session{}
	// nil authorizer -> default client
	def := sess.HTTPClient(nil)
	assert.Same(t, http.DefaultClient, def)

	// with authorizer
	tok := &oauth2.Token{AccessToken: "tok"}
	src := oauth2.StaticTokenSource(tok)
	sess.SetAuthorizer(src)
	client := sess.HTTPClient(&http.Client{})
	tr := client.Transport
	// should be oauth2.Transport
	otr, ok := tr.(*oauth2.Transport)
	assert.True(t, ok)
	assert.True(t, src == otr.Source)
}

// ensure file gets auto-saved if dirty on load. we simulate by manually
// writing a file with an empty config so that unmarshal marks dirty=true.
func TestLoadSessionAutoSave(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "session.json")
	initial := []byte(`{"url":"u","token":null,"oauth2":null}`)
	assert.NoError(t, os.WriteFile(path, initial, 0600))

	// load should not error and should write back (dirty->save)
	_, err := LoadSession(tmp)
	assert.NoError(t, err)
	// after load the file should still be valid json but not equal to initial
	after, err := os.ReadFile(path)
	assert.NoError(t, err)
	assert.NotEqual(t, string(initial), string(after))
}
