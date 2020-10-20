package extensions

import (
	"github.com/golang/protobuf/ptypes/any"
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
