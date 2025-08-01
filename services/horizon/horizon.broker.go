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
	Run(ctx context.Context) error
	Stop(ctx context.Context) error
	Publish(ctx context.Context, topic string, payload any) error
	Dispatch(ctx context.Context, topics []string, payload any) error
	Subscribe(ctx context.Context, topic string, handler func(any) error) error
}

type HorizonMessageBroker struct {
	host     string
	port     int
	nc       *nats.Conn
	clientId string
}

// NewHorizonMessageBroker initializes the broker with optional TLS cert paths
func NewHorizonMessageBroker(host string, port int, clientId string) MessageBrokerService {
	return &HorizonMessageBroker{
		host:     host,
		port:     port,
		clientId: clientId,
	}
}

func (h *HorizonMessageBroker) Run(ctx context.Context) error {
	natsURL := fmt.Sprintf("nats://%s:%d", h.host, h.port)
	options := []nats.Option{
		nats.ErrorHandler(func(_ *nats.Conn, sub *nats.Subscription, err error) {
			fmt.Printf("Error in subscription to %s: %v\n", sub.Subject, err)
		}),
	}

	nc, err := nats.Connect(natsURL, options...)
	if err != nil {
		return eris.Wrap(err, "failed to connect to NATS")
	}
	h.nc = nc
	return nil
}

func (h *HorizonMessageBroker) Stop(ctx context.Context) error {
	if h.nc != nil {
		h.nc.Close()
		h.nc = nil
	}
	return nil
}

func (h *HorizonMessageBroker) Dispatch(ctx context.Context, topics []string, payload any) error {
	if h.nc == nil {
		return eris.New("NATS connection not initialized")
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return eris.Wrap(err, "failed to marshal payload")
	}

	for _, topic := range topics {
		if err := h.nc.Publish(topic, data); err != nil {
			return eris.Wrap(err, fmt.Sprintf("failed to publish to topic %s", topic))
		}
	}
	return nil
}

func (h *HorizonMessageBroker) Publish(ctx context.Context, topic string, payload any) error {
	if h.nc == nil {
		return eris.New("NATS connection not initialized")
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return eris.Wrap(err, "failed to marshal payload")
	}

	if err := h.nc.Publish(topic+h.clientId, data); err != nil {
		return eris.Wrap(err, fmt.Sprintf("failed to publish to topic %s", topic))
	}

	return nil
}

func (h *HorizonMessageBroker) Subscribe(ctx context.Context, topic string, handler func(any) error) error {
	if h.nc == nil {
		return eris.New("NATS connection not initialized")
	}

	_, err := h.nc.Subscribe(topic+h.clientId, func(msg *nats.Msg) {
		var payload map[string]any
		if err := json.Unmarshal(msg.Data, &payload); err != nil {
			fmt.Printf("Failed to unmarshal message from topic %s: %v\n", topic, err)
			return
		}

		if err := handler(payload); err != nil {
			fmt.Printf("Handler error for topic %s: %v\n", topic, err)
		}
	})

	if err != nil {
		return eris.Wrap(err, fmt.Sprintf("failed to subscribe to topic %s", topic))
	}

	return nil
}
