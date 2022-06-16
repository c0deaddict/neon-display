package homeassistant

import (
	"context"
	"fmt"
	"strings"

	pb "github.com/c0deaddict/neon-display/hal_proto"
	"github.com/nats-io/nats.go"
)

type device interface {
	String() string
	subscribe(ctx context.Context, nc *nats.Conn) error
	announce(ctx context.Context, nc *nats.Conn) error
	handleEvent(event *pb.Event, nc *nats.Conn) error
}

type baseDevice struct {
	typ      string
	hostname string
	id       string
}

func (d *baseDevice) String() string {
	return fmt.Sprintf("device %s %s %s", d.typ, d.hostname, d.id)
}

func (d *baseDevice) topic(action string) string {
	return strings.Join([]string{"homeassistant", d.typ, d.hostname, d.id, action}, ".")
}

func (d *baseDevice) stateTopic() string {
	return d.topic("state")
}

func (d *baseDevice) commandTopic() string {
	return d.topic("set")
}

func (d *baseDevice) announceTopic() string {
	return d.topic("config")
}

func (d *baseDevice) uniqueId() string {
	return d.hostname + "_" + d.id
}
