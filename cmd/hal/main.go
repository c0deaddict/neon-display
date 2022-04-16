package main

import (
	"os/signal"
	"syscall"
	"time"

	"flag"
	"os"

	"github.com/peterbourgon/ff"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/c0deaddict/neon-display/hal"
)

var (
	socketPath     = flag.String("hal-socket-path", "/run/neon-display/hal.sock", "HAL unix domain socket path")
	exporterListen = flag.String("exporter-listen", ":9989", "Prometheus exporter listen address")
)

func main() {
	ff.Parse(flag.CommandLine, os.Args[1:],
		ff.WithEnvVarPrefix("NEON"),
		ff.WithConfigFileFlag("config"),
		ff.WithConfigFileParser(ff.PlainParser))

	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)

	h := hal.Hal{SocketPath: *socketPath, ExporterListen: *exporterListen}
	go func() {
		err := h.Run()
		if err != nil {
			log.Error().Err(err).Msg("hal")
		}
	}()

	<-sigs
	h.Stop()
}
