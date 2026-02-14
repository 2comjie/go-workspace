package codec

import (
	"encoding/json"

	"google.golang.org/protobuf/proto"
)

type ISerializer interface {
	Marshal(v any) ([]byte, error)
	Unmarshal(data []byte, v any) error
}

type ProtoSerializer struct {
}

func (p ProtoSerializer) Marshal(v any) ([]byte, error) {
	pbV, ok := v.(proto.Message)
	if !ok {
		return nil, TypeNotSupportErr
	}
	return proto.Marshal(pbV)
}

func (p ProtoSerializer) Unmarshal(data []byte, v any) error {
	pbV, ok := v.(proto.Message)
	if !ok {
		return TypeNotSupportErr
	}
	return proto.Unmarshal(data, pbV)
}

type JsonSerializer struct {
}

func (j JsonSerializer) Marshal(v any) ([]byte, error) {
	return json.Marshal(v)
}

func (j JsonSerializer) Unmarshal(data []byte, v any) error {
	return json.Unmarshal(data, v)
}
