package extensions

import (
	"encoding/json"
	"github.com/golang/protobuf/ptypes/any"
)

type Extension struct {
	any map[string]*any.Any
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

func WithAny(anyMap map[string]*any.Any) Extension {
	return Extension{any: anyMap}
}

func (ex Extension) WithAny(anyMap map[string]*any.Any) Extension {
	if ex.any == nil {
		ex.any = map[string]*any.Any{}
	}

	for k, v := range anyMap {
		ex.any[k] = v
	}

	return ex
}

func (ex Extension) Extensions() map[string]*any.Any {
	return ex.any
}