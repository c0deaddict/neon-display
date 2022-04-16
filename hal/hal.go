package hal

import (
	"context"
	"net"
	"sync"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"

	pb "github.com/c0deaddict/neon-display/hal_proto"
)

type Hal struct {
	pb.UnimplementedHalServer
	socketFile string
	server     *grpc.Server
	rw         sync.RWMutex // protects watchers
	watchers   []pb.Hal_WatchEventsServer
}

func New(socketFile string) Hal {
	return Hal{socketFile: socketFile}
}

func (h *Hal) Run() error {
	addr, err := net.ResolveUnixAddr("unix", h.socketFile)
	if err != nil {
		return err
	}

	lis, err := net.ListenUnix("unix", addr)
	if err != nil {
		return err
	}

	w := watchGpios(h)
	defer w.Close()

	h.server = grpc.NewServer()
	pb.RegisterHalServer(h.server, h)
	return h.server.Serve(lis)
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
	return nil, nil
}

func (h *Hal) publishEvent(event *pb.Event) {
	h.rw.RLock()
	defer h.rw.RUnlock()

	for _, w := range h.watchers {
		w.Send(event)
	}
}
