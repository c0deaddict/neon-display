package homeassistant

import (
	"context"
	"encoding/json"
	"strings"

	pb "github.com/c0deaddict/neon-display/hal_proto"
	"github.com/nats-io/nats.go"
)

type buttonDevice struct {
	baseDevice
	source pb.EventSource
}

type buttonConfig struct {
	Name       string `json:"name"`
	StateTopic string `json:"state_topic"`
}

func newButtonDevice(hostname string, id string, source pb.EventSource) *buttonDevice {
	return &buttonDevice{baseDevice{"binary_sensor", hostname, id}, source}
}

func (b *buttonDevice) subscribe(ctx context.Context, nc *nats.Conn) error {
	return nil
}

func (b *buttonDevice) announce(ctx context.Context, nc *nats.Conn) error {
	cfg := pirConfig{
		Name:       b.baseDevice.hostname + " " + b.baseDevice.id,
		StateTopic: strings.ReplaceAll(b.stateTopic(), ".", "/"),
	}
	data, err := json.Marshal(&cfg)
	if err != nil {
		return err
	}

	return nc.Publish(b.announceTopic(), data)
}

func (b *buttonDevice) handleEvent(event *pb.Event, nc *nats.Conn) error {
	if event.Source == b.source {
		state := "OFF"
		if event.State {
			state = "ON"
		}
		return nc.Publish(b.stateTopic(), []byte(state))
	}
	return nil
}
