package main

import (
	"GnetLrn/codec"
	"GnetLrn/log"
	"net"
	"time"
)

// Example command: go run client.go
func main2() {
	conn, err := net.Dial("tcp", "127.0.0.1:9999")
	if err != nil {
		log.Fatalf("net dial error, %v", err)
	}
	defer conn.Close()

	// 接受服务端返回
	go func() {
		for {
			data := []byte("hello world")
			buf, err := codec.ClientEncode(data)
			if err != nil {
				log.Errorf("client buf data error, %v", err)
				return
			}
			n, err := conn.Write(buf)
			if err != nil {
				log.Errorf("write data error, n:%v, err:%v", n, err)
				return
			}
			log.Debugf("client write len:%v, data:%v", len(data), buf)

			time.Sleep(time.Second * 5)
		}
	}()

	for {
		response, err := codec.ClientDecode(conn)
		if err != nil {
			log.Errorf("client decode error, %v", err)
		}
		log.Debugf("client received data:%s", string(response))
		time.Sleep(time.Second * 1)
	}

	select {}
}
