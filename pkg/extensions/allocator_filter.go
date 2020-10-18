package extensions

import (
	"github.com/golang/protobuf/ptypes/any"
)

type AllocatorFilter struct {
	Labels map[string]string `json:"labels"`
	Fields map[string]string `json:"fields"`
}

func (f AllocatorFilter) Any() map[string]*any.Any {
	return map[string]*any.Any{
		"filter": ToAny("agones.openmatch.filter", f),
	}
}
