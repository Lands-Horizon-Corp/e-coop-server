package horizon

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/nats-io/nats.go"
	"github.com/rotisserie/eris"
)

// MessageBroker defines the interface for pub/sub messaging systems
type MessageBrokerService interface {
	// Run connects to a broker cluster
	Run(ctx context.Context) error

	// Stop closes all producer/consumer connections
	Stop(ctx context.Context) error

	// Publish sends a message to a single topic
	Publish(ctx context.Context, topic string, payload any) error

	// DispatchBatch sends a message to multiple topics
	Dispatch(ctx context.Context, topics []string, payload any) error

	// Subscribe registers a message handler for a topic
	Subscribe(ctx context.Context, topic string, handler func(any) error) error
}

type HorizonMessageBroker struct {
	host string
	port int
	nc   *nats.Conn
}

func NewHorizonMessageBroker(host string, port int) MessageBrokerService {
	return &HorizonMessageBroker{
		host: host,
		port: port,
	}
}

// Run implements MessageBroker.
func (h *HorizonMessageBroker) Run(ctx context.Context) error {
	natsURL := fmt.Sprintf("nats://%s:%d", h.host, h.port)
	nc, err := nats.Connect(natsURL)
	if err != nil {
		return eris.Wrap(err, "failed to connect to NATS")
	}
	h.nc = nc
	return nil
}

// Stop implements MessageBroker.
func (h *HorizonMessageBroker) Stop(ctx context.Context) error {
	if h.nc != nil {
		h.nc.Close()
		h.nc = nil
	}
	return nil
}

// DispatchBatch implements MessageBroker.
func (h *HorizonMessageBroker) Dispatch(ctx context.Context, topics []string, payload any) error {
	if h.nc == nil {
		return eris.New("NATS connection not initialized")
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return eris.Wrap(err, "failed to marshal payload for topics")
	}
	for _, topic := range topics {
		if err := h.nc.Publish(topic, data); err != nil {
			return eris.Wrap(err, fmt.Sprintf("failed to publish to topic %s", topic))
		}
	}

	return nil
}

// Publish implements MessageBroker.
func (h *HorizonMessageBroker) Publish(ctx context.Context, topic string, payload any) error {
	if h.nc == nil {
		return eris.New("NATS connection not initialized")
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return eris.Wrap(err, "failed to marshal payload for topic")
	}
	if err := h.nc.Publish(topic, data); err != nil {
		return eris.Wrap(err, fmt.Sprintf("failed to publish to topic %s", topic))
	}
	return nil
}

// Subscribe implements MessageBroker.
func (h *HorizonMessageBroker) Subscribe(ctx context.Context, topic string, handler func(any) error) error {
	if h.nc == nil {
		return eris.New("NATS connection not initialized")
	}
	_, err := h.nc.Subscribe(topic, func(msg *nats.Msg) {
		var payload map[string]interface{}
		if err := json.Unmarshal(msg.Data, &payload); err != nil {
			fmt.Printf("failed to unmarshal message from topic %s: %v\n", topic, err)
			return
		}
		if err := handler(payload); err != nil {
			fmt.Printf("handler error for topic %s: %v\n", topic, err)
		}
	})

	if err != nil {
		return eris.Wrap(err, fmt.Sprintf("failed to subscribe to topic %s", topic))
	}
	return nil
}
