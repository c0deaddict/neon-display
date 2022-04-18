package hal

import (
	"time"

	ws2811 "github.com/rpi-ws281x/rpi-ws281x-go"
)

const (
	brightness = 128
	ledCounts  = 40
	gpioPin    = 12
	freq       = 800000
	sleepTime  = 50
)

type wsEngine interface {
	Init() error
	Render() error
	Wait() error
	Fini()
	Leds(channel int) []uint32
}

func checkError(err error) {
	if err != nil {
		panic(err)
	}
}

type colorWipe struct {
	ws wsEngine
}

func (cw *colorWipe) setup() error {
	return cw.ws.Init()
}

func (cw *colorWipe) display(color uint32) error {
	for i := 0; i < len(cw.ws.Leds(0)); i++ {
		cw.ws.Leds(0)[i] = color
		if err := cw.ws.Render(); err != nil {
			return err
		}
		time.Sleep(sleepTime * time.Millisecond)
	}
	return nil
}

func doLeds() {
	opt := ws2811.DefaultOptions
	opt.Channels[0].Brightness = brightness
	opt.Channels[0].LedCount = ledCounts
	opt.Channels[0].GpioPin = gpioPin
	opt.Frequency = freq

	dev, err := ws2811.MakeWS2811(&opt)
	checkError(err)

	cw := &colorWipe{
		ws: dev,
	}
	checkError(cw.setup())
	defer dev.Fini()

	cw.display(uint32(0x0000ff))
	cw.display(uint32(0x00ff00))
	cw.display(uint32(0xff0000))
	cw.display(uint32(0x000000))
}
