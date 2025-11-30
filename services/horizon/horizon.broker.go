package horizon

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/nats-io/nats.go"
	"github.com/rotisserie/eris"
)

// MessageBrokerService defines the interface for pub/sub messaging systems.
type MessageBrokerService interface {
	Run(ctx context.Context) error
	Stop(ctx context.Context) error
	Publish(topic string, payload any) error
	Dispatch(topics []string, payload any) error
	Subscribe(topic string, handler func(any) error) error
}

// MessageBroker provides a NATS-based implementation of MessageBrokerService.
type MessageBroker struct {
	host     string
	port     int
	nc       *nats.Conn
	clientID string
	natsUser string
	natsPass string
}

func NewHorizonMessageBroker(host string, port int, clientID, natsUser, natsPass string) MessageBrokerService {
	return &MessageBroker{
		host:     host,
		port:     port,
		clientID: clientID,
		natsUser: natsUser,
		natsPass: natsPass,
	}
}

// Run starts the message broker connection to NATS.
func (h *MessageBroker) Run(_ context.Context) error {
	natsURL := fmt.Sprintf("nats://%s:%d", h.host, h.port)
	options := []nats.Option{
		nats.UserInfo(h.natsUser, h.natsPass),
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

// Stop closes the message broker connection to NATS.
func (h *MessageBroker) Stop(_ context.Context) error {
	if h.nc != nil {
		h.nc.Close()
		h.nc = nil
	}
	return nil
}

// Dispatch publishes the same payload to multiple topics efficiently.
func (h *MessageBroker) Dispatch(topics []string, payload any) error {
	if h.nc == nil {
		return eris.New("NATS connection not initialized")
	}

	// Marshal payload once, reuse for all topics
	data, err := json.Marshal(payload)
	if err != nil {
		return eris.Wrap(err, "failed to marshal payload")
	}

	msg := &nats.Msg{Data: data}

	for _, topic := range topics {
		msg.Subject = h.clientID + topic
		if err := h.nc.PublishMsg(msg); err != nil {
			return eris.Wrap(err, fmt.Sprintf("failed to publish to topic %s", topic))
		}
	}

	return nil
}

// Publish sends a payload to a single topic (high performance).
func (h *MessageBroker) Publish(topic string, payload any) error {
	if h.nc == nil {
		return eris.New("NATS connection not initialized")
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return eris.Wrap(err, "failed to marshal payload")
	}

	msg := &nats.Msg{
		Subject: h.formatTopic(topic),
		Data:    data,
	}

	if err := h.nc.PublishMsg(msg); err != nil {
		return eris.Wrap(err, fmt.Sprintf("failed to publish to topic %s", topic))
	}

	return nil
}

// Subscribe registers a handler function for messages on a specific topic.
func (h *MessageBroker) Subscribe(topic string, handler func(any) error) error {
	if h.nc == nil {
		return eris.New("NATS connection not initialized")
	}

	_, err := h.nc.Subscribe(h.formatTopic(topic), func(msg *nats.Msg) {
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

func (h *MessageBroker) formatTopic(topic string) string {
	return fmt.Sprintf("%s.%s", h.clientID, topic)
}
