package horizon

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"

	"github.com/nats-io/nats.go"
	"github.com/rotisserie/eris"
)

// MessageBroker defines the interface for pub/sub messaging systems
type MessageBrokerService interface {
	Run(ctx context.Context) error
	Stop(ctx context.Context) error
	Publish(ctx context.Context, topic string, payload interface{}) error
	Dispatch(ctx context.Context, topics []string, payload interface{}) error
	Subscribe(ctx context.Context, topic string, handler func(interface{}) error) error
}

type HorizonMessageBroker struct {
	host     string
	port     int
	ssl      bool
	nc       *nats.Conn
	certPath string
	keyPath  string
}

// NewHorizonMessageBroker initializes the broker with optional TLS cert paths
func NewHorizonMessageBroker(host string, port int, ssl bool, certPath, keyPath string) MessageBrokerService {
	return &HorizonMessageBroker{
		host:     host,
		port:     port,
		ssl:      ssl,
		certPath: certPath,
		keyPath:  keyPath,
	}
}

func (h *HorizonMessageBroker) Run(ctx context.Context) error {
	var natsURL string
	if h.ssl {
		natsURL = fmt.Sprintf("tls://%s:%d", h.host, h.port)
	} else {
		natsURL = fmt.Sprintf("nats://%s:%d", h.host, h.port)
	}

	options := []nats.Option{
		nats.ErrorHandler(func(_ *nats.Conn, sub *nats.Subscription, err error) {
			fmt.Printf("Error in subscription to %s: %v\n", sub.Subject, err)
		}),
	}

	if h.ssl {
		cert, err := tls.LoadX509KeyPair(h.certPath, h.keyPath)
		if err != nil {
			return eris.Wrap(err, "failed to load TLS certificate and key")
		}

		tlsConfig := &tls.Config{
			Certificates:       []tls.Certificate{cert},
			InsecureSkipVerify: false,
		}

		options = append(options, nats.Secure(tlsConfig))
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

func (h *HorizonMessageBroker) Dispatch(ctx context.Context, topics []string, payload interface{}) error {
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

func (h *HorizonMessageBroker) Publish(ctx context.Context, topic string, payload interface{}) error {
	if h.nc == nil {
		return eris.New("NATS connection not initialized")
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return eris.Wrap(err, "failed to marshal payload")
	}

	if err := h.nc.Publish(topic, data); err != nil {
		return eris.Wrap(err, fmt.Sprintf("failed to publish to topic %s", topic))
	}

	return nil
}

func (h *HorizonMessageBroker) Subscribe(ctx context.Context, topic string, handler func(interface{}) error) error {
	if h.nc == nil {
		return eris.New("NATS connection not initialized")
	}

	_, err := h.nc.Subscribe(topic, func(msg *nats.Msg) {
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
