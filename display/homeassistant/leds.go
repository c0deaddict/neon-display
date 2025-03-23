package homeassistant

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/types/known/emptypb"

	pb "github.com/c0deaddict/neon-display/hal_proto"
)

type ledsDevice struct {
	baseDevice
	hal pb.HalClient
}

type ledsConfig struct {
	Name                string   `json:"name"`
	UniqueId            string   `json:"unique_id"`
	StateTopic          string   `json:"state_topic"`
	CommandTopic        string   `json:"command_topic"`
	Schema              string   `json:"schema"`
	Brightness          bool     `json:"brightness"`
	Effect              bool     `json:"effect"`
	ColorMode           bool     `json:"color_mode"`
	SupportedColorModes []string `json:"supported_color_modes"`
	EffectList          []string `json:"effect_list"`
}

type color struct {
	R uint8 `json:"r"`
	G uint8 `json:"g"`
	B uint8 `json:"b"`
}

type state struct {
	Brightness uint   `json:"brightness"`
	ColorMode  string `json:"color_mode"`
	Color      color  `json:"color"`
	Effect     string `json:"effect"`
	State      string `json:"state"`
}

type command struct {
	Brightness *int    `json:"brightness"`
	Color      *color  `json:"color"`
	Effect     *string `json:"effect"`
	State      *string `json:"state"`
}

func (c color) uint32() uint32 {
	return uint32(c.R&0xff)<<16 | uint32(c.G&0xff)<<8 | uint32(c.B&0xff)
}

func (c command) ledState() *pb.LedState {
	s := pb.LedState{Effect: c.Effect}
	if c.State != nil {
		state := strings.ToLower(*c.State) == "on"
		s.State = &state
	}
	if c.Brightness != nil {
		brightness := uint32(*c.Brightness)
		s.Brightness = &brightness
	}
	if c.Color != nil {
		color := c.Color.uint32()
		s.Color = &color
	}
	return &s
}

func newLedsDevice(hal pb.HalClient, hostname string, id string) *ledsDevice {
	return &ledsDevice{baseDevice{"light", hostname, id}, hal}
}

func (l *ledsDevice) subscribe(ctx context.Context, nc *nats.Conn) error {
	// TODO: resubscribe when nats connection is lost
	_, err := nc.Subscribe(l.commandTopic(), func(msg *nats.Msg) {
		var command command
		err := json.Unmarshal(msg.Data, &command)
		if err != nil {
			log.Error().Err(err).Msgf("parse homeassistant command %v", msg)
			return
		}

		ledState, err := l.hal.UpdateLeds(ctx, command.ledState())
		if err != nil {
			log.Error().Err(err).Msgf("homeassistant command %v", command)
			return
		}

		// Transform to HA state and publish.
		state := state{
			Brightness: uint(*ledState.Brightness),
			Effect:     *ledState.Effect,
			Color: color{
				R: uint8((*ledState.Color & 0xff)),
				G: uint8((*ledState.Color >> 8) & 0xff),
				B: uint8((*ledState.Color >> 16) & 0xff),
			},
		}
		if *ledState.State {
			state.State = "ON"
		} else {
			state.State = "OFF"
		}

		data, err := json.Marshal(&state)
		if err != nil {
			log.Error().Err(err).Msg("json marshal state")
			return
		}

		err = nc.Publish(l.stateTopic(), data)
		if err != nil {
			log.Error().Err(err).Msg("nats publish to " + l.stateTopic())
		}
	})
	if err != nil {
		return fmt.Errorf("nats subscribe to %s: %v", l.commandTopic(), err)
	}
	return nil
}

func (l *ledsDevice) announce(ctx context.Context, nc *nats.Conn) error {
	list, err := l.hal.GetLedEffects(ctx, &emptypb.Empty{})
	if err != nil {
		return err
	}

	cfg := ledsConfig{
		Name:                l.baseDevice.hostname,
		UniqueId:            l.baseDevice.uniqueId(),
		CommandTopic:        strings.ReplaceAll(l.commandTopic(), ".", "/"),
		StateTopic:          strings.ReplaceAll(l.stateTopic(), ".", "/"),
		Schema:              "json",
		Brightness:          true,
		Effect:              true,
		ColorMode:           true,
		SupportedColorModes: []string{"rgb"},
		EffectList:          list.Effects,
	}
	data, err := json.Marshal(&cfg)
	if err != nil {
		return err
	}

	return nc.Publish(l.announceTopic(), data)
}

func (l *ledsDevice) handleEvent(event *pb.Event, nc *nats.Conn) error {
	return nil
}
