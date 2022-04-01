package controlplane

import (
	"context"
	"fmt"
	"github.com/keptn/go-utils/pkg/api/models"
	"github.com/keptn/keptn/lib-cp-connector/pkg/nats"
	"log"
	"time"
)

type EventSenderKeyType struct{}

var EventSenderKey = EventSenderKeyType{}

type EventSender func(ce models.KeptnContextExtendedCE) error

type EventSource interface {
	Start(context.Context, chan models.KeptnContextExtendedCE) error
	OnSubscriptionUpdate([]models.EventSubscription)
	Subscriptions() []models.EventSubscription
	Sender() EventSender
}

type NATSEventSource struct {
	currentSubscriptions []models.EventSubscription
	connector            *nats.NatsConnector
	eventProcessFn       nats.ProcessEventFn
}

func NewNATSEventSource(natsConnector *nats.NatsConnector) *NATSEventSource {
	return &NATSEventSource{
		currentSubscriptions: []models.EventSubscription{},
		connector:            natsConnector,
		eventProcessFn:       func(event models.KeptnContextExtendedCE) error { return nil },
	}
}

func (n *NATSEventSource) Start(ctx context.Context, eventChannel chan models.KeptnContextExtendedCE) error {
	n.eventProcessFn = func(event models.KeptnContextExtendedCE) error {
		eventChannel <- event
		return nil
	}
	if err := n.connector.SubscribeMultiple(n.currentSubscriptions, n.eventProcessFn); err != nil {
		return fmt.Errorf("could not start NATS event source: %w", err)
	}
	return nil
}

func (n *NATSEventSource) OnSubscriptionUpdate(subscriptions []models.EventSubscription) {
	n.currentSubscriptions = subscriptions
	err := n.connector.UnsubscribeAll()
	if err != nil {
		log.Printf("error during handling of subscription update: %v\n", err)
	}
	if err := n.connector.SubscribeMultiple(n.currentSubscriptions, n.eventProcessFn); err != nil {
		log.Printf("error during handling of subscription update: %v\n", err)
	}
}

func (n *NATSEventSource) Subscriptions() []models.EventSubscription {
	return n.currentSubscriptions
}

func (n *NATSEventSource) Sender() EventSender {
	return n.connector.Publish
}

type HTTPEventSource struct {
	currentSubscriptions []models.EventSubscription
}

func (e *HTTPEventSource) Start(ctx context.Context, eventChannel chan models.KeptnContextExtendedCE) error {
	go func() {
		for _, sub := range e.currentSubscriptions {
			events := e.pollEvents(ctx, sub)
			for _, ev := range events {
				eventChannel <- ev
			}
		}
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(5 * time.Second):
				for _, sub := range e.currentSubscriptions {
					events := e.pollEvents(ctx, sub)
					for _, ev := range events {
						eventChannel <- ev
					}
				}
			}
		}
	}()
	return nil
}

func (e *HTTPEventSource) OnSubscriptionUpdate(subscriptions []models.EventSubscription) {
	e.currentSubscriptions = subscriptions
}

func (e *HTTPEventSource) Subscriptions() []models.EventSubscription {
	return e.currentSubscriptions
}

func (e *HTTPEventSource) pollEvents(ctx context.Context, subscription models.EventSubscription) []models.KeptnContextExtendedCE {
	return []models.KeptnContextExtendedCE{
		{ID: "eventID"},
	}
}
