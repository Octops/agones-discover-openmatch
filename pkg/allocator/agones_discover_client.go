package allocator

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

const (
	GET_GAMESERVER_PATH = "api/v1/gameservers"
)

type AgonesDiscoverClientHTTP struct {
	cli       *http.Client
	ServerURI string
}

func NewAgonesDiscoverClientHTTP(serverURI string) (*AgonesDiscoverClientHTTP, error) {
	uri, err := url.Parse(serverURI)
	if err != nil {
		return nil, err
	}

	return &AgonesDiscoverClientHTTP{
		cli: &http.Client{
			Timeout: time.Second * 5,
		},
		ServerURI: uri.String(),
	}, nil
}

func (c *AgonesDiscoverClientHTTP) ListGameServers(ctx context.Context, filter map[string]string) ([]byte, error) {
	u := BuildQueryParams(GET_GAMESERVER_PATH, filter)
	resp, err := c.cli.Get(fmt.Sprintf("%s/%s", c.ServerURI, u))
	if err != nil {
		return nil, errors.Wrap(err, "failed to list gameservers")
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrGameServersNotFound
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to to ready gameservers response")
	}

	return body, nil
}

func BuildQueryParams(path string, filter map[string]string) string {
	return fmt.Sprintf("%s?%s", path, EncodeFilter(filter))
}

func EncodeFilter(filter map[string]string) string {
	params := url.Values{}

	for k, v := range filter {
		params.Add(k, v)
	}

	return params.Encode()
}
