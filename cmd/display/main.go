package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os/signal"
	"syscall"
	"time"

	"flag"
	"os"

	"github.com/peterbourgon/ff"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/c0deaddict/neon-display/display"
)

var (
	configFile = flag.String("config", "", "Config file")
)

func main() {
	ff.Parse(flag.CommandLine, os.Args[1:], ff.WithEnvVarPrefix("NEON"))

	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})

	config, err := loadConfig(*configFile)
	if err != nil {
		log.Fatal().Err(err).Msg("load config")
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)

	d := display.New(config)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan bool)
	go func() {
		d.Run(ctx)
		done <- true
	}()

	select {
	case <-sigs:
		cancel()
		<-done
	case <-done:
	}
}

func loadConfig(configFile string) (display.Config, error) {
	config := display.Config{
		HalSocketPath: "/run/neon-display/hal.sock",
		WebBind:       "127.0.0.1",
		WebPort:       8080,
	}

	contents, err := ioutil.ReadFile(configFile)
	if err != nil {
		return config, fmt.Errorf("read config file: %v", err)
	}

	err = json.Unmarshal(contents, &config)
	if err != nil {
		return config, fmt.Errorf("parse config file: %v", err)
	}

	return config, nil
}
