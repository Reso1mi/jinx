package jinx

import (
	"log"
	"net"
)

type EventLoopGroup interface {
	Next(addr net.Addr) *eventloop
	Register(e *eventloop)
	StopAll() error
}

type eventLoopGroup struct {
	loops       []*eventloop
	loadBalance interface {
		next(loops []*eventloop, addr net.Addr) *eventloop
	}
}

func newEventGroup(lb LoadBalance) EventLoopGroup {
	switch lb {
	case LeastConnections:
		return &eventLoopGroup{loadBalance: &leastConnections{}}
	case Random:
		return &eventLoopGroup{loadBalance: &random{}}
	case RoundRobin:
		return &eventLoopGroup{loadBalance: &roundRobin{}}
	default:
		return &eventLoopGroup{loadBalance: &roundRobin{}}
	}
}

func (g *eventLoopGroup) Next(addr net.Addr) *eventloop {
	return g.loadBalance.next(g.loops, addr)
}

func (g *eventLoopGroup) Register(e *eventloop) {
	g.loops = append(g.loops, e)
}

func (g *eventLoopGroup) StopAll() error {
	// 关闭所有 conn
	if err := g.CloseAllConn(); err != nil {
		log.Printf("close CloseAllConn error  %v \n", err)
		return err
	}

	for _, loop := range g.loops {
		if err := loop.Close(); err != nil {
			log.Printf("close eventloop error  %v \n", err)
			return err
		}
	}
	return nil
}

func (g *eventLoopGroup) CloseAllConn() error {
	for _, loop := range g.loops {
		for _, conn := range loop.reactor {
			if err := conn.Close(); err != nil {
				return err
			}
		}
	}
	return nil
}
