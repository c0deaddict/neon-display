package display

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"

	pb "github.com/c0deaddict/neon-display/hal_proto"
)

type Config struct {
	HalSocketPath string
	WebBind       string
	WebPort       uint16
	PhotosPath    string
}

type Display struct {
	config  Config
	nc      *nats.Conn
	browser *os.Process

	mu      sync.Mutex // protects clients, also serves as WriteMessage sync.
	clients []*websocket.Conn
}

func New(config Config) Display {
	return Display{config: config}
}

func (d *Display) Run() {
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
		d.config.HalSocketPath,
		grpc.WithInsecure(),
		grpc.WithDialer(func(addr string, timeout time.Duration) (net.Conn, error) {
			return net.DialTimeout("unix", addr, timeout)
		}))
	if err != nil {
		log.Fatal().Err(err).Msg("grpc unix dial")
	}
	defer conn.Close()
	c := pb.NewHalClient(conn)

	// Contact the server.
	_, err = c.SetLedsPower(context.Background(), &pb.LedsPower{Power: true})
	if err != nil {
		log.Error().Err(err).Msg("set leds power")
	}

	_, err = c.SetDisplayPower(context.Background(), &pb.DisplayPower{Power: true})
	if err != nil {
		log.Error().Err(err).Msg("set display power")
	}

	url := fmt.Sprintf("http://localhost:%d", d.config.WebPort)
	p, err := startBrowser(url)
	if err != nil {
		log.Error().Err(err).Msg("start browser")
	} else {
		d.browser = p
	}

	stream, err := c.WatchEvents(context.Background(), &emptypb.Empty{})
	if err != nil {
		log.Error().Err(err).Msg("watch events")
	} else {
		for {
			event, err := stream.Recv()
			if err == io.EOF {
				log.Info().Msg("events stream eof")
				break
			} else if err != nil {
				log.Fatal().Err(err).Msg("watch events")
			}

			log.Info().Msgf("event: %s %v", event.Source, event.State)
		}
	}

	if d.browser != nil {
		d.browser.Kill()
		d.browser.Wait()
	}
}

func startBrowser(url string) (*os.Process, error) {
	cmd := exec.Command("firefox", "-kiosk", url)
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	return cmd.Process, nil
}
