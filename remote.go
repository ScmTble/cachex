package cachex

import (
	"context"
	"time"
)

type Remote interface {
	Set(ctx context.Context, key string, value []byte, expire time.Duration) error

	Get(ctx context.Context, key string) ([]byte, error)

	Del(ctx context.Context, key string) error

	Nil() error
}
