package hal_proto

type Event struct {
	Source string
	State  bool
}

const (
	PirSource          = "pir"
	RedButtonSource    = "red"
	YellowButtonSource = "yellow"
)
