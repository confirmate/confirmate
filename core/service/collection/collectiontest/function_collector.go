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

package collectiontest

import (
	"confirmate.io/core/api/ontology"
	"confirmate.io/core/service/collection"
)

// functionCollector is a simple implementation of the [collection.Collector] interface that allows
// defining the [collection.Collector.Collect] function as a field. This is useful for testing to
// create collectors with custom behavior without needing to define new types for each case.
type functionCollector struct {
	collect func() ([]ontology.IsResource, error)
}

// NewFunctionCollector creates a new [functionCollector] with the provided collect function.
func NewFunctionCollector(collect func() ([]ontology.IsResource, error)) collection.Collector {
	return &functionCollector{
		collect: collect,
	}
}

func (c functionCollector) Collect() (list []ontology.IsResource, err error) {
	if c.collect == nil {
		return nil, nil
	}

	return c.collect()
}
