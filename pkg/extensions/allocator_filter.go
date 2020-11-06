package extensions

import (
	"encoding/json"
	"fmt"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/pkg/errors"
	"strings"
)

type AllocatorFilterExtension struct {
	Labels map[string]string `json:"labels"`
	Fields map[string]string `json:"fields"`
}

func (f AllocatorFilterExtension) Any() map[string]*any.Any {
	return map[string]*any.Any{
		"filter": ToAny(f),
	}
}

func (f *AllocatorFilterExtension) Map() map[string]string {
	m := map[string]string{}
	m["labels"] = joinMapValues(f.Labels)
	m["fields"] = joinMapValues(f.Fields)

	return m
}

func ToFilter(value []byte) (*AllocatorFilterExtension, error) {
	var filter AllocatorFilterExtension

	value = value[2:]
	err := json.Unmarshal(value, &filter)
	if err != nil {
		return nil, errors.Wrap(err, "can't parse to filter")
	}

	return &filter, nil
}

func joinMapValues(list map[string]string) string {
	var values []string

	for k, v := range list {
		values = append(values, fmt.Sprintf("%s=%s", k, v))
	}

	return strings.Join(values, ",")
}
