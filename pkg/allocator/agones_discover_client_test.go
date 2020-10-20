package allocator

import (
	"context"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAgonesDiscoverClientHTTP_ListGameServers(t *testing.T) {
	testCases := []struct {
		name     string
		handler  func(writer http.ResponseWriter, request *http.Request)
		filter   map[string]string
		response []byte
	}{
		{
			name: "it should request GameServers endpoint",
			handler: func(writer http.ResponseWriter, request *http.Request) {
				writer.WriteHeader(http.StatusOK)
				writer.Write([]byte(request.URL.Path))
			},
			response: []byte(GET_GAMESERVER_PATH),
		},
		{
			name: "it should request GameServers endpoint and filter with one label",
			handler: func(writer http.ResponseWriter, request *http.Request) {
				writer.WriteHeader(http.StatusOK)
				writer.Write([]byte(request.URL.RawQuery))
			},
			response: []byte("filter=status.state%3DReady"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mux := http.NewServeMux()
			mux.HandleFunc(GET_GAMESERVER_PATH, tc.handler)
			server := httptest.NewServer(mux)
			defer server.Close()

			discoverClient, err := NewAgonesDiscoverClientHTTP(server.URL)
			require.NoError(t, err)
			require.NotNil(t, discoverClient, "AgonesDiscover client HTTP can't be nil")

			response, err := discoverClient.ListGameServers(context.Background(), map[string]string{})
			require.NoError(t, err)
			require.Equal(t, string(tc.response), string(response))
		})
	}
}

func TestEncodeFilter(t *testing.T) {
	testCases := []struct {
		name   string
		filter map[string]string
		want   string
	}{
		{
			name: "it should return filter for one labelSelector",
			filter: map[string]string{
				"labelSelector": "region=us-east-1",
			},
			want: "labelSelector=region%3Dus-east-1",
		},
		{
			name: "it should return filter for two labelSelector",
			filter: map[string]string{
				"labelSelector": "region=us-east-1,cluster=gke-1.16",
			},
			want: "labelSelector=region%3Dus-east-1%2Ccluster%3Dgke-1.16",
		},
		{
			name: "it should return filter for three labelSelector",
			filter: map[string]string{
				"labelSelector": "region=us-east-1,cluster=gke-1.16,version=1.0",
			},
			want: "labelSelector=region%3Dus-east-1%2Ccluster%3Dgke-1.16%2Cversion%3D1.0",
		},
		{
			name: "it should return filter for State Ready",
			filter: map[string]string{
				"status.state": "Ready",
			},
			want: "status.state=Ready",
		},
		{
			name: "it should return filter for State Scheduled",
			filter: map[string]string{
				"status.state": "Scheduled",
			},
			want: "status.state=Scheduled",
		},
		{
			name: "it should return filter for one labelSelector and State Ready",
			filter: map[string]string{
				"labelSelector": "region=us-east-1",
				"status.state":  "Ready",
			},
			want: "labelSelector=region%3Dus-east-1&status.state=Ready",
		},
		{
			name: "it should return filter for one labelSelector and State Scheduled",
			filter: map[string]string{
				"labelSelector": "region=us-east-1",
				"status.state":  "Scheduled",
			},
			want: "labelSelector=region%3Dus-east-1&status.state=Scheduled",
		},
		{
			name: "it should return filter for two labelSelector and State Ready",
			filter: map[string]string{
				"labelSelector": "region=us-east-1,cluster=gke-1.16",
				"status.state":  "Ready",
			},
			want: "labelSelector=region%3Dus-east-1%2Ccluster%3Dgke-1.16&status.state=Ready",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := EncodeFilter(tc.filter)
			require.Equal(t, tc.want, got)
		})
	}
}
