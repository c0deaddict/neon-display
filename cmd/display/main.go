package main

import (
	"time"

	"flag"
	"os"

	"github.com/peterbourgon/ff"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/c0deaddict/neon-display/display"
)

var (
	halSocketPath = flag.String("hal-socket-path", "/run/neon-display/hal.sock", "HAL unix domain socket path")
)

func main() {
	ff.Parse(flag.CommandLine, os.Args[1:],
		ff.WithEnvVarPrefix("NEON"),
		ff.WithConfigFileFlag("config"),
		ff.WithConfigFileParser(ff.PlainParser))

	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})

	d := display.Display{
		HalSocketPath: *halSocketPath,
	}
	d.Run()
}
