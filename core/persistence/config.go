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

package persistence

import "fmt"

// DefaultConfig contains the default [Config] for the persistence layer.
var DefaultConfig = Config{
	Host:             "localhost",
	Port:             5432,
	DBName:           "confirmate",
	User:             "confirmate",
	Password:         "confirmate",
	SSLMode:          "disable",
	MaxConn:          10,
	InMemoryDB:       false,
	Types:            []any{},
	CustomJoinTables: []CustomJoinTable{},
}

// Config contains configuration parameters for the persistence layer.
type Config struct {
	// Host is the host of the database server, unless [Config.InMemoryDB] is set to true.
	Host string

	// Port is the port of the database server, unless [Config.InMemoryDB] is set to true.
	Port int

	// DBName is the name of the database to connect to, unless [Config.InMemoryDB] is set to true.
	DBName string

	// Password is the password to use for authentication with the database server, unless
	// [Config.InMemoryDB] is set to true.
	Password string

	// User is the username to use for authentication with the database server, unless
	// [Config.InMemoryDB] is set to true.
	User string

	// SSLMode is the SSL mode to use for the database connection, unless [Config.InMemoryDB] is set
	// to true.
	SSLMode string

	// InMemoryDB indicates whether to use an in-memory database instead of the database server
	// described by [Config.Host], [Config.Port], [Config.DBName], [Config.User], [Config.Password]
	// and [Config.SSLMode].
	//
	// This also forces the maximum number of connections ([Config.MaxConn]) to 1.
	InMemoryDB bool

	// MaxConn is the maximum number of open connections to the database.
	MaxConn int

	// Types contains a list of all types that should be registered with the persistence layer.
	Types []any

	// CustomJoinTables contains a list of custom join tables to be registered with the persistence
	// layer.
	CustomJoinTables []CustomJoinTable
}

// buildDSN builds the Data Source Name (DSN) for connecting to the database, used by GORM.
func (db *Config) buildDSN() string {
	{
		return fmt.Sprintf("host=%s port=%d dbname=%s user=%s password=%s sslmode=%s",
			db.Host,
			db.Port,
			db.DBName,
			db.User,
			db.Password,
			db.SSLMode,
		)
	}
}
