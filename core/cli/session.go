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
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"confirmate.io/core/api"

	"golang.org/x/oauth2"
)

var (
	// DefaultSessionFolder is the default directory for CLI sessions.
	DefaultSessionFolder string

	// ErrSessionNotFound is returned when no session file exists.
	ErrSessionNotFound = errors.New("session not found")
)

const SessionFolderFlag = "session-directory"

// Session represents an authenticated CLI session.
type Session struct {
	*oauth2.Config

	authorizer api.Authorizer

	// URL is the base URL of the API server.
	URL string `json:"url"`

	Folder string `json:"-"`

	dirty bool
}

func init() {
	if home, err := os.UserHomeDir(); err == nil {
		DefaultSessionFolder = filepath.Join(home, ".confirmate")
	}
}

// NewSession creates a new session and initializes the authorizer.
func NewSession(url string, config *oauth2.Config, token *oauth2.Token, folder string) (session *Session, err error) {
	session = &Session{
		URL:        url,
		Folder:     folder,
		Config:     config,
		authorizer: api.NewOAuthAuthorizerFromConfig(config, token),
	}

	return session, nil
}

// LoadSession loads a session from disk.
func LoadSession(folder string) (session *Session, err error) {
	var filePath string
	var file *os.File

	filePath = filepath.Join(folder, "session.json")

	file, err = os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrSessionNotFound
		}
		return nil, err
	}
	defer func() {
		_ = file.Close()
	}()

	session = new(Session)
	session.Folder = folder

	if err = json.NewDecoder(file).Decode(&session); err != nil {
		return nil, fmt.Errorf("could not parse session file: %w", err)
	}

	if session.dirty {
		_ = session.Save()
	}

	return session, nil
}

// Save writes the session to disk.
func (s *Session) Save() (err error) {
	var filePath string
	var file *os.File

	if err := os.MkdirAll(s.Folder, 0700); err != nil {
		return fmt.Errorf("could not create session directory: %w", err)
	}

	filePath = filepath.Join(s.Folder, "session.json")
	file, err = os.OpenFile(filePath, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("could not save session.json: %w", err)
	}
	defer func() {
		_ = file.Close()
	}()

	if err = json.NewEncoder(file).Encode(&s); err != nil {
		return fmt.Errorf("could not serialize JSON: %w", err)
	}

	s.dirty = false
	return nil
}

// Authorizer returns the session authorizer.
func (s *Session) Authorizer() (authorizer api.Authorizer) {
	return s.authorizer
}

// SetAuthorizer sets the session authorizer.
func (s *Session) SetAuthorizer(authorizer api.Authorizer) {
	s.authorizer = authorizer
}

// HTTPClient returns an HTTP client that injects OAuth2 tokens.
func (s *Session) HTTPClient(base *http.Client) (client *http.Client) {
	return api.NewOAuthHTTPClient(base, s.authorizer)
}

// MarshalJSON serializes a session with token and OAuth config.
func (s *Session) MarshalJSON() (b []byte, err error) {
	var token *oauth2.Token
	if s.authorizer != nil {
		token, _ = s.authorizer.Token()
	}

	b, err = json.Marshal(&struct {
		URL    string         `json:"url"`
		Token  *oauth2.Token  `json:"token"`
		Config *oauth2.Config `json:"oauth2"`
	}{
		URL:    s.URL,
		Token:  token,
		Config: s.Config,
	})
	return b, err
}

// UnmarshalJSON deserializes a session and rebuilds the authorizer.
func (s *Session) UnmarshalJSON(data []byte) (err error) {
	var v struct {
		URL    string         `json:"url"`
		Token  *oauth2.Token  `json:"token"`
		Config *oauth2.Config `json:"oauth2"`
	}

	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	s.URL = v.URL
	s.Config = v.Config
	s.authorizer = api.NewOAuthAuthorizerFromConfig(v.Config, v.Token)

	if s.authorizer == nil {
		s.dirty = true
	}

	return nil
}
