package jinx

import (
	"github.com/imlgw/jinx/codec"
	"log"
	"runtime"
	"sync"
)

type Server interface {
	Handler
	Run() error
}

type server struct {
	name      string
	network   string
	addr      string
	options   *Options
	ln        *listener
	started   bool
	wg        sync.WaitGroup
	loopGroup EventLoopGroup
	loopNum   int
	pipeline  handler
}

func (s *server) OnOpen(f func(c *connection)) {
	s.pipeline.onOpen = f
}

func (s *server) OnClose(f func(c *connection)) {
	s.pipeline.onClose = f
}

func (s *server) OnRead(f func(c *connection)) {
	s.pipeline.onRead = f
}

func (s *server) OnWrite(f func(c *connection)) {
	s.pipeline.onWrite = f
}

func (s *server) OnData(f func(c *connection)) {
	s.pipeline.onData = f
}

func NewServer(network, addr string, opts ...Option) (Server, error) {
	s := new(server)

	// option 加载
	s.options = LoadOptions(opts...)
	if s.options.Codec == nil {
		s.options.Codec = codec.NewDefaultLengthFieldCodec()
	}

	if s.options.LoopNum <= 0 {
		// 不设置默认是 cpu 个数
		s.loopNum = runtime.NumCPU()
	}

	s.network = network
	s.addr = addr

	// 初始化 loopGroup
	s.loopGroup = newEventGroup(s.options.Lb)

	// 创建 listener
	listener, err := newListener(s.network, s.addr)
	if err != nil {
		return nil, err
	}
	s.ln = listener

	return s, nil
}

func (s *server) Run() error {

	// 启动 listener
	s.wg.Add(1)
	go func() {
		if err := s.ln.run(); err != nil {
			s.wg.Done()
			return
		}
		s.wg.Done()
	}()

	// 创建并启动 loopNum 个事件循环
	for i := 0; i < s.loopNum; i++ {
		loop, err := newLoop(i)
		if err != nil {
			return err
		}
		s.loopGroup.Register(loop)
		go func() {
			err := loop.poll()
			if err != nil {
				log.Printf("create and run loop error, %v \n", err)
			}
		}()
	}
	return nil
}
