package homeassistant

import (
	"context"
	"encoding/json"
	"strings"

	pb "github.com/c0deaddict/neon-display/hal_proto"
	"github.com/nats-io/nats.go"
)

type pirDevice struct {
	baseDevice
}

type pirConfig struct {
	Name        string `json:"name"`
	DeviceClass string `json:"device_class"`
	StateTopic  string `json:"state_topic"`
}

func newPirDevice(hostname string, id string) *pirDevice {
	return &pirDevice{baseDevice{"binary_sensor", hostname, id}}
}

func (p *pirDevice) subscribe(ctx context.Context, nc *nats.Conn) error {
	return nil
}

func (p *pirDevice) announce(ctx context.Context, nc *nats.Conn) error {
	cfg := pirConfig{
		Name:        p.baseDevice.hostname + " occupancy",
		DeviceClass: "motion",
		StateTopic:  strings.ReplaceAll(p.stateTopic(), ".", "/"),
	}
	data, err := json.Marshal(&cfg)
	if err != nil {
		return err
	}

	return nc.Publish(p.announceTopic(), data)
}

func (p *pirDevice) handleEvent(event *pb.Event, nc *nats.Conn) error {
	switch event.Source {
	case pb.EventSource_Pir:
		state := "OFF"
		if event.State {
			state = "ON"
		}
		return nc.Publish(p.stateTopic(), []byte(state))
	}
	return nil
}
