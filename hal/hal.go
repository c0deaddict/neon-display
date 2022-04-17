package hal

import (
	"context"
	"net"
	"net/http"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	pb "github.com/c0deaddict/neon-display/hal_proto"
)

func Run() {

}

type Hal struct {
	pb.UnimplementedHalServer

	SocketPath     string
	ExporterListen string

	server   *grpc.Server
	rw       sync.RWMutex // protects watchers
	watchers []pb.Hal_WatchEventsServer
}

func (h *Hal) Run() error {
	addr, err := net.ResolveUnixAddr("unix", h.SocketPath)
	if err != nil {
		return err
	}

	lis, err := net.ListenUnix("unix", addr)
	if err != nil {
		return err
	}

	defer func() {
		if err := os.RemoveAll(h.SocketPath); err != nil {
			log.Error().Err(err).Msg("remove hal socket path")
		}
	}()

	w := watchGpios(h)
	defer w.Close()

	done := make(chan bool)
	go func() {
		state := false
		ticker := time.NewTicker(2 * time.Second)
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				state = !state
				h.publishEvent(&pb.Event{
					Source: pb.EventSource_Pir,
					State:  state,
				})
			}
		}
	}()

	h.server = grpc.NewServer()
	pb.RegisterHalServer(h.server, h)
	err = h.server.Serve(lis)
	log.Info().Err(err).Msg("server stopped")
	done <- true
	return err
}

func (h *Hal) Stop() {
	if h.server != nil {
		h.server.Stop()
		h.server = nil
	}
}

func (h *Hal) startMetrics() {
	http.Handle("/metrics", promhttp.Handler())
	log.Fatal().Err(http.ListenAndServe(h.ExporterListen, nil))
}

func (h *Hal) WatchEvents(_ *emptypb.Empty, stream pb.Hal_WatchEventsServer) error {
	h.addWatcher(stream)
	<-stream.Context().Done()
	h.removeWatcher(stream)
	return nil
}

func (h *Hal) addWatcher(stream pb.Hal_WatchEventsServer) {
	h.rw.Lock()
	defer h.rw.Unlock()
	h.watchers = append(h.watchers, stream)
	log.Info().Msgf("added watcher %v", stream)
}

func (h *Hal) removeWatcher(stream pb.Hal_WatchEventsServer) {
	h.rw.Lock()
	defer h.rw.Unlock()

	for i, other := range h.watchers {
		if other == stream {
			h.watchers = append(h.watchers[:i], h.watchers[i+1:]...)
			log.Info().Msgf("removed watcher %v", stream)
			return
		}
	}

	log.Warn().Msgf("remove watcher: %v is not found", stream)
}

func (h *Hal) SetDisplayPower(ctx context.Context, power *pb.DisplayPower) (*emptypb.Empty, error) {
	state := "0"
	if power.Power {
		state = "1"
	}

	err := exec.Command("vcgencmd", "display_power", state).Run()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &emptypb.Empty{}, nil
}

func (h *Hal) SetLedsPower(ctx context.Context, power *pb.LedsPower) (*emptypb.Empty, error) {
	log.Info().Msgf("Request to set leds power to: %v", power.Power)
	return &emptypb.Empty{}, nil
}

func (h *Hal) publishEvent(event *pb.Event) {
	h.rw.RLock()
	defer h.rw.RUnlock()

	log.Info().Msgf("publish event %v to watchers %d", event, len(h.watchers))

	for _, w := range h.watchers {
		w.Send(event)
	}
}