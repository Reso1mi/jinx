package jinx

import (
	"math/rand"
	"net"
	"time"
)

var loopGroup EventLoopGroup

type EventLoopGroup interface {
	Next(addr net.Addr) *eventloop
	Register(e *eventloop)
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
		return &eventLoopGroup{loadBalance: &roundRobin{0}}
	case Random:
		return &eventLoopGroup{loadBalance: &random{}}
	case RoundRobin:
		return &eventLoopGroup{loadBalance: &leastConnections{}}
	default:
		return &eventLoopGroup{loadBalance: &roundRobin{0}}
	}
}

func (g *eventLoopGroup) Next(addr net.Addr) *eventloop {
	return g.loadBalance.next(g.loops, addr)
}

func (g *eventLoopGroup) Register(e *eventloop) {
	g.loops = append(g.loops, e)
}

type LoadBalance int

const (
	// RoundRobin 轮询法，即将请求按照顺序轮流的分配到服务器上，均衡的对待每一台后端的服务器,不关心服务器的的连接数和负载情况
	RoundRobin LoadBalance = iota

	// LeastConnections 根据当前的连接情况，动态的选取其中当前积压连接数最少的一台服务器来处理当前请求
	// 尽可能的提高后台服务器利用率，将负载合理的分流到每一台服务器。
	LeastConnections

	// Random 根据服务器列表的大小来随机获取其中的一台来访问，随着调用量的增大，实际效果越来越近似于平均分配到没一台服务器，和轮询的效果类似
	Random
)

// RoundRobin
type roundRobin struct {
	idx int
}

func (r *roundRobin) next(loops []*eventloop, addr net.Addr) (e *eventloop) {
	e = loops[r.idx]
	r.idx = (r.idx + 1) % len(loops)
	return
}

// LeastConnections
type leastConnections struct {
}

func (l *leastConnections) next(loops []*eventloop, addr net.Addr) (e *eventloop) {
	return
}

// Random
type random struct {
	idx int
}

func (r *random) next(loops []*eventloop, addr net.Addr) (e *eventloop) {
	rand.Seed(time.Now().UnixNano())
	return loops[rand.Intn(len(loops))]
}
