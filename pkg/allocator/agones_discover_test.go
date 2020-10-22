package allocator

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Octops/agones-discover-openmatch/pkg/extensions"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"math/rand"
	"open-match.dev/open-match/pkg/pb"
	"strconv"
	"testing"
	"time"
)

func TestAgonesDiscoverAllocator_Allocate(t *testing.T) {
	filter := &extensions.AllocatorFilterExtension{
		Labels: map[string]string{
			"region": "us-east-1",
		},
		Fields: map[string]string{
			"status.state": "Ready",
		},
	}

	testCases := []struct {
		name          string
		filter        *extensions.AllocatorFilterExtension
		ticketRequest *pb.AssignTicketsRequest
		gameServers   int
	}{
		{
			name:   "it should set Connection for Assignment if GameServers are returned",
			filter: filter,
			ticketRequest: &pb.AssignTicketsRequest{
				Assignments: []*pb.AssignmentGroup{
					{
						TicketIds: []string{
							uuid.New().String(),
							uuid.New().String(),
							uuid.New().String(),
						},
						Assignment: &pb.Assignment{
							Extensions: filter.Any(),
						},
					},
				},
			},
			gameServers: 1,
		},
		{
			name:   "it should not set Connection for Assignment if GameServers are not returned",
			filter: filter,
			ticketRequest: &pb.AssignTicketsRequest{
				Assignments: []*pb.AssignmentGroup{
					{
						TicketIds: []string{
							uuid.New().String(),
							uuid.New().String(),
							uuid.New().String(),
						},
						Assignment: &pb.Assignment{
							Extensions: filter.Any(),
						},
					},
				},
			},
			gameServers: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client := &mockAgonesDiscoverClient{}
			discoverAllocator := &AgonesDiscoverAllocator{
				Client: client,
			}

			client.On("ListGameServers", context.Background(), filter.Map()).
				Return(createRawGameServersWithLabels(tc.gameServers, filter.Labels), nil)

			err := discoverAllocator.Allocate(context.Background(), tc.ticketRequest)
			require.NoError(t, err)
			client.AssertExpectations(t)

			if tc.gameServers > 0 {
				for _, assignment := range tc.ticketRequest.Assignments {
					require.NotEmpty(t, assignment.Assignment.Connection)
				}
			} else {
				for _, assignment := range tc.ticketRequest.Assignments {
					require.Empty(t, assignment.Assignment.Connection)
				}
			}
		})
	}
}

func TestAgonesDiscoverAllocator_Call_FindGameServer(t *testing.T) {
	t.Run("it should call ListGameServers and not return error", func(t *testing.T) {
		client := &mockAgonesDiscoverClient{}
		discoverAllocator := &AgonesDiscoverAllocator{
			Client: client,
		}

		client.On("ListGameServers", context.Background(), map[string]string{}).Return([]byte{}, nil)
		_, err := discoverAllocator.FindGameServer(context.Background(), map[string]string{})
		require.NoError(t, err)

		client.AssertExpectations(t)
	})

	t.Run("it should call ListGameServers and return error", func(t *testing.T) {
		client := &mockAgonesDiscoverClient{}
		discoverAllocator := &AgonesDiscoverAllocator{
			Client: client,
		}

		client.On("ListGameServers", context.Background(), map[string]string{}).Return([]byte{}, errors.New("error"))
		_, err := discoverAllocator.FindGameServer(context.Background(), map[string]string{})
		require.Error(t, err)

		client.AssertExpectations(t)
	})
}

func TestAgonesDiscoverAllocator_HasCapacity(t *testing.T) {
	type wantError struct {
		want bool
		err  error
	}

	testCases := []struct {
		name           string
		gameservers    int
		assignments    int
		tickets        int
		playerCapacity int
		playerCount    int
		wantAssigned   int
		wantErr        wantError
	}{
		{
			name:           "it should return error for empty tickets",
			gameservers:    1,
			assignments:    1,
			tickets:        0,
			playerCapacity: 0,
			playerCount:    0,
			wantAssigned:   0,
			wantErr:        wantError{true, errors.New("assignment group has not tickets")},
		},
		{
			name:           "it should not assign for 0 ticket",
			gameservers:    1,
			assignments:    0,
			tickets:        0,
			playerCapacity: 0,
			playerCount:    0,
			wantAssigned:   0,
			wantErr:        wantError{false, nil},
		},
		{
			name:           "it should assign 1 ticket for capacity 1",
			gameservers:    1,
			assignments:    1,
			tickets:        1,
			playerCapacity: 1,
			playerCount:    0,
			wantAssigned:   1,
			wantErr:        wantError{false, nil},
		},
		{
			name:           "it should assign 1 ticket for capacity 2",
			gameservers:    1,
			assignments:    1,
			tickets:        1,
			playerCapacity: 2,
			playerCount:    0,
			wantAssigned:   1,
			wantErr:        wantError{false, nil},
		},
		{
			name:           "it should assign 2 ticket for capacity 2",
			gameservers:    1,
			assignments:    2,
			tickets:        2,
			playerCapacity: 2,
			playerCount:    0,
			wantAssigned:   2,
			wantErr:        wantError{false, nil},
		},
	}

	/*
		The number of tickets assigned should match the capacity available from the returned GameServers
		The capacity is a compute that uses Players Capacity - Players Count: 10-5 = 5 Max Tickets count to be assigned
		-
		- Group with 10 tickets and GS capacity = 10 == All tickets assigned
		- Group with 10 tickets and GS capacity = 5 == 5 Tickets assigned
		- Group with 20 Tickets and GS capacity = 0 == 0 Tickets assigned
		- Group with 0 Tickets and GS capacity = 0 == 0 Tickets assigned
		- Group with 0 Tickets and GS capacity = 10 == 0 Tickets assigned
	*/
	for _, tc := range testCases {
		filter := &extensions.AllocatorFilterExtension{
			Labels: map[string]string{
				"region": "us-east-1",
			},
			Fields: map[string]string{
				"status.state": "Ready",
			},
		}

		t.Run(tc.name, func(t *testing.T) {
			client := &mockAgonesDiscoverClient{}
			discoverAllocator := &AgonesDiscoverAllocator{
				Client: client,
			}

			gameservers := createGameServersWithCapacity(tc.gameservers, tc.playerCapacity, tc.playerCount, map[string]string{})
			gs, err := json.Marshal(gameservers)
			require.NoError(t, err)

			client.On("ListGameServers", context.Background(), filter.Map()).
				Return(gs, tc.wantErr.err)

			req := &pb.AssignTicketsRequest{
				Assignments: generateAssignments(tc.assignments, generateTicketsIds(tc.tickets), filter),
			}

			err = discoverAllocator.Allocate(context.Background(), req)
			if tc.wantErr.want {
				client.AssertNumberOfCalls(t, "ListGameServers", 0)
				require.Error(t, err)
				require.EqualError(t, err, tc.wantErr.err.Error())
			} else {
				totalAssigned := 0
				for _, a := range req.Assignments {
					client.AssertExpectations(t)
					if len(a.Assignment.Connection) > 0 {
						totalAssigned++
					}
				}

				require.NoError(t, err)
				require.Equal(t, tc.wantAssigned, totalAssigned)
			}
		})
	}
}

func generateAssignments(count int, tickets []string, filter *extensions.AllocatorFilterExtension) []*pb.AssignmentGroup {
	var group []*pb.AssignmentGroup

	for i := 0; i < count; i++ {
		group = append(group, &pb.AssignmentGroup{
			TicketIds: tickets,
			Assignment: &pb.Assignment{
				Extensions: filter.Any(),
			},
		})
	}

	return group
}

type mockAgonesDiscoverClient struct {
	mock.Mock
}

func (m *mockAgonesDiscoverClient) ListGameServers(ctx context.Context, filter map[string]string) ([]byte, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]byte), args.Error(1)
}

func createGameServersWithCapacity(count, playerCapacity, playerCount int, labels map[string]string) []*GameServer {
	gsRaw := createRawGameServersWithLabels(count, labels)

	var gameServers []*GameServer

	if err := json.Unmarshal(gsRaw, &gameServers); err != nil {
		return nil
	}

	for _, gs := range gameServers {
		gs.Status.Players.Capacity = int64(playerCapacity)
		gs.Status.Players.Count = int64(playerCount)
	}

	return gameServers
}

func createRawGameServersWithLabels(count int, labels map[string]string) []byte {
	gameservers := []*GameServer{}
	for i := 0; i < count; i++ {
		gs := &GameServer{
			UID:             uuid.New().String(),
			Name:            fmt.Sprintf("gameserver-%d", i),
			Namespace:       "default",
			ResourceVersion: strconv.Itoa(rand.Intn(10000)),
			Labels:          labels,
			Status: &GameServerStatus{
				State:   "Ready",
				Address: generateAddress(),
				Players: &PlayerStatus{
					Count:    int64(rand.Intn(100)),
					Capacity: 100,
					IDs:      nil,
				},
			},
		}
		gameservers = append(gameservers, gs)
	}

	b, err := json.Marshal(gameservers)
	if err != nil {
		return nil
	}

	return b
}

func generateTicketsIds(count int) []string {
	ticketsIds := []string{}
	for i := 0; i < count; i++ {
		ticketsIds = append(ticketsIds, uuid.New().String())
	}

	return ticketsIds
}

func generateAddress() string {
	rand.Seed(time.Now().UTC().UnixNano())
	return fmt.Sprintf("%d.%d.%d.%d:%d", rand.Intn(256), rand.Intn(256), rand.Intn(256), rand.Intn(256), rand.Intn(8000-7000)+7000)
}
