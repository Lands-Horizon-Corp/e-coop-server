package horizon

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/nats-io/nats.go"
	"github.com/rotisserie/eris"
)

type MessageBrokerImpl struct {
	host     string
	port     int
	nc       *nats.Conn
	clientID string
	natsUser string
	natsPass string
}

func NewMessageBrokerImpl(host string, port int, clientID, natsUser, natsPass string) *MessageBrokerImpl {
	return &MessageBrokerImpl{
		host:     host,
		port:     port,
		clientID: clientID,
		natsUser: natsUser,
		natsPass: natsPass,
	}
}

func (h *MessageBrokerImpl) Run() error {
	natsURL := fmt.Sprintf("nats://%s:%d", h.host, h.port)
	options := []nats.Option{
		nats.UserInfo(h.natsUser, h.natsPass),
		nats.ErrorHandler(func(_ *nats.Conn, sub *nats.Subscription, err error) {
			log.Printf("Error in subscription to %s: %v\n", sub.Subject, err)
		}),
	}

	nc, err := nats.Connect(natsURL, options...)
	if err != nil {
		return eris.Wrap(err, "failed to connect to NATS")
	}
	h.nc = nc
	return nil
}

func (h *MessageBrokerImpl) Stop() error {
	if h.nc != nil {
		h.nc.Close()
		h.nc = nil
	}
	return nil
}

func (h *MessageBrokerImpl) Dispatch(topics []string, payload any) error {
	if h.nc == nil {
		if err := h.Run(); err != nil {
			return eris.Wrap(err, "Dispatch: NATS connection not initialized and failed to connect")
		}
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return eris.Wrap(err, "failed to marshal payload")
	}
	msg := &nats.Msg{Data: data}
	for _, topic := range topics {
		msg.Subject = h.formatTopic(topic)
		if err := h.nc.PublishMsg(msg); err != nil {
			return eris.Wrap(err, fmt.Sprintf("failed to publish to topic %s", topic))
		}
	}
	return nil
}

func (h *MessageBrokerImpl) Publish(topic string, payload any) error {
	if h.nc == nil {
		if err := h.Run(); err != nil {
			return eris.Wrap(err, "Publish: NATS connection not initialized and failed to connect")
		}
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return eris.Wrap(err, "failed to marshal payload")
	}
	if err := h.nc.PublishMsg(&nats.Msg{Subject: h.formatTopic(topic), Data: data}); err != nil {
		return eris.Wrap(err, fmt.Sprintf("failed to publish to topic %s", topic))
	}
	return nil
}

func (h *MessageBrokerImpl) Subscribe(topic string, handler func(any) error) error {
	if h.nc == nil {
		if err := h.Run(); err != nil {
			return eris.Wrap(err, "Subscribe: NATS connection not initialized and failed to connect")
		}
	}
	_, err := h.nc.Subscribe(h.formatTopic(topic), func(msg *nats.Msg) {
		var payload map[string]any
		if err := json.Unmarshal(msg.Data, &payload); err != nil {
			log.Printf("Failed to unmarshal message from topic %s: %v\n", topic, err)
			return
		}
		if err := handler(payload); err != nil {
			log.Printf("Handler error for topic %s: %v\n", topic, err)
		}
	})
	if err != nil {
		return eris.Wrap(err, fmt.Sprintf("failed to subscribe to topic %s", topic))
	}
	return nil
}

func (h *MessageBrokerImpl) formatTopic(topic string) string {
	return fmt.Sprintf("%s.%s", h.clientID, topic)
}
