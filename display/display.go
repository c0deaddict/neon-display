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

	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/c0deaddict/neon-display/display/homeassistant"
	"github.com/c0deaddict/neon-display/display/ws_proto"
	pb "github.com/c0deaddict/neon-display/hal_proto"
	"github.com/c0deaddict/neon-display/nats_helper"
)

type Config struct {
	HalSocketPath string `json:"hal_socket_path"`
	WebBind       string `json:"web_bind"`
	WebPort       uint16 `json:"web_port"`
	PhotosPath    string `json:"photos_path"`
	Sites         []Site `json:"sites"`
	InitTitle     string `json:"init_title"`
}

type Display struct {
	config         Config
	nc             *nats.Conn
	currentContent content

	mu      sync.Mutex // protects clients, also serves as WriteMessage sync.
	clients []client
}

func New(config Config) Display {
	return Display{config: config}
}

func (d *Display) Run(ctx context.Context) {
	err := d.initContent()
	if err != nil {
		log.Fatal().Err(err).Msg("init content")
	}

	// Connect to the HAL.
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
	hal := pb.NewHalClient(conn)

	// Connect to NATS.
	nc, err := nats_helper.Connect()
	if err != nil {
		log.Error().Err(err).Msg("connect to nats")
	} else {
		defer nc.Close()
		// Setup subscriptions for LEDs handling.
		homeassistant.Start(ctx, hal, nc)
	}

	// Start webserver.
	err = d.startWebserver()
	if err != nil {
		log.Fatal().Err(err).Msg("start webserver")
	}

	// Start browser process.
	url := fmt.Sprintf("http://localhost:%d", d.config.WebPort)
	p, err := startBrowser(url)
	if err != nil {
		log.Fatal().Err(err).Msg("start browser")
	}
	// Stop the browser at exit.
	defer p.Kill()
	defer p.Wait()

	// Display starts in off state.
	_, err = hal.SetDisplayPower(ctx, &pb.DisplayPower{Power: false})
	if err != nil {
		log.Error().Err(err).Msg("set display power off")
	}

	// Watch events from HAL and process them.
	stream, err := hal.WatchEvents(ctx, &emptypb.Empty{})
	if err != nil {
		log.Error().Err(err).Msg("watch events")
	} else {
		for {
			event, err := stream.Recv()
			if err == io.EOF {
				log.Info().Msg("events stream eof")
				break
			} else if err != nil {
				log.Error().Err(err).Msg("watch events")
				break
			}

			d.handleEvent(event)
		}
	}
}

func (d *Display) handleEvent(event *pb.Event) {
	log.Info().Msgf("event: %s %v", event.Source, event.State)

	switch event.Source {
	case pb.EventSource_Pir:
		if event.State {
			color := "red"
			d.showMessage(ws_proto.ShowMessage{
				Text:        "motion detected",
				Color:       &color,
				ShowSeconds: 5,
			})
			// stop off timer
		} else {
			color := "red"
			d.showMessage(ws_proto.ShowMessage{
				Text:        "no motion",
				Color:       &color,
				ShowSeconds: 5,
			})
			// reset off timer
		}

	case pb.EventSource_RedButton:
		if event.State {
			d.prevContent()
		}

	case pb.EventSource_YellowButton:
		if event.State {
			d.nextContent()
		}
	}
}

func startBrowser(url string) (*os.Process, error) {
	cmd := exec.Command("firefox", "-kiosk", url)
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	return cmd.Process, nil
}
