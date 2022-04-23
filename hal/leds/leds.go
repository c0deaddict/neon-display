package leds

import (
	"sync"

	ws2811 "github.com/rpi-ws281x/rpi-ws281x-go"

	pb "github.com/c0deaddict/neon-display/hal_proto"
)

const (
	brightness = 128
	ledCount   = 47
	gpioPin    = 12
	freq       = 800000
	sleepTime  = 50
)

type Leds struct {
	dev *ws2811.WS2811

	mu         sync.Mutex
	state      bool
	brightness int
	color      uint32
	effect     LedEffect
}

func Start() (*Leds, error) {
	opt := ws2811.DefaultOptions
	opt.Channels[0].Brightness = brightness
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

	l := Leds{dev: dev}
	go l.loop()

	return &l, nil
}

func (l *Leds) Stop() {
	l.dev.Fini()
}

func (l *Leds) Update(s *pb.LedState) *pb.LedState {
	l.mu.Lock()
	defer l.mu.Unlock()

	if s.State != nil {
		l.state = *s.State
		// TODO if state changes, turn off or on (init effect).
	}

	if s.Brightness != nil {
		l.brightness = int(*s.Brightness)
		l.dev.SetBrightness(0, l.brightness)
	}

	if s.Color != nil {
		l.color = *s.Color
	}

	if s.Effect != nil {
		// TODO: if effect changes, call init on it.
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
	return &result
}

func (l *Leds) loop() {
	l.render()
}

func (l *Leds) render() {
	l.mu.Lock()
	defer l.mu.Unlock()

}
