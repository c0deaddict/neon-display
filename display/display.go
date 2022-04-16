package display

import (
	"os"
	"os/exec"

	"github.com/nats-io/nats.go"
)

type Display struct {
	HalSocketPath string

	nc      *nats.Conn
	browser *os.Process
}

func (d *Display) Run() {
	// Connect to HAL unix socket
	// Connect to NATS
	// Start firefox
	// Initialize slideshow

	// nc, err := nats_helper.Connect()
	// if err != nil {
	// 	log.Error().Err(err).Msg("failed to connect to nats")
	// }
	// defer nc.Close()
	d.StartWebsocket()
}

func startBrowser(url string) (*os.Process, error) {
	cmd := exec.Command("firefox", "-kiosk", url)
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	return cmd.Process, nil
}
