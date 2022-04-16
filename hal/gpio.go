package hal

import (
	"time"

	"github.com/rs/zerolog/log"
	"github.com/warthog618/gpiod"
	"github.com/warthog618/gpiod/device/rpi"

	pb "github.com/c0deaddict/neon-display/hal_proto"
)

const (
	gpioDevice = "gpiochip0"

	// Pin numbering: https://elinux.org/RPi_BCM2835_GPIOs
	pirPin          = rpi.J8p11 // GPIO 17
	redButtonPin    = rpi.J8p15 // GPIO 22
	yellowButtonPin = rpi.J8p13 // GPIO 27
	debounceDelay   = time.Duration(50 * time.Millisecond)
)

type gpioWatcher struct {
	lines []*gpiod.Line
}

func watchPir(h *Hal) (*gpiod.Line, error) {
	line, err := gpiod.RequestLine(gpioDevice, pirPin,
		gpiod.WithBothEdges,
		gpiod.WithEventHandler(func(evt gpiod.LineEvent) {
			h.publishEvent(&pb.Event{
				Source: pb.EventSource_Pir,
				State:  evt.Type == gpiod.LineEventRisingEdge,
			})
		}))
	return line, err
}

func watchButton(h *Hal, pin int, source pb.EventSource) (*gpiod.Line, error) {
	line, err := gpiod.RequestLine(gpioDevice, pin,
		gpiod.WithPullUp,
		gpiod.WithBothEdges,
		gpiod.WithDebounce(debounceDelay),
		gpiod.WithEventHandler(func(evt gpiod.LineEvent) {
			h.publishEvent(&pb.Event{
				Source: source,
				State:  evt.Type == gpiod.LineEventRisingEdge,
			})
		}))
	return line, err
}

func watchGpios(h *Hal) *gpioWatcher {
	g := &gpioWatcher{}

	l, err := watchPir(h)
	if err != nil {
		log.Error().Err(err).Msg("watch pir")
	} else {
		g.lines = append(g.lines, l)
	}

	l, err = watchButton(h, redButtonPin, pb.EventSource_RedButton)
	if err != nil {
		log.Error().Err(err).Msg("watch red button")
	} else {
		g.lines = append(g.lines, l)
	}

	l, err = watchButton(h, yellowButtonPin, pb.EventSource_YellowButton)
	if err != nil {
		log.Error().Err(err).Msg("watch yellow button")
	} else {
		g.lines = append(g.lines, l)
	}

	return g
}

func (g *gpioWatcher) Close() {
	for _, l := range g.lines {
		if err := l.Close(); err != nil {
			log.Error().Err(err).Msg("close gpio line")
		}
	}
}
