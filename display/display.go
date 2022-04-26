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
	"github.com/c0deaddict/neon-display/display/nats_helper"
	"github.com/c0deaddict/neon-display/display/ws_proto"
	pb "github.com/c0deaddict/neon-display/hal_proto"
)

type Config struct {
	HalSocketPath string             `json:"hal_socket_path"`
	WebBind       string             `json:"web_bind"`
	WebPort       uint16             `json:"web_port"`
	PhotosPath    string             `json:"photos_path,omitempty"`
	VideosPath    string             `json:"videos_path,omitempty"`
	FirefoxBin    string             `json:"firefox_bin,omitempty"`
	Sites         []Site             `json:"sites"`
	InitTitle     string             `json:"init_title"`
	Nats          nats_helper.Config `json:"nats"`
	OffTimeout    uint               `json:"off_timeout"`
	// TODO: add power off hours (schedule)
}

type Display struct {
	config         Config
	nc             *nats.Conn
	currentContent content
	hal            pb.HalClient
	offTimer       *time.Timer

	mu      sync.Mutex // protects clients, power and also serves as WriteMessage sync.
	clients []client
	power   bool
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
	d.hal = pb.NewHalClient(conn)

	// Connect to NATS.
	nc, err := nats_helper.Connect(&d.config.Nats)
	if err != nil {
		log.Error().Err(err).Msg("connect to nats")
	} else {
		defer nc.Close()
		// Setup subscriptions for LEDs handling.
		homeassistant.Start(ctx, d.hal, nc)
	}

	// Start webserver.
	err = d.startWebserver()
	if err != nil {
		log.Fatal().Err(err).Msg("start webserver")
	}

	// Start browser process.
	url := fmt.Sprintf("http://localhost:%d", d.config.WebPort)
	p, err := d.startBrowser(url)
	if err != nil {
		log.Fatal().Err(err).Msg("start browser")
	}
	go func() {
		state, err := p.Wait()
		if err != nil {
			log.Fatal().Err(err).Msg("browser process wait")
		}
		log.Fatal().Msgf("browser exitted with %v", state.ExitCode())
	}()
	// Stop the browser at exit.
	defer p.Kill()

	// Turn off display after config.OffTimeout seconds.
	d.startOffTimer()

	// Watch events from HAL and process them.
	stream, err := d.hal.WatchEvents(ctx, &emptypb.Empty{})
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

func (d *Display) startOffTimer() {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.offTimer != nil {
		d.offTimer.Stop()
	}

	wait := time.Duration(d.config.OffTimeout) * time.Second
	d.offTimer = time.AfterFunc(wait, d.powerOff)
}

func (d *Display) powerOff() {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.power {
		log.Info().Msg("no motion: turning off display and pausing content")
		_, err := d.hal.SetDisplayPower(context.Background(), &pb.DisplayPower{Power: false})
		if err != nil {
			log.Error().Err(err).Msg("set display power off")
		} else {
			d.power = false
			d.pauseContent()
		}
	}
}

func (d *Display) powerOn() {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.offTimer != nil {
		d.offTimer.Stop()
		d.offTimer = nil
	}

	if !d.power {
		log.Info().Msg("motion detected: turning on display and resuming content")
		_, err := d.hal.SetDisplayPower(context.Background(), &pb.DisplayPower{Power: true})
		if err != nil {
			log.Error().Err(err).Msg("set display power on")
		} else {
			d.power = true
			d.resumeContent()
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
			d.powerOn()
		} else {
			color := "red"
			d.showMessage(ws_proto.ShowMessage{
				Text:        "no motion",
				Color:       &color,
				ShowSeconds: 5,
			})
			d.startOffTimer()
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

func (d *Display) startBrowser(url string) (*os.Process, error) {
	cmd := exec.Command(d.config.FirefoxBin, "-kiosk", url)
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	return cmd.Process, nil
}
