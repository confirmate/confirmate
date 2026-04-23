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
	"errors"
	"log/slog"
	"sync"
	"time"

	"confirmate.io/core/api/ontology"
)

// CollectorResult captures the outcome of a single collector execution.
type CollectorResult struct {
	CollectorID   string
	CollectorName string
	Resources     []ontology.IsResource
	Err           error
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
		ticker         *time.Ticker
		runResult      CollectionResult
		collectorNames []string
	)

	collectorNames = make([]string, len(svc.cfg.Collectors))
	for i := range svc.cfg.Collectors {
		collectorNames[i] = svc.cfg.Collectors[i].Name()
	}

	slog.Info("Starting collection loop",
		slog.Duration("interval", svc.cfg.Interval),
		slog.Any("collector_names", collectorNames),
	)

	defer close(resultCh)

	ticker = time.NewTicker(svc.cfg.Interval)
	defer ticker.Stop()

	runResult = svc.runOnce(ctx)

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
			runResult = svc.runOnce(ctx)

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
	return svc.runOnce(context.Background())
}

func (svc *Service) runOnce(ctx context.Context) (res CollectionResult) {
	var (
		wait    sync.WaitGroup
		results []CollectorResult
	)

	res.StartedAt = time.Now()
	results = make([]CollectorResult, len(svc.cfg.Collectors))

	for i := range svc.cfg.Collectors {
		collectorIndex := i
		collector := svc.cfg.Collectors[collectorIndex]

		// Run the collector in a separate goroutine to allow concurrent execution of all
		// collectors. The results are collected in the results slice, which is protected by the
		// wait group to ensure that all collectors have finished before the final result is
		// returned.
		wait.Add(1)
		go func() {
			defer wait.Done()

			var (
				resources  []ontology.IsResource
				collectErr error
				storeErr   error
			)

			resources, collectErr = collector.Collect()
			storeErr = svc.sendResourcesToEvidenceStore(ctx, collector, resources)

			results[collectorIndex] = CollectorResult{
				CollectorID:   collector.ID(),
				CollectorName: collector.Name(),
				Resources:     resources,
				Err:           errors.Join(collectErr, storeErr),
			}
		}()
	}

	wait.Wait()

	res.FinishedAt = time.Now()
	res.Collectors = results

	return res
}
