package allocator

import (
	pb_agones "agones.dev/agones/pkg/allocation/go"
	"context"
	"fmt"
	"github.com/Octops/agones-discover-openmatch/internal/runtime"
	"github.com/Octops/agones-discover-openmatch/pkg/extensions"
	"github.com/pkg/errors"
	"google.golang.org/grpc/status"
	"open-match.dev/open-match/pkg/pb"
)

var _ GameServerAllocatorService = (*AgonesAllocator)(nil)

var (
	NotAvailableGameServerToAllocateMessage = "there is no available GameServer to allocate"
)

type AgonesAllocator struct {
	Client *AgonesAllocatorClient
}

func NewAgonesAllocator(client *AgonesAllocatorClient) *AgonesAllocator {
	return &AgonesAllocator{Client: client}
}

func (a *AgonesAllocator) Allocate(ctx context.Context, req *pb.AssignTicketsRequest) error {
	logger := runtime.Logger().WithField("component", "allocator")

	for _, assignmentGroup := range req.Assignments {
		filter, err := extensions.ExtractFilterFromExtensions(assignmentGroup.Assignment.Extensions)
		if err != nil {
			return errors.Wrap(err, "the assignment does not have a valid filter extension")
		}

		//TODO: Add PreferredGameServerSelector, MetaPatch, Scheduling. It must be part of the extensions
		request := &pb_agones.AllocationRequest{
			Namespace: a.Client.Config.Namespace,
			GameServerSelectors: []*pb_agones.GameServerSelector{
				{
					MatchLabels: filter.Labels,
				},
			},
			MultiClusterSetting: &pb_agones.MultiClusterSetting{
				Enabled: a.Client.Config.MultiCluster,
			},
		}

		resp, err := a.Client.Allocate(ctx, request)
		if err != nil {
			errStatus, ok := status.FromError(err)
			if ok && errStatus.Message() == NotAvailableGameServerToAllocateMessage {
				logger.Debug(NotAvailableGameServerToAllocateMessage)
				continue
			}

			return err
		}

		if len(resp.GetPorts()) > 0 {
			address := fmt.Sprintf("%s:%d", resp.Address, resp.Ports[0].Port)
			assignmentGroup.Assignment.Connection = address
			logger.Infof("gameserver %s connection %s assigned to request, total tickets: %d", resp.GameServerName, assignmentGroup.Assignment.Connection, len(assignmentGroup.TicketIds))
		}
	}

	return nil
}

func ValueIsEmpty(value string, err error) (bool, error) {
	if len(value) == 0 {
		return true, err
	}

	return false, nil
}
