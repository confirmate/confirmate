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
<<<<<<< HEAD
=======

	"github.com/google/uuid"
>>>>>>> oxisto/collection-service
)

// functionCollector is a simple implementation of the [collection.Collector] interface that allows
// defining the [collection.Collector.Collect] function as a field. This is useful for testing to
// create collectors with custom behavior without needing to define new types for each case.
type functionCollector struct {
<<<<<<< HEAD
=======
	id      string
	name    string
>>>>>>> oxisto/collection-service
	collect func() ([]ontology.IsResource, error)
}

// NewFunctionCollector creates a new [functionCollector] with the provided collect function.
<<<<<<< HEAD
func NewFunctionCollector(collect func() ([]ontology.IsResource, error)) collection.Collector {
	return &functionCollector{
=======
func NewFunctionCollector(name string, collect func() ([]ontology.IsResource, error)) collection.Collector {
	return &functionCollector{
		id:      uuid.NewString(),
		name:    name,
>>>>>>> oxisto/collection-service
		collect: collect,
	}
}

<<<<<<< HEAD
func (c functionCollector) Collect() (list []ontology.IsResource, err error) {
=======
func (c *functionCollector) ID() string {
	return c.id
}

func (c *functionCollector) Name() string {
	return c.name
}

func (c *functionCollector) Collect() (list []ontology.IsResource, err error) {
>>>>>>> oxisto/collection-service
	if c.collect == nil {
		return nil, nil
	}

	return c.collect()
}
