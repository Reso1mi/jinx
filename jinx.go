package jinx

import (
	"github.com/imlgw/jinx/codec"
	"log"
	"runtime"
)

type Server interface {

	// Start 启动服务器
	Start()

	// Stop 停止服务器
	Stop()

	// Serve 开始服务
	Serve()
}

type server struct {
	name      string
	ipVersion string
	ip        string
	port      int
	opts      *Options
}

func (s *server) Start() {

}

func (s *server) Stop() {

}

func (s *server) Serve() {
}

func Run(network, addr string, opts ...Option) error {
	options := LoadOptions(opts...)
	if options.Codec == nil {
		options.Codec = codec.NewDefaultLengthFieldCodec()
	}

	// 初始化 loopGroup
	loopGroup = NewEventGroup(options.Lb)

	// 创建并启动 listener
	listener, err := NewListener(network, addr)
	if err != nil {
		return err
	}
	go listener.Run()

	loopNum := options.LoopNum
	if loopNum <= 0 {
		// 不设置默认是 cpu 个数
		loopNum = runtime.NumCPU()
	}

	// 创建并启动 loopNum 个事件循环
	for i := 0; i < loopNum; i++ {
		loop, err := NewLoop(i)
		if err != nil {
			return err
		}
		loopGroup.Register(loop)
		go func() {
			err := loop.Loop()
			if err != nil {
				log.Printf("create and run loop error, %v \n", err)
			}
		}()
	}
	return nil
}
