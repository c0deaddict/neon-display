package homeassistant

import (
	"bytes"
	"context"

	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog/log"

	pb "github.com/c0deaddict/neon-display/hal_proto"
)

const hostname = "neon"

type HomeAssistant struct {
	hal     pb.HalClient
	nc      *nats.Conn
	devices []device
}

func New(hal pb.HalClient, nc *nats.Conn) *HomeAssistant {
	devices := []device{
		newLedsDevice(hal, hostname, "leds"),
		newPirDevice(hostname, "pir"),
		newButtonDevice(hostname, "push-button-red", pb.EventSource_RedButton),
		newButtonDevice(hostname, "push-button-yellow", pb.EventSource_YellowButton),
	}
	return &HomeAssistant{hal, nc, devices}
}

func (h *HomeAssistant) announceAll(ctx context.Context) {
	for _, d := range h.devices {
		err := d.announce(ctx, h.nc)
		if err != nil {
			log.Error().Err(err).Msgf("homeassistant announce: %s", d)
		}
	}
}

func (h *HomeAssistant) Start(ctx context.Context) {
	_, err := h.nc.Subscribe("homeassistant.status", func(msg *nats.Msg) {
		if bytes.Compare(msg.Data, []byte("online")) == 0 {
			h.announceAll(ctx)

		}
	})
	if err != nil {
		log.Error().Err(err).Msg("nats subscribe to homeassistant.status")
	}

	for _, d := range h.devices {
		err := d.subscribe(ctx, h.nc)
		if err != nil {
			log.Error().Err(err).Msgf("homeassistant subscribe: %s", d)
		}
	}

	h.announceAll(ctx)
}

func (h *HomeAssistant) HandleEvent(event *pb.Event) {
	for _, d := range h.devices {
		if err := d.handleEvent(event, h.nc); err != nil {
			log.Error().Err(err).Msgf("homeassistant handleEvent: %s", d)
		}
	}
}
