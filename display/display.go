package display

import (
	"context"
	"log"
	"net"
	"os"
	"os/exec"
	"time"

	"github.com/nats-io/nats.go"
	"google.golang.org/grpc"

	pb "github.com/c0deaddict/neon-display/hal_proto"
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
	go d.StartWebsocket()

	conn, err := grpc.Dial(
		d.HalSocketPath,
		grpc.WithInsecure(),
		grpc.WithDialer(func(addr string, timeout time.Duration) (net.Conn, error) {
			return net.DialTimeout("unix", addr, timeout)
		}))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewHalClient(conn)

	// Contact the server.
	c.SetLedsPower(context.Background(), &pb.LedsPower{Power: true})
	if err != nil {
		log.Fatalf("could not set leds power: %v", err)
	}
}

func startBrowser(url string) (*os.Process, error) {
	cmd := exec.Command("firefox", "-kiosk", url)
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	return cmd.Process, nil
}
