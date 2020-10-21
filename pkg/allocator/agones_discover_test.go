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

/*
Call Allocate and find a gameserver using the filter DeepEqual is enough
Assign Conn to tickets
Validate AssignTicketsRequest
*/
func TestAgonesDiscoverAllocator_Allocate(t *testing.T) {
	t.Run("it should assign conn to tickets", func(t *testing.T) {
		client := &mockAgonesDiscoverClient{}
		discoverAllocator := &AgonesDiscoverAllocator{
			Client: client,
		}

		filter := &extensions.AllocatorFilterExtension{
			Labels: map[string]string{
				"region": "us-east-1",
			},
			Fields: map[string]string{
				"status.state": "Ready",
			},
		}

		client.On("ListGameServers", context.Background(), filter.Map()).
			Return(createGameServersWithLabels(1, filter.Labels), nil)

		req := &pb.AssignTicketsRequest{
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
		}

		err := discoverAllocator.Allocate(context.Background(), req)
		require.NoError(t, err)
		client.AssertExpectations(t)

	})
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

type mockAgonesDiscoverClient struct {
	mock.Mock
}

func (m *mockAgonesDiscoverClient) ListGameServers(ctx context.Context, filter map[string]string) ([]byte, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]byte), args.Error(1)
}

func createGameServersWithLabels(count int, labels map[string]string) []byte {
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

func generateAddress() string {
	rand.Seed(time.Now().UTC().UnixNano())
	return fmt.Sprintf("%d.%d.%d.%d:%d", rand.Intn(256), rand.Intn(256), rand.Intn(256), rand.Intn(256), rand.Intn(8000-7000)+7000)
}
