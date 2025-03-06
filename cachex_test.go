package cachex

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/redis/rueidis"
	"github.com/stretchr/testify/assert"
)

type mockRemote struct {
	data map[string][]byte
	err  error
}

func (m *mockRemote) Set(ctx context.Context, key string, value []byte, expire time.Duration) error {
	if m.err != nil {
		return m.err
	}
	m.data[key] = value
	return nil
}

func (m *mockRemote) Get(ctx context.Context, key string) ([]byte, error) {
	if m.err != nil {
		return nil, m.err
	}
	val, ok := m.data[key]
	if !ok {
		return nil, rueidis.Nil
	}
	if bytes.Equal(val, notFoundPlaceholder) {
		return val, nil
	}
	return val, nil
}

func (m *mockRemote) Del(ctx context.Context, key string) error {
	delete(m.data, key)
	return nil
}

func (m *mockRemote) Nil() error {
	return rueidis.Nil
}

func TestOnce_Success(t *testing.T) {
	remote := &mockRemote{data: make(map[string][]byte)}
	value := "test value"

	got, err := Once(context.Background(), "key",
		func(ctx context.Context) (*string, error) {
			return &value, nil
		},
		WithRemote(remote),
	)

	assert.NoError(t, err)
	assert.Equal(t, value, *got)
	assert.Equal(t, `"test value"`, string(remote.data["key"]))
}

func TestOnce_NotFound(t *testing.T) {
	remote := &mockRemote{data: make(map[string][]byte)}

	got, err := Once(context.Background(), "key",
		func(ctx context.Context) (*string, error) {
			return nil, rueidis.Nil
		},
		WithRemote(remote),
		WithErrNotFound(rueidis.Nil),
	)

	assert.ErrorIs(t, err, rueidis.Nil)
	assert.Nil(t, got)
	assert.Equal(t, "*", string(remote.data["key"]))
}

func TestOnce_RemoteError(t *testing.T) {
	remote := &mockRemote{err: errors.New("remote error")}

	_, err := Once(context.Background(), "key",
		func(ctx context.Context) (*string, error) {
			return nil, nil
		},
		WithRemote(remote),
	)

	assert.ErrorContains(t, err, "remote error")
}

func TestOnce_Concurrent(t *testing.T) {
	remote := &mockRemote{data: make(map[string][]byte)}
	value := "concurrent value"

	results := make(chan *string, 10)
	for i := 0; i < 10; i++ {
		go func() {
			got, err := Once(context.Background(), "key",
				func(ctx context.Context) (*string, error) {
					time.Sleep(100 * time.Millisecond)
					return &value, nil
				},
				WithRemote(remote),
			)
			assert.NoError(t, err)
			results <- got
		}()
	}

	for i := 0; i < 10; i++ {
		got := <-results
		assert.Equal(t, value, *got)
	}
}

func BenchmarkOnce_CacheHit(b *testing.B) {
	remote := &mockRemote{data: make(map[string][]byte)}
	value := "benchmark value"
	// Pre-populate cache
	Once(context.Background(), "key",
		func(ctx context.Context) (*string, error) {
			return &value, nil
		},
		WithRemote(remote),
	)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Once(context.Background(), "key",
			func(ctx context.Context) (*string, error) {
				return &value, nil
			},
			WithRemote(remote),
		)
	}
}

func BenchmarkOnce_CacheMiss(b *testing.B) {
	remote := &mockRemote{data: make(map[string][]byte)}
	value := "benchmark value"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Once(context.Background(), "key",
			func(ctx context.Context) (*string, error) {
				return &value, nil
			},
			WithRemote(remote),
		)
	}
}

func BenchmarkOnce_Concurrent(b *testing.B) {
	remote := &mockRemote{data: make(map[string][]byte)}
	value := "benchmark value"

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			Once(context.Background(), "key",
				func(ctx context.Context) (*string, error) {
					return &value, nil
				},
				WithRemote(remote),
			)
		}
	})
}

func BenchmarkOnce_NotFound(b *testing.B) {
	remote := &mockRemote{data: make(map[string][]byte)}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Once(context.Background(), "key",
			func(ctx context.Context) (*string, error) {
				return nil, rueidis.Nil
			},
			WithRemote(remote),
		)
	}
}
