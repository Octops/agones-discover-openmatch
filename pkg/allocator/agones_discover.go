package allocator

import (
	"context"
	"fmt"
	"github.com/Octops/agones-discover-openmatch/internal/runtime"
	"math/rand"
	"open-match.dev/open-match/pkg/pb"
)

var _ GameSessionAllocatorService = (*AgonesDiscoverAllocator)(nil)

type AgonesDiscoverAllocator struct {
}

// Agones Discover
func (c *AgonesDiscoverAllocator) Allocate(ctx context.Context, req *pb.AssignTicketsRequest) error {
	// Extract filters from Extensions field and query Agones Discover
	for _, group := range req.Assignments {
		port := rand.Intn(8000-7000) + 7000
		conn := fmt.Sprintf("%d.%d.%d.%d:%d", rand.Intn(256), rand.Intn(256), rand.Intn(256), rand.Intn(256), port)
		group.Assignment.Connection = conn
		runtime.Logger().WithField("component", "allocator").Debugf("extension %v", group.Assignment.Extensions)
		runtime.Logger().WithField("component", "allocator").Debugf("connection %s assigned to request", conn)
	}

	return nil
}

func (c *AgonesDiscoverAllocator) FindGameServer(ctx context.Context, filters map[interface{}]interface{}) ([]interface{}, error) {
	// TODO: Consume from Agones Discover API passing filter
	panic("implement me")
}
