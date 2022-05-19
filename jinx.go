package jinx

import (
	"log"
	"runtime"
	"sync"
)

type Server interface {
	Handler
	Run() error
	Stop() error
}

type server struct {
	name       string
	network    string
	addr       string
	opts       *Options
	ln         *listener
	started    bool
	wg         sync.WaitGroup
	loopGroup  EventLoopGroup
	onBoot     func(s *server)
	onOpen     func(c *connection)
	onClose    func(c *connection)
	onRead     func(c *connection)
	onWrite    func(c *connection)
	onShutdown func(s *server)
}

func NewServer(network, addr string, opts ...Option) (Server, error) {
	s := new(server)
	// option 加载
	options := LoadOptions(opts...)
	if options.LoopNum <= 0 {
		// 不设置默认是 cpu 个数
		options.LoopNum = runtime.NumCPU()
	}

	s.opts = options
	s.network = network
	s.addr = addr
	// 初始化 loopGroup
	s.loopGroup = newEventGroup(s.opts.Lb)

	// 创建 listener
	listener, err := newListener(s.network, s.addr, s)
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
	for i := 0; i < s.opts.LoopNum; i++ {
		s.wg.Add(1)
		loop, err := newLoop(i, s)
		if err != nil {
			return err
		}
		s.loopGroup.Register(loop)
		go func() {
			s.wg.Done()
			err := loop.poll()
			if err != nil {
				log.Printf("create and run loop error, %v \n", err)
			}
		}()
	}
	s.wg.Wait()
	s.started = true

	if s.onBoot != nil {
		s.onBoot(s)
	}
	return nil
}

func (s *server) Stop() error {
	if s.onShutdown != nil {
		s.onShutdown(s)
	}

	// 关闭所有 connection 以及 subReactor 的 eventloop
	if err := s.loopGroup.StopAll(); err != nil {
		return err
	}

	// 关闭 listener 以及 mainReactor 的 eventloop
	if err := s.ln.Close(); err != nil {
		return err
	}

	return nil
}

func (s *server) OnBoot(f func(s *server))      { s.onBoot = f }
func (s *server) OnOpen(f func(c *connection))  { s.onOpen = f }
func (s *server) OnClose(f func(c *connection)) { s.onClose = f }
func (s *server) OnRead(f func(c *connection))  { s.onRead = f }
func (s *server) OnWrite(f func(c *connection)) { s.onWrite = f }
func (s *server) OnShutdown(f func(s *server))  { s.onShutdown = f }
