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
	"fmt"
	"net"
	"net/http"
	"time"

	confcli "confirmate.io/core/cli"

	oauth2go "github.com/oxisto/oauth2go"
	"github.com/urfave/cli/v3"
	"golang.org/x/oauth2"
)

const (
	OAuth2AuthURLFlag  = "oauth2-auth-url"
	OAuth2TokenURLFlag = "oauth2-token-url"
	OAuth2ClientIDFlag = "oauth2-client-id"

	DefaultOAuth2AuthURL  = "http://localhost:8080/v1/auth/authorize"
	DefaultOAuth2TokenURL = "http://localhost:8080/v1/auth/token"
	DefaultOAuth2ClientID = "cli"

	DefaultCallbackServerAddress = "localhost:10000"
)

var (
	DefaultCallback   = fmt.Sprintf("http://%s/callback", DefaultCallbackServerAddress)
	VerifierGenerator = oauth2go.GenerateSecret
)

// LoginCommand returns the CLI login command.
func LoginCommand() (command *cli.Command) {
	command = &cli.Command{
		Name:  "login",
		Usage: "Log in to Confirmate",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  OAuth2AuthURLFlag,
				Usage: "OAuth 2.0 authorization URL",
				Value: DefaultOAuth2AuthURL,
			},
			&cli.StringFlag{
				Name:  OAuth2TokenURLFlag,
				Usage: "OAuth 2.0 token URL",
				Value: DefaultOAuth2TokenURL,
			},
			&cli.StringFlag{
				Name:  OAuth2ClientIDFlag,
				Usage: "OAuth 2.0 client ID",
				Value: DefaultOAuth2ClientID,
			},
		},
		Action: func(_ context.Context, cmd *cli.Command) (err error) {
			var config *oauth2.Config
			var srv *callbackServer
			var sock net.Listener
			var token *oauth2.Token
			var addr string
			var folder string
			var session *confcli.Session
			var code string

			config = &oauth2.Config{
				ClientID: cmd.String(OAuth2ClientIDFlag),
				Endpoint: oauth2.Endpoint{
					AuthURL:  cmd.String(OAuth2AuthURLFlag),
					TokenURL: cmd.String(OAuth2TokenURLFlag),
				},
				RedirectURL: DefaultCallback,
			}

			srv = newCallbackServer(config)
			defer srv.Close()

			go func() {
				sock, err = net.Listen("tcp", srv.Addr)
				if err != nil {
					fmt.Printf("Could not start web server for OAuth 2.0 authorization code flow: %v", err)
					return
				}
				_ = srv.Serve(sock)
			}()

			code = <-srv.code
			token, err = srv.config.Exchange(context.Background(), code,
				oauth2.SetAuthURLParam("code_verifier", srv.verifier),
			)
			if err != nil {
				return err
			}

			addr = cmd.Root().String("addr")
			folder = cmd.Root().String(confcli.SessionFolderFlag)

			session, err = confcli.NewSession(addr, config, token, folder)
			if err != nil {
				return err
			}

			if err = session.Save(); err != nil {
				return fmt.Errorf("could not save session: %w", err)
			}

			fmt.Print("\nLogin successful\n")
			return nil
		},
	}

	return command
}

type callbackServer struct {
	http.Server

	verifier string
	config   *oauth2.Config
	code     chan string
}

func newCallbackServer(config *oauth2.Config) (srv *callbackServer) {
	var mux *http.ServeMux
	var challenge string
	var url string

	mux = http.NewServeMux()

	srv = &callbackServer{
		Server: http.Server{
			Handler:           mux,
			Addr:              DefaultCallbackServerAddress,
			ReadHeaderTimeout: 2 * time.Second,
		},
		verifier: VerifierGenerator(),
		config:   config,
		code:     make(chan string),
	}

	mux.HandleFunc("/callback", srv.handleCallback)

	challenge = oauth2go.GenerateCodeChallenge(srv.verifier)
	url = srv.config.AuthCodeURL("",
		oauth2.SetAuthURLParam("code_challenge", challenge),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
	)

	fmt.Printf("Please open %s in your browser to continue\n", url)
	return srv
}

func (srv *callbackServer) handleCallback(w http.ResponseWriter, r *http.Request) {
	var code string

	_, _ = w.Write([]byte("Success. You can close this browser tab now"))
	code = r.URL.Query().Get("code")
	srv.code <- code
}
