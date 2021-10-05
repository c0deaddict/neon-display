package nats_helper

import (
	"bufio"
	"flag"
	"log"
	"os"
	"strings"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	serverUrl    = flag.String("nats-url", "nats://nats:4222", "Nats URL")
	username     = flag.String("nats-username", "", "Nats username")
	password     = flag.String("nats-password", "", "Nats password")
	passwordFile = flag.String("nats-password-file", "", "Read nats password from this file")

	up = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "smart_meter",
		Name:      "nats_up",
		Help:      "Status of the nats connection",
	})

	disconnects = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "smart_meter",
		Name:      "nats_disconnects_total",
		Help:      "Total number of disconnects from a nats server",
	})
)

func Connect() (*nats.Conn, error) {
	var opts []nats.Option
	if *username != "" {
		password, err := readPassword()
		if err != nil {
			return nil, err
		}
		opts = append(opts, nats.UserInfo(*username, *password))
	}

	// Try to reconnect every 2 seconds, forever.
	opts = append(opts, nats.MaxReconnects(-1))
	opts = append(opts, nats.ReconnectWait(2*time.Second))

	opts = append(opts, nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
		disconnects.Inc()
		log.Printf("Nats got disconnected from %v. Reason: %q\n", nc.ConnectedUrl(), err)
	}))
	opts = append(opts, nats.ReconnectHandler(func(nc *nats.Conn) {
		up.Set(1)
		log.Printf("Nats got reconnected to %v\n", nc.ConnectedUrl())
	}))
	opts = append(opts, nats.ClosedHandler(func(nc *nats.Conn) {
		up.Set(0)
		log.Printf("Nats connection closed. Reason: %q\n", nc.LastError())
	}))

	nc, err := nats.Connect(*serverUrl, opts...)
	if err != nil {
		return nil, err
	}

	up.Set(1)

	return nc, nil
}

func readPassword() (*string, error) {
	if *passwordFile != "" {
		info, err := os.Stat(*passwordFile)
		if err != nil {
			return nil, err
		}

		if info.Mode()&0o077 != 0 {
			log.Printf("Warning: permissions are too open on %v\n", *passwordFile)
		}

		if password, err := readFirstLine(*passwordFile); err != nil {
			return nil, err
		} else {
			return password, nil
		}
	}

	return password, nil
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
