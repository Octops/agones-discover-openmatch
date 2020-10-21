package extensions

import (
	"fmt"
	"github.com/golang/protobuf/ptypes/any"
	"strings"
)

type AllocatorFilterExtension struct {
	Labels map[string]string `json:"labels"`
	Fields map[string]string `json:"fields"`
}

func (f AllocatorFilterExtension) Any() map[string]*any.Any {
	return map[string]*any.Any{
		"filter": ToAny("agones.openmatch.filter", f),
	}
}

func (f *AllocatorFilterExtension) Map() map[string]string {
	m := map[string]string{}
	m["labels"] = joinMapValues(f.Labels)
	m["fields"] = joinMapValues(f.Fields)

	return m
}

func joinMapValues(list map[string]string) string {
	var values []string

	for k, v := range list {
		values = append(values, fmt.Sprintf("%s=%s", k, v))
	}

	return strings.Join(values, ",")
}
