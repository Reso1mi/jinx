package jinx

import (
	"log"
	"runtime"
	"sync"
)

type Server interface {
	handler
	Run() error
	Stop() error
	ServerName() string
	Network() string
	ServerAddr() string
	Started() bool
}

type server struct {
	network    string
	addr       string
	opts       *Options
	ln         *listener
	started    bool
	wg         sync.WaitGroup
	loopGroup  *eventLoopGroup
	onBoot     func(s Server)
	onOpen     func(c Conn)
	onClose    func(c Conn)
	onRead     func(c Conn)
	onWrite    func(c Conn)
	onShutdown func(s Server)
}

func NewServer(network, addr string, opts ...Option) (Server, error) {
	s := new(server)
	// option 加载
	options := LoadOptions(opts...)
	if options.LoopNum <= 0 {
		// 不设置默认是 cpu 个数
		options.LoopNum = runtime.NumCPU()
	}
	if options.ServerName == "" {
		options.ServerName = "baobao"
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
		s.wg.Done()
		if err := s.ln.run(); err != nil {
			s.wg.Done()
			return
		}
	}()

	// 创建并启动 loopNum 个事件循环
	for i := 0; i < s.opts.LoopNum; i++ {
		s.wg.Add(1)
		loop, err := newLoop(i, s)
		if err != nil {
			return err
		}
		s.loopGroup.register(loop)
		go func() {
			// 服务启动 latch
			s.wg.Done()
			if err := loop.poll(); err != nil {
				log.Printf("create and run loop error, %v \n", err)
			}
			// 服务关闭 latch，避免并发修改 map 异常
			s.loopGroup.wg.Done()
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

	// TODO: 唤醒所有 eventloop 退出 loop

	// 等待所有 loop
	// s.loopGroup.wg.Wait()

	// 关闭所有 connection 以及 subReactor 的 eventloop
	if err := s.loopGroup.stopAll(); err != nil {
		return err
	}

	// 关闭 listener 以及 mainReactor 的 eventloop
	if err := s.ln.Close(); err != nil {
		return err
	}

	return nil
}

func (s *server) ServerName() string          { return s.opts.ServerName }
func (s *server) Network() string             { return s.network }
func (s *server) ServerAddr() string          { return s.addr }
func (s *server) Started() bool               { return s.started }
func (s *server) OnBoot(f func(s Server))     { s.onBoot = f }
func (s *server) OnOpen(f func(c Conn))       { s.onOpen = f }
func (s *server) OnClose(f func(c Conn))      { s.onClose = f }
func (s *server) OnRead(f func(c Conn))       { s.onRead = f }
func (s *server) OnWrite(f func(c Conn))      { s.onWrite = f }
func (s *server) OnShutdown(f func(s Server)) { s.onShutdown = f }
