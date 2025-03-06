package cachex

import (
	"errors"
	"time"
)

var (
	ErrNotFound = errors.New("not found")
)

type (
	Options struct {
		errNotFound    error
		remoteExpiry   time.Duration
		notFoundExpiry time.Duration
		remote         Remote
		codec          Codec
	}

	Option func(o *Options)
)

func newOptions(opts ...Option) *Options {
	o := &Options{
		errNotFound:    ErrNotFound,
		remoteExpiry:   time.Hour,
		notFoundExpiry: time.Minute,
		codec:          defaultCodec,
	}
	for _, opt := range opts {
		opt(o)
	}
	return o
}

func WithErrNotFound(err error) Option {
	return func(o *Options) {
		o.errNotFound = err
	}
}

func WithRemote(remote Remote) Option {
	return func(o *Options) {
		o.remote = remote
	}
}

func WithRemoteExpiry(expiry time.Duration) Option {
	return func(o *Options) {
		o.remoteExpiry = expiry
	}
}

func WithNotFoundExpiry(expiry time.Duration) Option {
	return func(o *Options) {
		o.notFoundExpiry = expiry
	}
}
