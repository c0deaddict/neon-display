package leds

type color uint32

func rgbColor(r byte, g byte, b byte) color {
	return color(uint32(r) | uint32(g)<<8 | uint32(b)<<16)
}

func (c color) multiply(f float64) color {
	r := byte(f * float64(c&0xff))
	g := byte(f * float64((c>>8)&0xff))
	b := byte(f * float64((c>>16)&0xff))
	return rgbColor(r, g, b)
}
