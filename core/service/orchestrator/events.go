package orchestrator

import (
	"context"

	"confirmate.io/core/api/orchestrator"
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

// publishEvent publishes a change event to all subscribers.
func (svc *Service) publishEvent(event *orchestrator.ChangeEvent) {
	svc.subscribersMutex.RLock()
	defer svc.subscribersMutex.RUnlock()

	for _, sub := range svc.subscribers {
		// Check filter
		if sub.filter != nil && len(sub.filter.Types) > 0 {
			found := false
			for _, t := range sub.filter.Types {
				if t == event.Type {
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
