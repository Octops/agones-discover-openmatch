package allocator

import (
	"context"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestAgonesDiscoverAllocator_Allocate(t *testing.T) {

}

func TestAgonesDiscoverAllocator_Call_FindGameServer(t *testing.T) {
	t.Run("it should call ListGameServers", func(t *testing.T) {
		client := &mockAgonesDiscoverClient{}
		discoverAllocator := &AgonesDiscoverAllocator{
			Client: client,
		}

		client.On("ListGameServers", context.Background(), map[string]string{}).Return([]byte{})
		_, err := discoverAllocator.FindGameServer(context.Background(), map[string]string{})
		require.NoError(t, err)

		client.AssertExpectations(t)
	})
}

type mockAgonesDiscoverClient struct {
	mock.Mock
}

func (m *mockAgonesDiscoverClient) ListGameServers(ctx context.Context, filter map[string]string) ([]byte, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]byte), nil
}
