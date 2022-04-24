package leds

import "time"

type SolidLedEffect struct{}

func (s *SolidLedEffect) Name() string {
	return "solid"
}

func (s *SolidLedEffect) Render(l *Leds) *time.Duration {
	l.fill(l.color)
	return nil
}
