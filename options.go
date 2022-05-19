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
	// 服务器名称
	ServerName string

	// both server & client options
	Codec codec.ICodec

	// 负载均衡配置
	Lb LoadBalance

	// subReactor 对应的 eventloop 数量
	LoopNum int
}

func WithServerName(name string) Option {
	return func(opts *Options) {
		opts.ServerName = name
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

func WithLoopNum(loopNum int) Option {
	return func(opts *Options) {
		opts.LoopNum = loopNum
	}
}
