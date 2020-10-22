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
	logger := runtime.Logger().WithField("component", "allocator")

	for _, group := range req.Assignments {
		if err := IsValidForAllocation(group); err != nil {
			return err
		}

		filter, err := ExtractFilterFromExtensions(group.Assignment.Extensions)
		if err != nil {
			return errors.Wrap(err, "the assignment does not have a valid filter extension")
		}

		resp, err := c.Client.ListGameServers(ctx, filter.Map())
		gameservers, err := ParseGameServersResponse(resp)
		if err != nil {
			return errors.Wrap(err, "the response does not contain GameServers")
		}

		if len(gameservers) == 0 {
			logger.Warn("request could not have a connection assigned, no gameservers found")
			continue
		}

		// TODO: Use the GameServer Player Capacity/Count field to validate if all tickets can be assigned.
		// NiceToHave: Filter GameServers by Capacity and Count
		// Remove not assigned tickets based on playersCapacity - Count

		// strategy: allTogether, CapacityBased FallBack
		for _, gs := range gameservers {
			if HasCapacity(group, gs) {
				group.Assignment.Connection = gs.Status.Address
				logger.Debugf("extension %v", group.Assignment.Extensions)
				logger.Debugf("connection %s assigned to request", group.Assignment.Connection)
				break
			}
		}
	}

	return nil
}

func IsValidForAllocation(group *pb.AssignmentGroup) error {
	if group.Assignment == nil || group.Assignment.Extensions == nil {
		return errors.New("assignment or extension is nil")
	}

	if len(group.TicketIds) == 0 {
		return errors.New("assignment group has not tickets")
	}

	return nil
}

func HasCapacity(group *pb.AssignmentGroup, gs *GameServer) bool {
	capacity := gs.Status.Players.Capacity - gs.Status.Players.Count
	return capacity >= int64(len(group.TicketIds))
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
