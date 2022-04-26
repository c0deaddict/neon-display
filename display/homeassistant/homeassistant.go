package homeassistant

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"

	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/types/known/emptypb"

	pb "github.com/c0deaddict/neon-display/hal_proto"
)

const (
	hostname = "neon"
	id       = "leds"
	uniqueId = hostname + "_" + id

	topic         = "homeassistant.light." + hostname + "." + id
	commandTopic  = topic + ".set"
	stateTopic    = topic + ".state"
	announceTopic = topic + ".config"

	stateOn  = "ON"
	stateOff = "OFF"
)

type config struct {
	Name                string   `json:"name"`
	UniqueId            string   `json:"unique_id"`
	CommandTopic        string   `json:"command_topic"`
	StateTopic          string   `json:"state_topic"`
	Schema              string   `json:"schema"`
	Brightness          bool     `json:"brightness"`
	Effect              bool     `json:"effect"`
	ColorMode           bool     `json:"color_mode"`
	SupportedColorModes []string `json:"supported_color_modes"`
	EffectList          []string `json:"effect_list"`
}

type color struct {
	R int `json:"r"`
	G int `json:"g"`
	B int `json:"b"`
}

type state struct {
	Brightness int    `json:"brightness"`
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

func Start(ctx context.Context, hal pb.HalClient, nc *nats.Conn) {
	_, err := nc.Subscribe("homeassistant.status", func(msg *nats.Msg) {
		if bytes.Compare(msg.Data, []byte("online")) == 0 {
			err := announce(ctx, hal, nc)
			if err != nil {
				log.Error().Err(err).Msg("homeassistant announce")
			}
		}
	})
	if err != nil {
		log.Error().Err(err).Msg("nats subscribe to homeassistant.status")
	}

	_, err = nc.Subscribe(commandTopic, func(msg *nats.Msg) {
		var command command
		err := json.Unmarshal(msg.Data, &command)
		if err != nil {
			log.Error().Err(err).Msgf("parse homeassistant command %v", msg)
			return
		}

		state, err := hal.UpdateLeds(ctx, command.ledState())
		if err != nil {
			log.Error().Err(err).Msgf("homeassistant command %v", command)
			return
		}

		// Publish state.
		data, err := json.Marshal(&state)
		if err != nil {
			log.Error().Err(err).Msg("json marshal state")
			return
		}

		err = nc.Publish(stateTopic, data)
		if err != nil {
			log.Error().Err(err).Msg("nats publish to " + stateTopic)
		}
	})
	if err != nil {
		log.Error().Err(err).Msg("nats subscribe to " + commandTopic)
	}

	err = announce(ctx, hal, nc)
	if err != nil {
		log.Error().Err(err).Msg("homeassistant announce")
	}
}

func announce(ctx context.Context, hal pb.HalClient, nc *nats.Conn) error {
	list, err := hal.GetLedEffects(ctx, &emptypb.Empty{})
	if err != nil {
		return err
	}

	cfg := config{
		Name:                hostname,
		UniqueId:            uniqueId,
		CommandTopic:        strings.ReplaceAll(commandTopic, ".", "/"),
		StateTopic:          strings.ReplaceAll(stateTopic, ".", "/"),
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

	return nc.Publish(announceTopic, data)
}

func (c color) uint32() uint32 {
	return uint32(c.R&0xff)<<16 | uint32(c.G&0xff)<<8 | uint32(c.B&0xff)
}

func (c command) ledState() *pb.LedState {
	s := pb.LedState{Effect: c.Effect}
	if c.State != nil {
		state := *c.State == stateOn
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
