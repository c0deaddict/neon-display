package leds

type LedEffect interface {
	Name() string
	Init(l *Leds) bool
	Update(l *Leds) bool
}

func effects() []LedEffect {
	return []LedEffect{
		&SolidLedEffect{},
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
