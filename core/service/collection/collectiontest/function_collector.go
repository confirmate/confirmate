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

	"github.com/google/uuid"
)

// functionCollector is a simple implementation of the [collection.Collector] interface that allows
// defining the [collection.Collector.Collect] function as a field. This is useful for testing to
// create collectors with custom behavior without needing to define new types for each case.
type functionCollector struct {
	id      string
	name    string
	collect func() ([]ontology.IsResource, error)
}

// NewFunctionCollector creates a new [functionCollector] with the provided collect function.
func NewFunctionCollector(collect func() ([]ontology.IsResource, error)) collection.Collector {
	return NewNamedFunctionCollector("function-collector", collect)
}

// NewNamedFunctionCollector creates a new [functionCollector] with the provided name and collect
// function.
func NewNamedFunctionCollector(name string, collect func() ([]ontology.IsResource, error)) collection.Collector {
	if name == "" {
		name = "function-collector"
	}

	return &functionCollector{
		id:      uuid.NewString(),
		name:    name,
		collect: collect,
	}
}

func (c *functionCollector) ID() string {
	return c.id
}

func (c *functionCollector) Name() string {
	return c.name
}

func (c *functionCollector) Collect() (list []ontology.IsResource, err error) {
	if c.collect == nil {
		return nil, nil
	}

	return c.collect()
}
