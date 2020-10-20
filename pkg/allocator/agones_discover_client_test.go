package allocator

import (
	"context"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAgonesDiscoverClientHTTP_ListGameServers(t *testing.T) {
	t.Run("it should list gameservers with filter State=Ready", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("/api/v1/gameservers/", func(writer http.ResponseWriter, request *http.Request) {
			response := []*GameServer{
				{
					UID:       uuid.New().String(),
					Name:      "gameserver-udp",
					Namespace: "default",
				},
			}

			gs, err := json.Marshal(response)
			if err != nil {
				t.Fatal(err)
			}
			writer.Write(gs)
		})
		server := httptest.NewServer(mux)
		defer server.Close()

		discoverClient, err := NewAgonesDiscoverClientHTTP(server.URL)
		require.NoError(t, err)
		require.NotNil(t, discoverClient, "AgonesDiscover client HTTP can't be nil")

		response, err := discoverClient.ListGameServers(context.Background(), map[string]string{})
		require.NoError(t, err)

		gameservers := []*GameServer{}

		err = json.Unmarshal(response, &gameservers)
		require.NoError(t, err)
	})
}
