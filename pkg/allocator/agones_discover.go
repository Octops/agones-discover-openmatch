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

type GameServersResponse struct {
	Data []*GameServer `json:"data"`
}

// Allocate will only assign a GameServer to an Assignment if the Capacity (Players.Status.Capacity - Players.Stats.Count)
// is <= the number of the TicketsIds part of the Assignment
func (c *AgonesDiscoverAllocator) Allocate(ctx context.Context, req *pb.AssignTicketsRequest) error {
	logger := runtime.Logger().WithField("component", "allocator")

	for _, assignmentGroup := range req.Assignments {
		if err := IsAssignmentGroupValidForAllocation(assignmentGroup); err != nil {
			return err
		}

		filter, err := ExtractFilterFromExtensions(assignmentGroup.Assignment.Extensions)
		if err != nil {
			return errors.Wrap(err, "the assignment does not have a valid filter extension")
		}

		gameservers, err := c.ListGameServers(ctx, filter)
		if err != nil {
			logger.Error(err)
			return err
		}

		if len(gameservers) == 0 {
			logger.Debugf("gameservers not found for request with filter %v", filter.Map())
			continue
		}

		// NiceToHave: Filter GameServers by Capacity and Count
		// Remove not assigned tickets based on playersCapacity - Count
		// strategy: allTogether, CapacityBased FallBack
		for _, gs := range gameservers {
			if HasCapacity(assignmentGroup, gs) {
				assignmentGroup.Assignment.Connection = gs.Status.Address
				//logger.Debugf("extension %v", assignmentGroup.Assignment.Extensions)
				logger.Infof("gameserver %s connection %s assigned to request, total tickets: %d", gs.Name, assignmentGroup.Assignment.Connection, len(assignmentGroup.TicketIds))
				break
			}
		}
	}

	return nil
}

func (c *AgonesDiscoverAllocator) ListGameServers(ctx context.Context, filter *extensions.AllocatorFilterExtension) ([]*GameServer, error) {
	resp, err := c.FindGameServers(ctx, filter.Map())
	if err != nil {
		if err == ErrGameServersNotFound {
			return nil, err
		}

		return nil, errors.Wrap(err, "the response does not contain GameServers")
	}

	gameservers, err := ParseGameServersResponse(resp)
	if err != nil {
		runtime.Logger().Errorf("%v", resp)
		return nil, errors.Wrap(err, "error parsing gameservers from response")
	}

	return gameservers, nil
}

func (c *AgonesDiscoverAllocator) FindGameServers(ctx context.Context, filters map[string]string) ([]byte, error) {
	return c.Client.ListGameServers(ctx, filters)
}

func IsAssignmentGroupValidForAllocation(group *pb.AssignmentGroup) error {
	if group.Assignment == nil || group.Assignment.Extensions == nil {
		return errors.New("assignment or extension is nil")
	}

	if len(group.TicketIds) == 0 {
		return errors.New("assignment group has not tickets")
	}

	return nil
}

func HasCapacity(group *pb.AssignmentGroup, gs *GameServer) bool {
	// Allow any number of users to join the GameServer if the PlayerTracking feature flag is not enabled
	if gs.Status.Players == nil {
		return true
	}

	capacity := gs.Status.Players.Capacity - gs.Status.Players.Count
	// If Count and Capacity are not it should allow allocation. This is just a possible scenario to be reviewed in the future
	return (capacity >= int64(len(group.TicketIds))) || (gs.Status.Players.Count == 0 && gs.Status.Players.Capacity == 0)
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
	var items GameServersResponse

	err := json.Unmarshal(resp, &items)
	if err != nil {
		return nil, err
	}

	return items.Data, nil
}
