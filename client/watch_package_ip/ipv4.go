package main

import (
	"GnetLrn/log"
	"fmt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"golang.org/x/net/ipv4"
	"net"
	"os"
	"syscall"
)

// window下可以用这个解决, 但可能仍然读取空的
// go 最终也是通过这个系统调用：func Socket(domain, typ, proto int) (fd Handle, err error) {
func mainWindows() {
	//syscall.AF_INET，表示服务器之间的网络通信
	//syscall.AF_UNIX表示同一台机器上的进程通信
	//syscall.AF_INET6表示以IPv6的方式进行服务器之间的网络通信
	//其他
	fd, _ := syscall.Socket(syscall.AF_INET, syscall.SOCK_RAW, syscall.IPPROTO_TCP)
	f := os.NewFile(uintptr(fd), fmt.Sprintf("fd %d", fd))
	for {
		buf := make([]byte, 1500)
		n, _ := f.Read(buf)
		log.Debug(n)
		ip4header, _ := ipv4.ParseHeader(buf[:20])
		log.Debugf("ipheader:%+v", ip4header)
		if n > 20 {
			log.Debugf("body:%v", buf[20:n])
		}
		// tcpheader := util.NewTCPHeader(buf[20:40])
		//log.Debugf("tcpheader:", tcpheader)
	}
}

func main() {
	// 返回的: PacketConn, 主要实现接口有以下:
	// IPConn, UdpConn, UnixConn, Icmp(PacketConn)
	c, err := net.ListenPacket("ip4:tcp", "0.0.0.0")
	if err != nil {
		log.Fatal(err)
	}

	rawConn, err := ipv4.NewRawConn(c)
	if err != nil {
		log.Fatal(err)
	}
	defer rawConn.Close()

	// 开启无线循环获取报文
	for {
		buf := make([]byte, 2048)
		// 获取报文同时解析IP报文，得到IP头和IP头后面部分（传输层和应用层）
		header, payload, cm, err := rawConn.ReadFrom(buf)
		if err != nil {
			log.Debugf("解析IP头错误:%v", err)
			return
		}

		if len(payload) == 40 {
			// 打印IP头
			log.Debugf("header:%+v", *header)
			log.Debugf("cm:%#v", cm)
			log.Debugf("payload:%v", payload)
			log.Debugf("string(payload):%s", string(payload))

			// 解析TCP
			log.Debugf("parse tcp...")
			p := gopacket.NewPacket(payload, layers.LayerTypeIPv4, gopacket.Default)
			if p.ErrorLayer() != nil {
				log.Errorf("Failed to decode packet:", p.ErrorLayer().Error())
			}
			tcp := p.Layer(layers.LayerTypeTCP).(*layers.TCP)
			if tcp == nil {
				log.Errorf("Expected TCP layer, but got none")
			}
			log.Debugf("tcp:%+v", tcp)
			log.Debugf("checksum:%v", tcp.Checksum)
		}
	}
}

// ListenIP 同样可以在IP层
// 返回的: PacketConn, 主要实现接口有以下:
// IPConn, UdpConn, UnixConn, Icmp(PacketConn)
//addr, err := net.ResolveIPAddr("ip4", "0.0.0.0")
//c, err := net.ListenIP("ip4:tcp", addr)
//rawConn, err := ipv4.NewRawConn(c)
// IPConn, 相当于可以处理IP层协议的内容, 但是无法感知IP头等
//func main() {
//	netaddr, _ := net.ResolveIPAddr("ip4", "172.17.0.3")
//	conn, _ := net.ListenIP("ip4:icmp", netaddr)
//	for {
//		buf := make([]byte, 1024)
//		n, addr, _ := conn.ReadFrom(buf)
//		msg,_:=icmp.ParseMessage(1,buf[0:n])
//		fmt.Println(n, addr, msg.Type,msg.Code,msg.Checksum)
//	}
//}
func rawMain() {
	c, err := net.ListenPacket("ip4:tcp", "0.0.0.0")
	if err != nil {
		log.Fatal(err)
	}

	// 只由IPConn才能获取raw conn
	// *net.UDPConn, not *net.IPConn
	rawConn, err := ipv4.NewRawConn(c)
	if err != nil {
		log.Fatal(err)
	}

	// NewPacketConn 数据报
	// ipv4.NewConn()
	// p := ipv4.NewPacketConn(c)
	// IPv4 transport
	// 控制信息
	//rawConn := ipv4.NewPacketConn(c)
	//_, cm, src, err := rawConn.ReadFrom(buf)

	defer rawConn.Close()
	// 开启无线循环获取报文
	for {
		buf := make([]byte, 2048)
		// 获取报文同时解析IP报文，得到IP头和IP头后面部分（传输层和应用层）
		header, payload, cm, err := rawConn.ReadFrom(buf)
		if err != nil {
			log.Debugf("解析IP头错误:%v", err)
			return
		}
		// 打印IP头
		log.Debugf("header:%+v", *header)
		log.Debugf("cm:%#v", cm)
		log.Debugf("payload:%v", payload)
		log.Debugf("string(payload):%s", string(payload))
	}
}
