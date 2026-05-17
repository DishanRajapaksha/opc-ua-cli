package uaclient

import (
	"context"
	"fmt"
	"time"

	"github.com/DishanRajapaksha/opc-ua-cli/internal/domain"
	"github.com/gopcua/opcua"
	"github.com/gopcua/opcua/monitor"
)

// Subscription wraps a server-side OPC UA subscription and exposes data changes as a channel.
type Subscription struct {
	Events <-chan domain.DataChange
	Errors <-chan error
	close  func()
}

func (s Subscription) Close() {
	if s.close != nil {
		s.close()
	}
}

func (s *Service) Monitor(ctx context.Context, nodes []string, interval time.Duration) (Subscription, error) {
	resolved := make([]string, 0, len(nodes))
	for _, node := range nodes {
		nodeID, _, err := s.ResolveNodeID(ctx, node)
		if err != nil {
			return Subscription{}, err
		}
		resolved = append(resolved, nodeID.String())
	}

	nodeMonitor, err := monitor.NewNodeMonitor(s.client)
	if err != nil {
		return Subscription{}, err
	}

	rawChanges := make(chan *monitor.DataChangeMessage, 32)
	subscription, err := nodeMonitor.ChanSubscribe(ctx, &opcua.SubscriptionParameters{Interval: interval}, rawChanges, resolved...)
	if err != nil {
		return Subscription{}, err
	}

	events := make(chan domain.DataChange, 32)
	errors := make(chan error, 8)

	go func() {
		defer close(events)
		defer close(errors)
		defer subscription.Unsubscribe(context.Background())

		for {
			select {
			case <-ctx.Done():
				return
			case change, ok := <-rawChanges:
				if !ok {
					return
				}
				if change == nil {
					continue
				}
				if change.Error != nil {
					select {
					case errors <- change.Error:
					case <-ctx.Done():
					}
					continue
				}

				value := interface{}(nil)
				if change.Value != nil {
					value = change.Value.Value()
				}

				event := domain.DataChange{
					NodeID:          fmt.Sprint(change.NodeID),
					Value:           value,
					SourceTimestamp: formatTimestamp(change.SourceTimestamp),
				}

				select {
				case events <- event:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return Subscription{
		Events: events,
		Errors: errors,
		close: func() {
			subscription.Unsubscribe(context.Background())
		},
	}, nil
}
