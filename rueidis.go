package cachex

import (
	"context"
	"time"

	"github.com/redis/rueidis"
)

type rueidisRemote struct {
	rueidis.Client
}

func NewRueidisRemote(client rueidis.Client) Remote {
	return &rueidisRemote{
		Client: client,
	}
}

func (r *rueidisRemote) Set(ctx context.Context, key string, value []byte, expire time.Duration) error {
	return r.Client.Do(ctx, r.Client.B().Set().Key(key).Value(string(value)).Ex(expire).Build()).Error()
}

func (r *rueidisRemote) Get(ctx context.Context, key string) ([]byte, error) {
	return r.Client.Do(ctx, r.Client.B().Get().Key(key).Build()).AsBytes()
}

func (r *rueidisRemote) Del(ctx context.Context, key string) error {
	return r.Client.Do(ctx, r.Client.B().Del().Key(key).Build()).Error()
}

func (r *rueidisRemote) Nil() error {
	return rueidis.Nil
}
