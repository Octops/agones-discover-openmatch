package allocator

import (
	"context"
	"encoding/json"
	"github.com/Octops/agones-discover-openmatch/internal/runtime"
	"github.com/Octops/agones-discover-openmatch/pkg/extensions"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/pkg/errors"
	"open-match.dev/open-match/pkg/pb"
)

var _ GameSessionAllocatorService = (*AgonesDiscoverAllocator)(nil)

type AgonesDiscoverClient interface {
	ListGameServers(ctx context.Context, filter map[string]string) ([]byte, error)
}

type AgonesDiscoverAllocator struct {
	Client AgonesDiscoverClient
}

func (c *AgonesDiscoverAllocator) Allocate(ctx context.Context, req *pb.AssignTicketsRequest) error {
	for _, group := range req.Assignments {
		filter, err := ExtractFilterFromExtensions(group.Assignment.Extensions)
		if err != nil {
			return errors.Wrap(err, "the assignment does not have a valid filter extension")
		}

		resp, err := c.Client.ListGameServers(ctx, filter.Map())
		gameservers, err := ParseGameServersResponse(resp)
		if err != nil {
			return errors.Wrap(err, "the response does not contain GameServers")
		}

		// TODO: Use the GameServer Player Capacity/Count field to validate if all tickets can be assigned.
		// NiceToHave: Filter GameServers by Capacity and Count
		if len(gameservers) > 0 {
			group.Assignment.Connection = gameservers[0].Status.Address
			runtime.Logger().WithField("component", "allocator").Debugf("extension %v", group.Assignment.Extensions)
			runtime.Logger().WithField("component", "allocator").Debugf("connection %s assigned to request", group.Assignment.Connection)
			continue
		}

		runtime.Logger().WithField("component", "allocator").Warn("request could not have a connection assigned")
	}

	return nil
}

func (c *AgonesDiscoverAllocator) FindGameServer(ctx context.Context, filters map[string]string) ([]byte, error) {
	return c.Client.ListGameServers(ctx, filters)
}

func ExtractFilterFromExtensions(extension map[string]*any.Any) (*extensions.AllocatorFilterExtension, error) {
	if _, ok := extension["filter"]; !ok {
		return nil, nil
	}

	filter, err := extensions.ToFilter(extension["filter"].Value)
	if err != nil {
		return nil, err
	}

	return filter, nil
}

func ParseGameServersResponse(resp []byte) ([]*GameServer, error) {
	var gameservers []*GameServer

	err := json.Unmarshal(resp, &gameservers)
	if err != nil {
		return nil, err
	}

	return gameservers, nil
}
