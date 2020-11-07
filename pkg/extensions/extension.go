package extensions

import (
	"encoding/json"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/golang/protobuf/ptypes/wrappers"
)

type Extension struct {
	any map[string]*any.Any
}

// Reference: https://stackoverflow.com/a/62585911
func ToAny(value interface{}) *any.Any {
	b, err := json.Marshal(value)
	if err != nil {
		return nil
	}

	bValues := &wrappers.BytesValue{
		Value: b,
	}

	mAny, err := ptypes.MarshalAny(bValues)
	if err != nil {
		return nil
	}

	return mAny
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
