// Package main
// @Description: 在9999端口启动TCP和UDP服务, UDP用于控制面通信
package main

import (
	"GnetLrn/log"
	"GnetLrn/server"
	"flag"
	"fmt"
)

// 命令行参数
var (
	port            int    // 默认起tcp, udp同一端口
	controlProtocol string // 与控制面通信协议

	allServer = make(map[string]*server.Server, 2)
)

// Example command: go run main.go --server --addr tcp://0.0.0.0:9999
func main() {
	flag.IntVar(&port, "port", 9999, "--port 9999")
	flag.StringVar(&controlProtocol, "controlProtocol", "tcp", "--controlProtocol [tcp/udp]")
	tcpAddr := fmt.Sprintf("tcp://0.0.0.0:%v", port)
	udpAddr := fmt.Sprintf("udp://0.0.0.0:%v", port)
	allServer["tcp"] = server.NewServer(tcpAddr)
	allServer["udp"] = server.NewServer(udpAddr)

	// 先启动一个服务, 用于业务通信
	go func() {
		s := getServer(controlProtocol, false)
		if s == nil {
			log.Fatalf("get a nil server controlProtocol:%s, match:false", controlProtocol)
			return
		}
		log.Fatal(s.Serve())
	}()

	// 启动服务
	ss := getServer(controlProtocol, true)
	if ss == nil {
		log.Fatalf("get a nil server controlProtocol:%s, match:false", controlProtocol)
		return
	}
	log.Fatal(ss.Serve())
}

// 找到是否匹配对应协议的服务
// 暂不处理其他协议，认为无错误
func getServer(protocol string, match bool) *server.Server {
	if match {
		return allServer[protocol]
	}

	// 找一个不是该协议的
	for p, s := range allServer {
		if p != protocol {
			return s
		}
	}

	return nil
}
