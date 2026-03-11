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
	"os"
	"strings"

	"confirmate.io/core/persistence"
	"confirmate.io/core/server"

	"github.com/urfave/cli/v3"
)

var (
	// logFlags contains the flags for configuring logging.
	logFlags = []cli.Flag{
		&cli.StringFlag{
			Name:    "log-level",
			Usage:   "Log level (TRACE, DEBUG, INFO, WARN, ERROR)",
			Value:   server.DefaultConfig.LogLevel,
			Sources: envVarSources("log-level"),
		},
	}

	// apiFlags contains the flags for configuring the API server (e.g., CORS settings).
	apiFlags = []cli.Flag{
		&cli.Uint16Flag{
			Name:    "api-port",
			Usage:   "Port to run the API server (Connect, gRPC, REST) on",
			Value:   server.DefaultConfig.Port,
			Sources: envVarSources("api-port"),
		},
		&cli.StringSliceFlag{
			Name:    "api-cors-allowed-origins",
			Usage:   "Specifies the origins allowed in CORS",
			Value:   server.DefaultConfig.CORS.AllowedOrigins,
			Sources: envVarSources("api-cors-allowed-origins"),
		},
		&cli.StringSliceFlag{
			Name:    "api-cors-allowed-methods",
			Usage:   "Specifies the methods allowed in CORS",
			Value:   server.DefaultConfig.CORS.AllowedMethods,
			Sources: envVarSources("api-cors-allowed-methods"),
		},
		&cli.StringSliceFlag{
			Name:    "api-cors-allowed-headers",
			Usage:   "Specifies the headers allowed in CORS",
			Value:   server.DefaultConfig.CORS.AllowedHeaders,
			Sources: envVarSources("api-cors-allowed-headers"),
		},
	}

	// authFlags contains the flags for configuring authentication and authorization for the
	// API server.
	authFlags = []cli.Flag{
		&cli.BoolFlag{
			Name:    "auth-enabled",
			Usage:   "Enable JWT authentication for RPC requests",
			Value:   false,
			Sources: envVarSources("auth-enabled"),
		},
		&cli.StringFlag{
			Name:    "auth-jwks-url",
			Usage:   "JWKS URL for JWT validation",
			Value:   server.DefaultJWKSURL,
			Sources: envVarSources("auth-jwks-url"),
		},
	}

	// serviceAuthFlags contains the flags for configuring service-to-service authentication using
	// OAuth 2.0 client credentials flow.
	serviceAuthFlags = []cli.Flag{
		&cli.StringFlag{
			Name:    "service-oauth2-token-endpoint",
			Usage:   "OAuth 2.0 token URL for service-to-service auth",
			Value:   DefaultServiceTokenURL,
			Sources: envVarSources("service-oauth2-token-endpoint"),
		},
		&cli.StringFlag{
			Name:    "service-oauth2-client-id",
			Usage:   "OAuth 2.0 client ID for service-to-service auth",
			Value:   DefaultServiceClientID,
			Sources: envVarSources("service-oauth2-client-id"),
		},
		&cli.StringFlag{
			Name:    "service-oauth2-client-secret",
			Usage:   "OAuth 2.0 client secret for service-to-service auth",
			Value:   DefaultServiceClientSecret,
			Sources: envVarSources("service-oauth2-client-secret"),
		},
	}

	// dbFlags contains the flags for configuring the database connection for services that require
	// persistence.
	dbFlags = []cli.Flag{
		&cli.StringFlag{
			Name:    "db-host",
			Usage:   "Specifies the server hostname",
			Value:   persistence.DefaultConfig.Host,
			Sources: envVarSources("db-host"),
		},
		&cli.IntFlag{
			Name:    "db-port",
			Usage:   "Specifies the server port",
			Value:   persistence.DefaultConfig.Port,
			Sources: envVarSources("db-port"),
		},
		&cli.StringFlag{
			Name:    "db-name",
			Usage:   "Specifies the database name",
			Value:   persistence.DefaultConfig.DBName,
			Sources: envVarSources("db-name"),
		},
		&cli.StringFlag{
			Name:    "db-user-name",
			Usage:   "Specifies the database user",
			Value:   persistence.DefaultConfig.User,
			Sources: envVarSources("db-user-name"),
		},
		&cli.StringFlag{
			Name:    "db-password",
			Usage:   "Specifies the database password",
			Value:   persistence.DefaultConfig.Password,
			Sources: envVarSources("db-password"),
		}, &cli.StringFlag{
			Name:    "db-ssl-mode",
			Usage:   "Specifies the database SSL mode (disable, require, verify-ca, verify-full)",
			Value:   persistence.DefaultConfig.SSLMode,
			Sources: envVarSources("db-ssl-mode"),
		}, &cli.BoolFlag{
			Name:    "db-in-memory",
			Usage:   "Use in-memory database instead of PostgreSQL (useful for testing)",
			Value:   persistence.DefaultConfig.InMemoryDB,
			Sources: envVarSources("db-in-memory"),
		}, &cli.IntFlag{
			Name:    "db-max-connections",
			Usage:   "Specifies the maximum number of database connections",
			Value:   persistence.DefaultConfig.MaxConn,
			Sources: envVarSources("db-max-connections"),
		},
	}
)

// envVarSources constructs a [cli.ValueSourceChain] that looks up the given flag name in
// environment variables with the prefix "CONFIRMATE_" and "CLOUDITOR_".
func envVarSources(flagName string) cli.ValueSourceChain {
	keys := []string{
		"CONFIRMATE_" + envVarSuffix(flagName),
		"CLOUDITOR_" + envVarSuffix(flagName),
	}

	return cli.EnvVars(keys...)
}

// envVarSuffix converts a flag name to an environment variable suffix by replacing dashes with
// underscores and converting to uppercase.
func envVarSuffix(flagName string) string {
	return strings.ToUpper(strings.ReplaceAll(flagName, "-", "_"))
}

// joinFlagSlices joins multiple cli.Flag slices into one slice while preserving order.
func joinFlagSlices(flagSlices ...[]cli.Flag) (flags []cli.Flag) {
	for _, flagSlice := range flagSlices {
		flags = append(flags, flagSlice...)
	}

	return flags
}

// ParseAndRun parses the command line arguments and runs the given command.
// If an error occurs, it is printed to stderr and the program exits with a non-zero
// status code.
//
// If the help flag is provided, the usage information is printed to stdout
// and the function returns without error.
func ParseAndRun(cmd *cli.Command) error {
	return cmd.Run(context.Background(), os.Args)
}
