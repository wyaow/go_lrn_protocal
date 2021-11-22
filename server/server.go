package server

import (
	codec2 "GnetLrn/codec"
	"GnetLrn/log"
	"strings"
	"time"

	"github.com/panjf2000/gnet"
	"github.com/panjf2000/gnet/pool/goroutine"
)

type Server = server

type server struct {
	*gnet.EventServer
	addr       string
	multiCore  bool
	asyncWrite bool
	codec      gnet.ICodec
	workerPool *goroutine.Pool
}

// 创建Server, 使用默认编解码
// 地址格式类型: "tcp://ip:port"
// 地址格式类型: "udp://ip:port"
// like `tcp://x.x.x.x:9851` or `unix://socket`.
func NewServer(addr string) *server {
	server := &server{
		addr:       addr,
		multiCore:  true,
		asyncWrite: false,
		codec:      codec2.NewCustomCodec(), // 使用自定义编解码
		workerPool: goroutine.Default(),
	}

	return server
}

func (s *server) Serve() (err error) {
	opts := []gnet.Option{
		gnet.WithCodec(s.codec),
		gnet.WithTicker(true),
		gnet.WithReusePort(false),
		gnet.WithLogger(log.Logger),
		gnet.WithMulticore(s.multiCore),
	}

	if strings.HasPrefix(s.addr, "tcp") {
		opts = append(opts, gnet.WithTCPKeepAlive(time.Minute*5))
	}

	return gnet.Serve(s, s.addr, opts...)
}

func (s *server) OnInitComplete(srv gnet.Server) (action gnet.Action) {
	log.Infof("server is listening on %s addr:%s [multi-core:%t,loops:%d,reuse-port:%t,tcp-keep-alive:%v]",
		srv.Addr.Network(), srv.Addr.String(), srv.Multicore, srv.NumEventLoop, srv.ReusePort, srv.TCPKeepAlive)
	return
}

// OnOpened fires when a new connection has been opened.
// The parameter:c has information about the connection such as it's local and remote address.
// Parameter:out is the return value which is going to be sent back to the client.
func (s *server) OnOpened(c gnet.Conn) (out []byte, action gnet.Action) {
	log.Infof("remote addr: [%s] is connected", c.RemoteAddr())
	return
}

// OnShutdown fires when the server is being shut down, it is called right after
// all event-loops and connections are closed.
func (s *server) OnShutdown(svr gnet.Server) {
	log.Infof("stopped serving at addr: [%s]", s.addr)
}

// OnClosed fires when a connection has been closed.
// The parameter:err is the last known connection error.
func (s *server) OnClosed(c gnet.Conn, err error) (action gnet.Action) {
	log.Infof("remote addr: [%s] is disconnected", c.RemoteAddr())
	return
}

// React fires when a connection sends the server data.
// Call c.Read() or c.ReadN(n) within the parameter:c to read incoming data from client.
// Parameter:out is the return value which is going to be sent back to the client.
func (s *server) React(frame []byte, c gnet.Conn) (out []byte, action gnet.Action) {
	log.Debugf("server recv len:%v, frame:%v", len(frame), frame)
	var err error
	// SetContext sets a user-defined context.
	// store customize protocol header param using `c.SetContext()`
	// item := protocol.CustomLengthFieldProtocol{Version: protocol.DefaultProtocolVersion, ActionType: protocol.ActionData}
	// c.SetContext(item)

	// 是否开启异步写, 会Encode
	if s.asyncWrite {
		data := append([]byte{}, frame...)
		err := s.workerPool.Submit(func() {
			if err := c.AsyncWrite(data); err != nil {
				log.Errorf("async write data err:%v", err)
			}
		})
		if err != nil {
			log.Errorf("work pool submit task err:%v", err)
		}

		return
	}

	// 要发送的数据
	out = []byte("response from server")

	// 如果是udp需要自行编码, SendTo一次发送一个包
	if !strings.HasPrefix(s.addr, "tcp") {
		out, err = s.codec.Encode(c, out)
		if err != nil {
			log.Errorf("before send packet, encode packet err, %v", err)
			return
		}
		_ = s.workerPool.Submit(
			func() {
				log.Debugf("send to data:%v", out)
				if err := c.SendTo(out); err != nil {
					log.Errorf("send packet, call SendTo err, %v", err)
				}
			})
		return
	}

	return
}

// PreWrite fires just before any data is written to any client socket, this event function is usually used to
// put some code of logging/counting/reporting or any prepositive operations before writing data to client.
func (s *server) PreWrite() {
	log.Debugf("begin to write data...")
	return
}

// 周期性的
// Tick fires immediate
