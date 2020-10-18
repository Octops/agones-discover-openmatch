package extensions

import (
	"encoding/json"
	"github.com/golang/protobuf/ptypes/any"
)

type AllocatorFilter struct {
	Labels map[string]string `json:"labels"`
	Fields map[string]string `json:"fields"`
}

func ToAny(typeURL string, value interface{}) *any.Any {
	b, err := json.Marshal(value)
	if err != nil {
		return nil
	}

	return &any.Any{
		TypeUrl: typeURL,
		Value:   b,
	}
}
