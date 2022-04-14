package hal

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog/log"
	"github.com/warthog618/gpiod"
	"github.com/warthog618/gpiod/device/rpi"

	"github.com/c0deaddict/neon-display/internal/hal_proto"
	"github.com/c0deaddict/neon-display/internal/nats_helper"
)

const (
	gpioDevice = "gpiochip0"

	// Pin numbering: https://elinux.org/RPi_BCM2835_GPIOs
	pirPin          = rpi.J8p11 // GPIO 17
	redButtonPin    = rpi.J8p15 // GPIO 22
	yellowButtonPin = rpi.J8p13 // GPIO 27
	debounceDelay   = time.Duration(50 * time.Millisecond)
)

type Hal struct {
	nc *nats.Conn
}

func Run() error {
	nc, err := nats_helper.Connect()
	if err != nil {
		return fmt.Errorf("failed to connect to nats: %v", err)
	}

	h := Hal{nc}

	l, err := h.watchPir()
	if err != nil {
		return err
	}
	defer l.Close()

	l, err = h.watchButton(redButtonPin, hal_proto.RedButtonSource)
	if err != nil {
		return err
	}
	defer l.Close()

	l, err = h.watchButton(yellowButtonPin, hal_proto.YellowButtonSource)
	if err != nil {
		return err
	}
	defer l.Close()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh

	return nil
}

func (h *Hal) publish(event hal_proto.Event) {
	payload, err := json.Marshal(event)
	if err != nil {
		log.Error().Err(err).Msg("failed to marshal event")
		return
	}
	if err := h.nc.Publish("hal.events", []byte(payload)); err != nil {
		log.Warn().Err(err).Msg("failed to publish event")
	}
}

func (h *Hal) watchPir() (*gpiod.Line, error) {
	l, err := gpiod.RequestLine(gpioDevice, pirPin,
		gpiod.WithBothEdges,
		gpiod.WithEventHandler(func(evt gpiod.LineEvent) {
			h.publish(hal_proto.Event{
				Source: hal_proto.PirSource,
				State:  evt.Type == gpiod.LineEventRisingEdge,
			})
		}))

	if err != nil {
		return nil, fmt.Errorf("failed to open GPIO for PIR: %v", err)

	} else {
		log.Info().Msgf("watching pir pin %d", pirPin)
		return l, nil
	}
}

func (h *Hal) watchButton(pin int, source string) (*gpiod.Line, error) {
	l, err := gpiod.RequestLine(gpioDevice, pin,
		gpiod.WithPullUp,
		gpiod.WithBothEdges,
		gpiod.WithDebounce(debounceDelay),
		gpiod.WithEventHandler(func(evt gpiod.LineEvent) {
			h.publish(hal_proto.Event{
				Source: source,
				State:  evt.Type == gpiod.LineEventRisingEdge,
			})
		}))

	if err != nil {
		return nil, fmt.Errorf("failed to open GPIO for %s: %v", source, err)

	} else {
		log.Info().Msgf("watching %s pin %d", source, pin)
		return l, nil
	}
}
