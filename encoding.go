package cachex

import "encoding/json"

var (
	defaultCodec = &jsonCodec{}
)

type Codec interface {
	Marshal(v any) ([]byte, error)
	Unmarshal(data []byte, v any) error
}

var _ Codec = (*jsonCodec)(nil)

type jsonCodec struct{}

func (c *jsonCodec) Marshal(v any) ([]byte, error) {
	return json.Marshal(v)
}

func (c *jsonCodec) Unmarshal(data []byte, v any) error {
	return json.Unmarshal(data, v)
}
