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

package orchestrator

import (
	"fmt"

	"confirmate.io/core/persistence"
)

// identifiable is implemented by all persisted messages that carry a string ID.
type identifiable interface {
	GetId() string
}

// loadByIds loads the full records for the given IDs and returns them in the order of ids. IDs
// without a matching record are skipped.
//
// It is the second step of the "latest by ..." queries: the DISTINCT ON itself only selects IDs,
// so that PostgreSQL can answer it from the (narrow) index alone instead of reading every heap
// row (including their large JSON/bytea columns) that it has to walk past to deduplicate and to
// honor OFFSET. The few rows of the requested page are then loaded by primary key.
func loadByIds[T identifiable](db persistence.DB, ids []string) (rows []T, err error) {
	var (
		loaded []T
		byId   map[string]T
		query  []string
		args   []any
	)

	if len(ids) == 0 {
		return nil, nil
	}

	query, args = persistence.AppendObjectIds(ids, query, args, "id")
	err = db.List(&loaded, "id", true, 0, len(ids), persistence.BuildConds(query, args)...)
	if err != nil {
		return nil, fmt.Errorf("could not load records by ID: %w", err)
	}

	byId = make(map[string]T, len(loaded))
	for _, r := range loaded {
		byId[r.GetId()] = r
	}

	rows = make([]T, 0, len(ids))
	for _, id := range ids {
		if r, ok := byId[id]; ok {
			rows = append(rows, r)
		}
	}

	return rows, nil
}
