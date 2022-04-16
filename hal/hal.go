package hal

import (
	"context"
	"net"
	"net/http"
	"os"
	"sync"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
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

	h.server = grpc.NewServer()
	pb.RegisterHalServer(h.server, h)
	err = h.server.Serve(lis)
	log.Info().Err(err).Msg("server stopped")
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
	{
		h.rw.Lock()
		defer h.rw.Unlock()
		h.watchers = append(h.watchers, stream)
	}

	<-stream.Context().Done()
	log.Info().Msg("client has disconnected WatchEvents stream")
	return nil
}

func (h *Hal) SetDisplayPower(ctx context.Context, power *pb.DisplayPower) (*emptypb.Empty, error) {
	return nil, nil
}

func (h *Hal) SetLedsPower(ctx context.Context, power *pb.LedsPower) (*emptypb.Empty, error) {
	log.Info().Msgf("Request to set leds power to: %v", power.Power)
	return nil, nil
}

func (h *Hal) publishEvent(event *pb.Event) {
	h.rw.RLock()
	defer h.rw.RUnlock()

	for _, w := range h.watchers {
		w.Send(event)
	}
}
