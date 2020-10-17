package allocator

import (
	"context"
	"fmt"
	"github.com/Octops/agones-discover-openmatch/internal/runtime"
	"github.com/sirupsen/logrus"
	"math/rand"
	"open-match.dev/open-match/pkg/pb"
)

var _ GameServerDiscoveryServiceClient = (*AgonesDiscoverClient)(nil)

type AgonesAllocatorService struct {
	logger *logrus.Entry
	GameServerAllocatorServiceClient
	GameServerDiscoveryServiceClient
}

func NewAgonesAllocatorService(gameServerAllocatorServiceClient GameServerAllocatorServiceClient, gameServerDiscoveryServiceClient GameServerDiscoveryServiceClient) *AgonesAllocatorService {
	return &AgonesAllocatorService{
		runtime.Logger().WithField("component", "agones_allocator"),
		gameServerAllocatorServiceClient,
		gameServerDiscoveryServiceClient,
	}
}

type FakeAllocatorServiceClient struct {
}

func (c *FakeAllocatorServiceClient) Allocate(ctx context.Context, req *pb.AssignTicketsRequest) error {
	for _, group := range req.Assignments {
		port := rand.Intn(8000-7000) + 7000
		conn := fmt.Sprintf("%d.%d.%d.%d:%d", rand.Intn(256), rand.Intn(256), rand.Intn(256), rand.Intn(256), port)
		group.Assignment.Connection = conn
		runtime.Logger().WithField("component", "allocator").Debugf("connection %s assigned to request", conn)
	}

	return nil
}

type AgonesDiscoverClient struct {
}

func (c *AgonesDiscoverClient) FindGameServer(ctx context.Context, filters map[interface{}]interface{}) ([]interface{}, error) {
	// TODO: Consume from Agones Discover API passing filter
	panic("implement me")
}
