package allocator

import (
	"context"
	"fmt"
	"github.com/Octops/agones-discover-openmatch/internal/runtime"
	"math/rand"
	"open-match.dev/open-match/pkg/pb"
)

var _ GameSessionAllocatorService = (*AgonesDiscoverAllocator)(nil)

type AgonesDiscoverClient interface {
	ListGameServers(ctx context.Context, filter map[string]string) ([]byte, error)
}

type AgonesDiscoverAllocator struct {
	client AgonesDiscoverClient
}

// Agones Discover
func (c *AgonesDiscoverAllocator) Allocate(ctx context.Context, req *pb.AssignTicketsRequest) error {
	// Extract filters from Extensions field and query Agones Discover
	for _, group := range req.Assignments {
		port := rand.Intn(8000-7000) + 7000
		conn := fmt.Sprintf("%d.%d.%d.%d:%d", rand.Intn(256), rand.Intn(256), rand.Intn(256), rand.Intn(256), port)
		group.Assignment.Connection = conn

		// extension map[filter:type_url:"agones.openmatch.filter" value:"{\"labels\":{\"region\":\"us-east-1\",\"world\":\"Pandora\"},\"fields\":{\"status.state\":\"Ready\"}}" ]
		// Stopped Here: Use Extensions["filter"] to unmarshall to extensions.AllocatorFilter and do the magic
		//

		runtime.Logger().WithField("component", "allocator").Debugf("extension %v", group.Assignment.Extensions)
		runtime.Logger().WithField("component", "allocator").Debugf("connection %s assigned to request", conn)
	}

	return nil
}

func (c *AgonesDiscoverAllocator) FindGameServer(ctx context.Context, filters map[interface{}]interface{}) ([]interface{}, error) {
	// TODO: Consume from Agones Discover API passing filter
	panic("implement me")
}
