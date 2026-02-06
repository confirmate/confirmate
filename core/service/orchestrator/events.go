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
	"context"
	"log/slog"

	"confirmate.io/core/api/orchestrator"
	"confirmate.io/core/log"
	"confirmate.io/core/service"

	"connectrpc.com/connect"
)

// Subscribe subscribes to change events.
func (svc *Service) Subscribe(
	ctx context.Context,
	req *connect.Request[orchestrator.SubscribeRequest],
	stream *connect.ServerStream[orchestrator.ChangeEvent],
) error {
	ch, id := svc.RegisterSubscriber(req.Msg.Filter)

	// Ensure cleanup on return
	defer svc.UnregisterSubscriber(id)

	// Send events to the stream
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case event, ok := <-ch:
			if !ok {
				return nil
			}
			if err := stream.Send(event); err != nil {
				return err
			}
		}
	}
}

// RegisterSubscriber registers a new subscriber for change events.
func (svc *Service) RegisterSubscriber(filter *orchestrator.SubscribeRequest_Filter) (<-chan *orchestrator.ChangeEvent, int64) {
	// Create a channel for this subscriber
	ch := make(chan *orchestrator.ChangeEvent, 100)

	svc.subscribersMutex.Lock()
	id := svc.nextSubscriberId
	svc.nextSubscriberId++

	// Register the subscriber
	svc.subscribers[id] = &subscriber{
		ch:     ch,
		filter: filter,
	}
	svc.subscribersMutex.Unlock()

	slog.Info("Registered subscriber", "id", id, "filter", filter)

	return ch, id
}

// UnregisterSubscriber un-registers a subscriber.
func (svc *Service) UnregisterSubscriber(id int64) {
	svc.subscribersMutex.Lock()
	defer svc.subscribersMutex.Unlock()

	if sub, ok := svc.subscribers[id]; ok {
		delete(svc.subscribers, id)
		close(sub.ch)
	}
}

// publishEvent publishes a [orchestrator.ChangeEvent] to all subscribers.
func (svc *Service) publishEvent(event *orchestrator.ChangeEvent) {
	svc.subscribersMutex.RLock()
	defer svc.subscribersMutex.RUnlock()

	if err := service.ValidateEvent(event); err != nil {
		slog.Error("Attempted to publish invalid event", "event", event, log.Err(err))
		return
	}

	for _, sub := range svc.subscribers {
		// Check category filter
		if sub.filter != nil && len(sub.filter.Categories) > 0 {
			found := false
			for _, c := range sub.filter.Categories {
				if c == event.Category {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		select {
		case sub.ch <- event:
		default:
			// Channel is full, skip this subscriber to avoid blocking
		}
	}
}
