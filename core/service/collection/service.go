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

package collection

import (
	"context"
	"fmt"
	"time"

	"confirmate.io/core/api/ontology"
)

// Collector is the interface that all collectors must implement. A collector is responsible for
// collecting evidence and translating them to ontology resources.
type Collector interface {
	// Collect executes the collection process and returns a list of collected resources or an error
	// if the collection failed.
	Collect() (list []ontology.IsResource, err error)
}

// Service is the service implementation for the collection service. It has one or more collectors,
// which do the actual work or collecting evidence. The service itself is responsible for
// orchestrating the extractors and
type Service struct {
	interval   time.Duration
	collectors []Collector
}

// DefaultConfig is the default configuration for the discovery service.
var DefaultConfig = Config{
	Interval: 5 * time.Minute,
}

// Config is the configuration for the collection service.
type Config struct {
	// Interval defines how often collectors are executed.
	Interval time.Duration

	// Collectors is a list of collectors to use for collecting evidence. At least one collector
	// must be provided.
	Collectors []Collector
}

// NewService creates a new collection service with the given configuration.
func NewService(cfg Config) (svc *Service, err error) {
	if cfg.Interval <= 0 {
		cfg.Interval = DefaultConfig.Interval
	}

	if len(cfg.Collectors) == 0 {
		return nil, fmt.Errorf("at least one collector must be provided")
	}

	svc = &Service{
		interval:   cfg.Interval,
		collectors: cfg.Collectors,
	}

	return svc, nil
}

// Start runs all collectors immediately and then repeatedly at the configured interval.
// The returned channel is closed when ctx is canceled.
func (svc *Service) Start(ctx context.Context) (resultCh <-chan CollectionResult) {
	var (
		results chan CollectionResult
	)

	results = make(chan CollectionResult)

	go func() {
		svc.runLoop(ctx, results)
	}()

	return results
}
