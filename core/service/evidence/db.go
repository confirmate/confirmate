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

package evidence

import (
	"confirmate.io/core/api/evidence"
)

// graphEdge is the GORM model for the graph_edges table. It stores pre-computed
// edges between resources, derived from ontology _id/_ids fields when a resource
// snapshot is saved. This avoids loading all snapshots into memory on every
// ListGraphEdges call.
type graphEdge struct {
	Id     string `gorm:"primaryKey"`
	Source string `gorm:"index:idx_graph_edges_source"`
	Target string `gorm:"index:idx_graph_edges_target"`
	Type   string
}

func (graphEdge) TableName() string {
	return "graph_edges"
}

var types = []any{
	&evidence.Evidence{},
	&evidence.ResourceSnapshot{},
	&graphEdge{},
}
