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
	GET_GAMESERVER_PATH = "api/v1/gameservers/"
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
	url := fmt.Sprintf("%s/%s", c.ServerURI, GET_GAMESERVER_PATH)
	resp, err := c.cli.Get(url)
	if err != nil {
		return []byte{}, errors.Wrap(err, "failed to list gameservers")
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, errors.Wrap(err, "failed to to ready gameservers response")
	}

	return body, nil
}
