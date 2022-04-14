package hal_proto

type Command string

const (
	CommandStartLeds Command = "leds:start"
	CommandStopLeds  Command = "leds:stop"
)
