package leds

import (
	"math"
	"time"
)

type BreatheLedEffect struct {
	wait  time.Duration
	tick  uint64
	cycle float64
}

func (b *BreatheLedEffect) Name() string {
	return "breathe"
}

func (b *BreatheLedEffect) Render(l *Leds) *time.Duration {
	x := (float64(b.tick) * math.Pi * 2.0) / b.cycle // sinus repeats every b.cycle ticks
	y := (1.0 + math.Sin(x)) / 4.0                   // y in range [0.0, 0.5]
	c := l.color.multiply(0.5 + y)
	l.fill(c)
	b.tick += 1
	return &b.wait
}
