package leds

type SolidLedEffect struct{}

func (s *SolidLedEffect) Name() string {
	return "solid"
}

func (s *SolidLedEffect) Init(l *Leds) bool {
	for i := 0; i < len(l.dev.Leds(0)); i++ {
		l.dev.Leds(0)[i] = l.color
	}
	return false
}

func (s *SolidLedEffect) Update(l *Leds) bool {
	return false
}
