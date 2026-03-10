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
	"log/slog"
	"sync"
	"time"

	"confirmate.io/core/api/ontology"
	"golang.org/x/sync/errgroup"
)

// CollectorResult captures the outcome of a single collector execution.
type CollectorResult struct {
	CollectorIndex int
	Resources      []ontology.IsResource
	Err            error
}

// CollectionResult captures the outcome of one full collection cycle.
type CollectionResult struct {
	StartedAt  time.Time
	FinishedAt time.Time
	Collectors []CollectorResult
}

// runLoop runs the collection loop, executing collectors immediately and then at the configured
// interval. It sends results to the provided channel and exits when the context is canceled.
func (svc *Service) runLoop(ctx context.Context, resultCh chan<- CollectionResult) {
	var (
		ticker    *time.Ticker
		runResult CollectionResult
	)

	slog.Info("Starting collection loop",
		slog.Duration("interval", svc.interval),
		slog.Int("num_collectors", len(svc.collectors)),
	)

	defer close(resultCh)

	ticker = time.NewTicker(svc.interval)
	defer ticker.Stop()

	runResult = svc.RunOnce()

	select {
	case <-ctx.Done():
		return
	case resultCh <- runResult:
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			runResult = svc.RunOnce()

			select {
			case <-ctx.Done():
				return
			case resultCh <- runResult:
			}
		}
	}
}

// RunOnce executes one full collection cycle concurrently for all configured collectors.
func (svc *Service) RunOnce() (res CollectionResult) {
	var (
		group   errgroup.Group
		mu      sync.Mutex
		results []CollectorResult
	)

	res.StartedAt = time.Now()
	results = make([]CollectorResult, 0, len(svc.collectors))

	for i := range svc.collectors {
		collectorIndex := i
		collector := svc.collectors[collectorIndex]

		group.Go(func() error {
			var (
				resources []ontology.IsResource
				err       error
			)

			resources, err = collector.Collect()

			mu.Lock()
			results = append(results, CollectorResult{
				CollectorIndex: collectorIndex,
				Resources:      resources,
				Err:            err,
			})
			mu.Unlock()

			// We intentionally return nil to avoid one collector failure affecting the others.
			return nil
		})
	}

	_ = group.Wait()

	res.FinishedAt = time.Now()
	res.Collectors = results

	return res
}
