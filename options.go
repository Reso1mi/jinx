package jinx

import (
	"github.com/imlgw/jinx/codec"
)

// Option is a function that will set up option.
type Option func(opts *Options)

func LoadOptions(options ...Option) *Options {
	opts := new(Options)
	for _, option := range options {
		option(opts)
	}
	return opts
}

// Options are configurations for the jinx application.
type Options struct {
	// server options
	Router Router

	// both server & client options
	Codec codec.ICodec

	// 负载均衡配置
	Lb LoadBalance
}

func WithRouter(r Router) Option {
	return func(opts *Options) {
		opts.Router = r
	}
}

func WithCodec(codec codec.ICodec) Option {
	return func(opts *Options) {
		opts.Codec = codec
	}
}

func WithLb(lb LoadBalance) Option {
	return func(opts *Options) {
		opts.Lb = lb
	}
}
