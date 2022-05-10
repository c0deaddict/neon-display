package leds

import "time"

type LedEffect interface {
	Name() string
	Render(l *Leds) *time.Duration
}

func effects() []LedEffect {
	return []LedEffect{
		&SolidLedEffect{},
		&BreatheLedEffect{
			wait:  fpsWait(60.0),
			cycle: 60.0 * 5.0, // repeat every 5 seconds.
		},
	}
}

func Effects() []string {
	list := effects()
	result := make([]string, len(list))
	for i := 0; i < len(list); i++ {
		result[i] = list[i].Name()
	}
	return result
}

func getEffect(name string) LedEffect {
	list := effects()
	for i := 0; i < len(list); i++ {
		if list[i].Name() == name {
			return list[i]
		}
	}
	return nil
}

func fpsWait(fps float64) time.Duration {
	return time.Duration(1000000.0/fps) * time.Microsecond
}
