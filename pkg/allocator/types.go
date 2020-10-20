package allocator

import (
	"context"
	"github.com/Octops/agones-discover-openmatch/internal/runtime"
	"github.com/sirupsen/logrus"
	"open-match.dev/open-match/pkg/pb"
)

type AllocatorService struct {
	logger *logrus.Entry
	GameServerAllocator
}

type GameServerAllocator interface {
	Allocate(ctx context.Context, req *pb.AssignTicketsRequest) error
}

// Allocate Dedicated GameServers
type GameServerAllocatorService interface {
	GameServerAllocatorClient
}

// Allocate GameServers by Sessions
type GameSessionAllocatorService interface {
	GameServerAllocatorClient
	GameServerDiscoveryClient
}

// GameServerAllocatorClient allocates GameServers without knowing them beforehand
// Usually that could be done by pushing a GameServerAllocation request or talking directly to
// the GameServer Allocator Service from Agones
type GameServerAllocatorClient interface {
	Allocate(ctx context.Context, req *pb.AssignTicketsRequest) error
}

// GameServerDiscoveryClient communicates with some sort of underlying infrastructure or service
// and return GameServers for a given filter
type GameServerDiscoveryClient interface {
	FindGameServer(ctx context.Context, filters map[string]string) ([]byte, error)
}

func NewAllocatorService(service GameServerAllocator) AllocatorService {
	return AllocatorService{
		runtime.Logger().WithField("component", "agones_allocator"),
		service,
	}
}

func (s *AllocatorService) Allocate(ctx context.Context, req *pb.AssignTicketsRequest) error {
	return s.GameServerAllocator.Allocate(ctx, req)
}
