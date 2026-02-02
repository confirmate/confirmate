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

package assessment

import (
	"context"
	"log/slog"
	"time"

	"confirmate.io/core/api/evidence"
	"confirmate.io/core/api/ontology"
)

// waitingRequest contains all information of an evidence request that still waits for
// more data
type waitingRequest struct {
	*evidence.Evidence

	started time.Time

	// waitingFor should ideally be empty at some point
	waitingFor map[string]bool

	resourceId string

	s *Service

	newResources chan string
	ctx          context.Context
}

func (l *waitingRequest) WaitAndHandle() {
	var (
		resource   string
		additional map[string]ontology.IsResource
		e          *evidence.Evidence
		ok         bool
		msg        ontology.IsResource
		duration   time.Duration
	)

	for {
		// Wait for an incoming resource
		resource = <-l.newResources

		// Check, if the incoming resource is of interest for us
		delete(l.waitingFor, resource)

		// Are we ready to assess?
		if len(l.waitingFor) == 0 {
			slog.Info("Evidence is now ready to assess", slog.Any("Evidence", l.Evidence.Id))

			// Gather our additional resources
			additional = make(map[string]ontology.IsResource)

			for _, r := range l.Evidence.ExperimentalRelatedResourceIds {
				l.s.em.RLock()

				e, ok = l.s.evidenceResourceMap[r]
				l.s.em.RUnlock()

				if !ok {
					slog.Error("Apparently, we are missing an evidence for a resource which we are supposed to have", slog.Any("Resource", r))
					break
				}

				msg = e.GetOntologyResource()
				if msg == nil {
					break
				}

				additional[r] = msg
			}

			// Let's go
			_, _ = l.s.handleEvidence(l.ctx, l.Evidence, l.Evidence.GetOntologyResource(), additional)

			duration = time.Since(l.started)

			slog.Info("Evidence was waiting", slog.String("evidenceId", l.Evidence.Id), slog.Duration("duration", duration))
			break
		}
	}

	// Lock requests for writing
	l.s.rm.Lock()
	// Remove ourselves from the list of requests
	delete(l.s.requests, l.Evidence.Id)
	// Unlock writing
	l.s.rm.Unlock()

	// Inform our wait group, that we are done
	l.s.wg.Done()
}

// informWaitingRequests informs any waiting requests of the arrival of a new resource ID, so that they might update
// their waiting decision.
func (svc *Service) informWaitingRequests(resourceId string) {
	// Lock requests for reading
	svc.rm.RLock()
	// Defer unlock at the exit of the go-routine
	defer svc.rm.RUnlock()
	for _, l := range svc.requests {
		if l.resourceId != resourceId {
			l.newResources <- resourceId
		}
	}
}
