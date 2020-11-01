package allocator

import (
	"context"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAgonesDiscoverClientHTTP_ListGameServers_Filters(t *testing.T) {
	testCases := []struct {
		name    string
		handler func(writer http.ResponseWriter, request *http.Request)
		filter  map[string]string
		want    []byte
	}{
		{
			name: "it should request GameServers endpoint",
			handler: func(writer http.ResponseWriter, request *http.Request) {
				writer.WriteHeader(http.StatusOK)
				writer.Write([]byte(request.URL.Path))
			},
			want: []byte(GET_GAMESERVER_PATH),
		},
		{
			name: "it should request GameServers endpoint and filter with one labelSelector",
			handler: func(writer http.ResponseWriter, request *http.Request) {
				writer.WriteHeader(http.StatusOK)
				writer.Write([]byte(request.URL.RawQuery))
			},
			filter: map[string]string{
				"labels": "region=us-east-1",
			},
			want: []byte("labels=region%3Dus-east-1"),
		},
		{
			name: "it should request GameServers endpoint and filter with two labels",
			handler: func(writer http.ResponseWriter, request *http.Request) {
				writer.WriteHeader(http.StatusOK)
				writer.Write([]byte(request.URL.RawQuery))
			},
			filter: map[string]string{
				"labels": "region=us-east-1,cluster=gke-1.16",
			},
			want: []byte("labels=region%3Dus-east-1%2Ccluster%3Dgke-1.16"),
		},
		{
			name: "it should request GameServers endpoint and filter with two labels and State Ready",
			handler: func(writer http.ResponseWriter, request *http.Request) {
				writer.WriteHeader(http.StatusOK)
				writer.Write([]byte(request.URL.RawQuery))
			},
			filter: map[string]string{
				"labels": "region=us-east-1,cluster=gke-1.16",
				"fields": "status.state=Ready",
			},
			want: []byte("fields=status.state%3DReady&labels=region%3Dus-east-1%2Ccluster%3Dgke-1.16"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mux := http.NewServeMux()
			//GET_GAMESERVER_PATH = "api/v1/gameservers"
			mux.HandleFunc("/api/v1/gameservers", tc.handler)
			server := httptest.NewServer(mux)
			defer server.Close()

			discoverClient, err := NewAgonesDiscoverClientHTTP(server.URL)
			require.NoError(t, err)
			require.NotNil(t, discoverClient, "AgonesDiscover client HTTP can't be nil")

			got, err := discoverClient.ListGameServers(context.Background(), tc.filter)
			require.NoError(t, err)
			require.Equal(t, string(tc.want), string(got))
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
			name:   "it should return empty filter",
			filter: map[string]string{},
			want:   "",
		},
		{
			name: "it should return filter for one labels",
			filter: map[string]string{
				"labels": "region=us-east-1",
			},
			want: "labels=region%3Dus-east-1",
		},
		{
			name: "it should return filter for two labels",
			filter: map[string]string{
				"labels": "region=us-east-1,cluster=gke-1.16",
			},
			want: "labels=region%3Dus-east-1%2Ccluster%3Dgke-1.16",
		},
		{
			name: "it should return filter for three labels",
			filter: map[string]string{
				"labels": "region=us-east-1,cluster=gke-1.16,version=1.0",
			},
			want: "labels=region%3Dus-east-1%2Ccluster%3Dgke-1.16%2Cversion%3D1.0",
		},
		{
			name: "it should return filter for State Ready",
			filter: map[string]string{
				"fields": "status.state=Ready",
			},
			want: "fields=status.state%3DReady",
		},
		{
			name: "it should return filter for State Scheduled",
			filter: map[string]string{
				"fields": "status.state=Scheduled",
			},
			want: "fields=status.state%3DScheduled",
		},
		{
			name: "it should return filter for one labels and State Ready",
			filter: map[string]string{
				"labels": "region=us-east-1",
				"fields": "status.state=Ready",
			},
			want: "fields=status.state%3DReady&labels=region%3Dus-east-1",
		},
		{
			name: "it should return filter for one labels and State Scheduled",
			filter: map[string]string{
				"labels": "region=us-east-1",
				"fields": "status.state=Scheduled",
			},
			want: "fields=status.state%3DScheduled&labels=region%3Dus-east-1",
		},
		{
			name: "it should return filter for two labels and State Ready",
			filter: map[string]string{
				"labels": "region=us-east-1,cluster=gke-1.16",
				"fields": "status.state=Ready",
			},
			want: "fields=status.state%3DReady&labels=region%3Dus-east-1%2Ccluster%3Dgke-1.16",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := EncodeFilter(tc.filter)
			require.Equal(t, tc.want, got)
		})
	}
}
