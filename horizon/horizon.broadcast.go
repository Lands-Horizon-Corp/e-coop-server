package horizon

import (
	"encoding/json"
	"fmt"

	"github.com/nats-io/nats.go"
	"github.com/rotisserie/eris"
)

type HorizonBroadcast struct {
	config *HorizonConfig
	nc     *nats.Conn
}

func NewHorizonBroadcast(config *HorizonConfig) (*HorizonBroadcast, error) {
	return &HorizonBroadcast{
		config: config,
		nc:     nil,
	}, nil
}

func (hb *HorizonBroadcast) Run() error {
	natsURL := fmt.Sprintf("nats://%s:%d", hb.config.NATSHost, hb.config.NATSClientPort)
	nc, err := nats.Connect(natsURL)
	if err != nil {
		return eris.Wrap(err, "failed to connect to NATS")
	}
	hb.nc = nc
	return nil
}

func (hb *HorizonBroadcast) Stop() {
	if hb.nc != nil {
		hb.nc.Close()
		hb.nc = nil
	}
}
func (hb *HorizonBroadcast) Publish(topic string, payload any) error {
	if hb.nc == nil {
		return eris.New("NATS connection not initialized")
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return eris.Wrap(err, "failed to marshal payload for topic")
	}
	if err := hb.nc.Publish(topic, data); err != nil {
		return eris.Wrap(err, fmt.Sprintf("failed to publish to topic %s", topic))
	}
	return nil
}

func (hb *HorizonBroadcast) Dispatch(topics []string, payload any) error {
	if hb.nc == nil {
		return eris.New("NATS connection not initialized")
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return eris.Wrap(err, "failed to marshal payload for topics")
	}
	for _, topic := range topics {
		if err := hb.nc.Publish(topic, data); err != nil {
			return eris.Wrap(err, fmt.Sprintf("failed to publish to topic %s", topic))
		}
	}

	return nil
}
