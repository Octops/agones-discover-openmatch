package allocator

import (
	"context"
	"open-match.dev/open-match/pkg/pb"
)

type GameServerDiscoveryServiceClient interface {
	FindGameServer(ctx context.Context, filters map[interface{}]interface{}) ([]interface{}, error)
}

type GameServerAllocatorServiceClient interface {
	Allocate(ctx context.Context, req *pb.AssignTicketsRequest) error
}

type GameServerAllocatorService interface {
	GameServerAllocatorServiceClient
	GameServerDiscoveryServiceClient
}
