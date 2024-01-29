package nats_helper

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/rs/zerolog/log"
)

type Config struct {
	ServerUrl    string  `json:"server_url"`
	Username     *string `json:"username,omitempty"`
	PasswordFile *string `json:"password_file,omitempty"`
}

var (
	up = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "neon_display",
		Name:      "nats_up",
		Help:      "Status of the nats connection",
	})

	disconnects = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "neon_display",
		Name:      "nats_disconnects_total",
		Help:      "Total number of disconnects from a nats server",
	})
)

func Connect(clientName string, config *Config) (*nats.Conn, error) {
	var opts []nats.Option
	if config.Username != nil && *config.Username != "" {
		password, err := readPassword(config)
		if err != nil {
			return nil, err
		}
		opts = append(opts, nats.UserInfo(*config.Username, *password))
	}

	// Set the client name.
	opts = append(opts, nats.Name(clientName))

	// Try to reconnect every 2 seconds, forever.
	opts = append(opts, nats.MaxReconnects(-1))
	opts = append(opts, nats.ReconnectWait(2*time.Second))

	opts = append(opts, nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
		disconnects.Inc()
		log.Warn().Err(err).Msgf("Nats got disconnected from %v", nc)
	}))
	opts = append(opts, nats.ReconnectHandler(func(nc *nats.Conn) {
		up.Set(1)
		log.Info().Msgf("Nats got reconnected to %v", nc.ConnectedUrl())
	}))
	opts = append(opts, nats.ClosedHandler(func(nc *nats.Conn) {
		up.Set(0)
		log.Warn().Err(nc.LastError()).Msg("Nats connection closed")
	}))

	nc, err := nats.Connect(config.ServerUrl, opts...)
	if err != nil {
		return nil, err
	}

	log.Info().Msgf("Connected to nats: %v", nc.ConnectedUrl())

	up.Set(1)

	return nc, nil
}

func readPassword(config *Config) (*string, error) {
	if config.PasswordFile == nil {
		return nil, fmt.Errorf("nats password required but not defined")
	}

	info, err := os.Stat(*config.PasswordFile)
	if err != nil {
		return nil, err
	}

	if info.Mode()&0o077 != 0 {
		log.Warn().Msgf("Warning: permissions are too open on %v", *config.PasswordFile)
	}

	if password, err := readFirstLine(*config.PasswordFile); err != nil {
		return nil, err
	} else {
		return password, nil
	}
}

func readFirstLine(path string) (*string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Scan()
	line := strings.TrimSpace(scanner.Text())
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return &line, nil
}
