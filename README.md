## jinx 简介

IO 多路复用 + Reactor 模式实现的网络库
> 练手项目，参考学习了许多开源项目：gnet，nbio，evio，netty，gev...

值得看一看的 issue [异步的优势 #4](https://github.com/Allenxuxu/gev/issues/4)

## API

**目前只支持 Linux2.6 以上系统，且只支持 TCP 协议**

```golang
// handler event callback
type handler interface {

    // OnBoot server 启动
    OnBoot(f func(s Server))

    // OnOpen newConn 连接建立
    OnOpen(f func(c Conn))

    // OnClose closeConn 连接关闭
    OnClose(f func(c Conn))

    // OnRead 可读事件，收到客户端发送的数据
    OnRead(f func(c Conn))

    // OnWrite 可写事件，在服务端发送数据到客户端之前
    OnWrite(f func(c Conn))

    // OnShutdown 服务关闭
    OnShutdown(f func(s Server))
}
```
## 演示 Demo
### echo-server
```go
package main

import (
    . "github.com/imlgw/jinx"
    "log"
)

func main() {
    network := "tcp"
    addr := ":9876"

    server, err := NewServer(network, addr, WithLb(RoundRobin), WithLoopNum(4), WithServerName("Resolmi"))
    if err != nil {
        log.Fatal(err)
        return
    }

    server.OnBoot(func(s Server) {
        log.Printf("\nserver info: \nname: [%s] \nnetwork: [%s] \naddr:[%s]\n",
            s.ServerName(), s.Network(), s.ServerAddr())
    })

    server.OnOpen(func(c Conn) {
        log.Printf("\nnew conn establish \nisOpen: [%v]  \nremoteAddr: [%v]",
            c.IsOpen(), c.RemoteAddr())
    })

    server.OnRead(func(c Conn) {
        buf := make([]byte, 1024)
        readn, err := c.Read(buf)
        if err != nil {
            log.Fatal(err)
        }

        if _, err := c.Write(buf[:readn]); err != nil {
            log.Fatal(err)
        }
    })

    server.OnClose(func(c Conn) {
        log.Printf("\nconn closed \nisOpen: [%v] \nlocalAddr: [%v] \nremoteAddr: [%v]",
            c.IsOpen(), c.LocalAddr(), c.RemoteAddr())
    })

    server.OnShutdown(func(s Server) {
        log.Printf("\nserver shutdown: \nname: [%s] \nnetwork: [%s] \naddr:[%s]",
            s.ServerName(), s.Network(), s.ServerAddr())
    })

    if err := server.Run(); err != nil {
        log.Fatal(err)
    }
}
```

演示环境：Linux VM-20-13-ubuntu 5.4.0-96-generic

![](https://static.imlgw.top/blog/qoch6-mimvl.gif)

## TODO
- [ ] 内部事件 + tricker
- [ ] udp 支持
- [ ] ws支持
- [ ] conn 零拷贝 sendfile 接口
- [ ] buffer 动态扩容
- [ ] windows 支持
