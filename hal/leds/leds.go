package leds

import (
	"sync"
	"time"

	ws2811 "github.com/rpi-ws281x/rpi-ws281x-go"
	"github.com/rs/zerolog/log"

	pb "github.com/c0deaddict/neon-display/hal_proto"
)

const (
	ledCount  = 47
	gpioPin   = 12
	freq      = 800000
	sleepTime = 50

	defaultBrightness = 128
	defaultColor      = 0xff0000
	defaultEffect     = "solid"
)

type Leds struct {
	dev    *ws2811.WS2811
	update chan bool
	stop   chan bool

	mu         sync.Mutex
	state      bool
	brightness int
	color      uint32
	effect     LedEffect
	timer      *time.Timer
}

func Start() (*Leds, error) {
	opt := ws2811.DefaultOptions
	opt.Channels[0].Brightness = defaultBrightness
	opt.Channels[0].LedCount = ledCount
	opt.Channels[0].GpioPin = gpioPin
	opt.Frequency = freq

	dev, err := ws2811.MakeWS2811(&opt)
	if err != nil {
		return nil, err
	}

	err = dev.Init()
	if err != nil {
		return nil, err
	}

	l := Leds{
		dev:        dev,
		stop:       make(chan bool),
		update:     make(chan bool),
		state:      false,
		brightness: defaultBrightness,
		color:      defaultColor,
		effect:     getEffect(defaultEffect),
		timer:      nil,
	}
	go l.loop()

	return &l, nil
}

func (l *Leds) Stop() {
	l.stop <- true
	l.dev.Fini()
}

func (l *Leds) Update(s *pb.LedState) *pb.LedState {
	l.mu.Lock()
	defer l.mu.Unlock()

	if s.State != nil {
		l.state = *s.State
	}

	if s.Brightness != nil {
		l.brightness = int(*s.Brightness)
		l.dev.SetBrightness(0, l.brightness)
	}

	if s.Color != nil {
		l.color = *s.Color
	}

	if s.Effect != nil {
		l.effect = getEffect(*s.Effect)
	}

	brightness := uint32(l.brightness)
	result := pb.LedState{
		State:      &l.state,
		Brightness: &brightness,
		Color:      &l.color,
	}
	if l.effect != nil {
		name := l.effect.Name()
		result.Effect = &name
	}

	// Notify the loop to trigger a re-render.
	l.update <- true

	return &result
}

/// Requires l.mu locked.
func (l *Leds) fill(color uint32) {
	for i := 0; i < len(l.dev.Leds(0)); i++ {
		l.dev.Leds(0)[i] = 0
	}
}

/// Requires l.mu locked.
func (l *Leds) off() {
	l.fill(0)
	err := l.dev.Render()
	if err != nil {
		log.Error().Err(err).Msg("leds render")
	}
}

func (l *Leds) loop() {
	var timer <-chan time.Time

	for {
		select {
		case <-l.stop:
			return
		case <-l.update:
			timer = optionalTimeAfter(l.render())
		case <-timer:
			timer = optionalTimeAfter(l.render())
		}
	}
}

func optionalTimeAfter(wait *time.Duration) <-chan time.Time {
	if wait == nil {
		return nil
	} else {
		return time.After(*wait)
	}
}

func (l *Leds) render() *time.Duration {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.state == false {
		l.off()
		return nil
	} else if l.effect != nil {
		start := time.Now()
		wait := l.effect.Render(l)
		err := l.dev.Render()
		if err != nil {
			log.Error().Err(err).Msg("leds render")
		}
		elapsed := time.Since(start)
		log.Debug().Dur("time", elapsed).Msg("leds render")
		if wait != nil {
			*wait -= elapsed
			if *wait < 0 {
				*wait = 0
			}
		}
		return wait
	}

	return nil
}
