package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
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
	webBind       = flag.String("web-bind", "127.0.0.1", "Web bind")
	webPort       = flag.Uint("web-port", 8080, "Web port")
	photosPath    = flag.String("photos-path", "/var/lib/neon-display/photos", "Photos used for slideshow")
	sitesFile     = flag.String("sites-file", "", "Sites file in JSON format")
	initTitle     = flag.String("init-title", "", "Content title to start with")
)

func main() {
	ff.Parse(flag.CommandLine, os.Args[1:],
		ff.WithEnvVarPrefix("NEON"),
		ff.WithConfigFileFlag("config"),
		ff.WithConfigFileParser(ff.PlainParser))

	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})

	sites := make([]display.Site, 0)
	if *sitesFile != "" {
		var err error
		sites, err = parseSites(*sitesFile)
		if err != nil {
			log.Fatal().Err(err).Msg("parse sites")
		}
	}

	d := display.New(display.Config{
		HalSocketPath: *halSocketPath,
		WebBind:       *webBind,
		WebPort:       uint16(*webPort),
		PhotosPath:    *photosPath,
		Sites:         sites,
		InitTitle:     *initTitle,
	})
	d.Run()
}

func parseSites(sitesFile string) ([]display.Site, error) {
	contents, err := ioutil.ReadFile(sitesFile)
	if err != nil {
		return nil, fmt.Errorf("read sites file: %v", err)
	}

	var sites []display.Site
	err = json.Unmarshal(contents, &sites)
	if err != nil {
		return nil, fmt.Errorf("parse sites file: %v", err)
	}

	return sites, nil
}
