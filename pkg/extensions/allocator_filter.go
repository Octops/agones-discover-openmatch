package extensions

import (
	"encoding/json"
	"fmt"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/golang/protobuf/ptypes/wrappers"
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

func ExtractFilterFromExtensions(extension map[string]*any.Any) (*AllocatorFilterExtension, error) {
	if _, ok := extension["filter"]; !ok {
		return nil, nil
	}

	filter, err := ToFilter(extension["filter"])
	if err != nil {
		return nil, err
	}

	return filter, nil
}

func ToFilter(obj *any.Any) (*AllocatorFilterExtension, error) {
	var filter AllocatorFilterExtension

	message := &wrappers.BytesValue{}
	err := ptypes.UnmarshalAny(obj, message)
	if err != nil {
		return nil, errors.Wrap(err, "can't parse Any to Message")
	}

	err = json.Unmarshal(message.Value, &filter)
	if err != nil {
		return nil, errors.Wrap(err, "can't parse Any to Filter")
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
