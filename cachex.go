package cachex

import (
	"bytes"
	"context"
	"errors"

	"golang.org/x/sync/singleflight"
)

var (
	notFoundPlaceholder = []byte("*")
	group               singleflight.Group
)

type Fn[T any] func(ctx context.Context) (*T, error)

func Once[T any](ctx context.Context, key string, fn Fn[T], opts ...Option) (*T, error) {
	opt := newOptions(opts...)
	unmarshalFn := func(data []byte) (*T, error) {
		var t T
		if err := opt.codec.Unmarshal(data, &t); err != nil {
			return nil, err
		}
		return &t, nil
	}

	remoteData, err := opt.remote.Get(ctx, key)
	if err == nil {
		if bytes.Equal(remoteData, notFoundPlaceholder) {
			return nil, opt.errNotFound
		}

		val, err := unmarshalFn(remoteData)
		if err != nil {
			_ = opt.remote.Del(ctx, key)
			return nil, err
		}

		return val, nil
	}
	if !errors.Is(err, opt.remote.Nil()) {
		return nil, err
	}

	val, err, _ := group.Do(key, func() (any, error) {
		val, err := fn(ctx)
		if errors.Is(err, opt.errNotFound) || (err == nil && val == nil) {
			if err := opt.remote.Set(ctx, key, notFoundPlaceholder, opt.notFoundExpiry); err != nil {
				return nil, err
			}
			return nil, opt.errNotFound
		}
		if err != nil {
			return nil, err
		}

		data, err := opt.codec.Marshal(val)
		if err != nil {
			return nil, err
		}

		if err := opt.remote.Set(ctx, key, data, opt.remoteExpiry); err != nil {
			return nil, err
		}

		return val, nil
	})
	if err != nil {
		return nil, err
	}

	return val.(*T), nil
}
