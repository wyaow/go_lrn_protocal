package main

import (
	"GnetLrn/log"
	"fmt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"golang.org/x/net/ipv4"
	"net"
	"time"
)

var seq uint32 = 1001
var srcIP = net.ParseIP("127.0.0.1")
var dstIP = net.ParseIP("127.0.0.1")

func main() {
	//"github.com/google/gopacket/examples/util"
	//defer util.Run()()
	SerilizeToUdp()
	return
}

func dial() {
	srcIP := net.ParseIP("127.0.0.1")
	dstIP := net.ParseIP("127.0.0.1")
	conn, err := net.Dial("ip4:tcp", "127.0.0.1")
	if err != nil {
		log.Fatalf("dial err:%v", err)
	}

	ipConn, ok := conn.(*net.IPConn)
	if !ok {
		log.Fatalf("convert not ok")
	}
	ipRawConn, err := ipv4.NewRawConn(ipConn)
	if err != nil {
		log.Fatalf("convert ipRawConn not ok")
	}

	ipLayer := layers.IPv4{
		Version:  4,
		TTL:      64,
		SrcIP:    srcIP,
		DstIP:    dstIP,
		Protocol: layers.IPProtocolUDP, //IPProtocolTCP, IPProtocolUDP...
	}

	// 设置TCP头
	//transLayer := layers.TCP{
	//	SrcPort: layers.TCPPort(9998),
	//	DstPort: layers.TCPPort(9999),
	//	Seq:     1001,
	//	SYN:     true,
	//}
	// 也可以发UDP
	transLayer := layers.UDP{
		SrcPort: layers.UDPPort(9998),
		DstPort: layers.UDPPort(9999),
	}
	if err := transLayer.SetNetworkLayerForChecksum(&ipLayer); err != nil {
		log.Errorf("set layer checksum error, %v", err)
		return
	}
	buf := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{
		FixLengths:       true,
		ComputeChecksums: true,
	}

	// 报文的内容
	// 写点内容, 可以用应用层的 ApplicationLayer的两个实现，
	// Fragment
	// fragment := gopacket.Fragment([]byte("hello")); &fragment
	err = gopacket.SerializeLayers(buf, opts, &ipLayer, &transLayer, gopacket.Payload([]byte("hello")))
	if err != nil {
		log.Errorf("serialize error, %v", err)
		return
	}

	log.Debugf("len:%v data:%v", len(buf.Bytes()), buf.Bytes())
	n, err := ipRawConn.Write(buf.Bytes())
	if n != len(buf.Bytes()) {
		log.Errorf("write n:%v less than buf size:%v, err:%v", n, len(buf.Bytes()), err)
	}
	return
}

func listenSend() {
	srcIP = net.ParseIP("127.0.0.1")
	//srcIPaddr := net.IPAddr{
	//	IP: srcIP,
	//}
	dstIPaddr := net.IPAddr{
		IP: dstIP,
	}
	srcIP = net.ParseIP("127.0.0.1")
	ipLayer := layers.IPv4{
		Version:  4,
		TTL:      64,
		SrcIP:    srcIP,
		DstIP:    dstIP,
		Protocol: layers.IPProtocolTCP,
	}

	// 设置TCP头
	tcpLayer := &layers.TCP{
		SrcPort: layers.TCPPort(9998),
		DstPort: layers.TCPPort(9999),
		SYN:     true,
		Seq:     1005,
		//FIN:     true,
	}
	err1 := tcpLayer.SetNetworkLayerForChecksum(&ipLayer)
	if err1 != nil {
		log.Errorf("SetNetworkLayerForChecksum err:%v", tcpLayer.Checksum)
	}

	buf := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{FixLengths: true, ComputeChecksums: true}
	err := gopacket.SerializeLayers(buf, opts, &ipLayer, tcpLayer)
	if err != nil {
		log.Fatal(err)
	}

	// LinkTypeMetadata[LinkTypeRaw] = EnumMetadata{DecodeWith: gopacket.DecodeFunc(decodeIPv4or6), Name: "Raw"}
	p := gopacket.NewPacket(buf.Bytes(), layers.LinkTypeRaw, gopacket.Default)
	if p.ErrorLayer() != nil {
		log.Fatal("Failed to decode packet:", p.ErrorLayer().Error())
	}
	layersCount := p.Layers()
	log.Infof("layers count:%v", len(layersCount))

	l, ok := p.Layer(layers.LayerTypeTCP).(*layers.TCP)
	if !ok {
		log.Fatal("No tcp layer type found in packet")
	}
	log.Infof("l:%+v", l)
	got := l.Checksum
	if got != tcpLayer.Checksum {
		log.Errorf("Bad checksum:got:%#v, want:%#v", got, tcpLayer.Checksum)
	}

	// ListenPacket创建的使用, 因为没指定目的地址: ipConn.WriteTo(buf.Bytes(), &dstIPaddr)；write: destination address required
	// Dial创建是使用，拨号指定过了目的地址：raw.Write(buf.Bytes())；use of WriteTo with pre-connected connection
	// 已经建立过连接则不能使用该接口: 用这个可以：net.ListenPacket("ip4:tcp", "127.0.0.1")
	// ipConn, err := net.DialIP("ip4:tcp", &srcIPaddr, &dstIPaddr)
	// if err != nil {
	//	 log.Fatalf("listen packet err:%v", err)
	// }
	ipConn, err := net.ListenPacket("ip4:tcp", "127.0.0.1")
	if err != nil {
		log.Fatalf("listen packet err:%v", err)
	}
	raw, err := ipv4.NewRawConn(ipConn)
	if err != nil {
		log.Fatalf("NewRawConn err:%v", err)
	}

	go func(rawConn *ipv4.RawConn) {
		for {
			buf := make([]byte, 4094)

			// 这个读取到的是IP上层 的报文,
			// func (c *IPConn) ReadFromIP(b []byte) (int, *IPAddr, error) {
			// func (c *packetHandler) ReadFrom(b []byte) (h *Header, p []byte, cm *ControlMessage, err error) {
			// 原始报文
			// func (c *conn) Read(b []byte) (int, error) {
			h, payload, _, err := rawConn.ReadFrom(buf)
			if len(payload) < 0 || err != nil {
				log.Errorf("read raw conn from [%s] fail, err:%v", h.Src.String(), err)
				return
			}
			if h.Src.String() != "127.0.0.1" {
				continue
			}

			log.Infof("addr src:%v, dst:%v", h.Src, h.Dst)
			log.Infof("p:%v", payload)
			p := gopacket.NewPacket(payload, layers.LayerTypeTCP, gopacket.Default)
			if p.ErrorLayer() != nil {
				log.Fatal("Failed to decode LinkTypeRaw packet:", p.ErrorLayer().Error())
			}

			// checkLayers(p, []gopacket.LayerType{LayerTypeIPv6, LayerTypeIPv6Destination, LayerTypeUDP}, t)
			var tcp *layers.TCP
			if l, ok := p.Layer(layers.LayerTypeTCP).(*layers.TCP); !ok {
				log.Fatal("No UDP layer type found in packet")
			} else {
				tcp = l
			}

			log.Infof("收到TCP报文:%+v", tcp)
		}
	}(raw)

	time.Sleep(time.Second * 1)
	n, err := raw.WriteToIP(buf.Bytes(), &dstIPaddr)
	if n < len(buf.Bytes()) {
		log.Fatalf("write expect:%v, actual:%v, err:%v", len(buf.Bytes()), n, err)
	}
	if err != nil {
		log.Fatalf("write to error:%v", err)
	}
	log.Infof("ip checksum:%v %x", ipLayer.Checksum, ipLayer.Checksum)
	log.Infof("tcp checksum:%v %x", tcpLayer.Checksum, tcpLayer.Checksum)
	log.Infof("send buf:%v", buf.Bytes())
	select {}
}

func SerilizeToUdp() {
	transLayer := layers.UDP{
		SrcPort: layers.UDPPort(9998),
		DstPort: layers.UDPPort(9999),
	}
	opts := gopacket.SerializeOptions{FixLengths: true, ComputeChecksums: true}

	buf := gopacket.NewSerializeBuffer()
	// transLayer.SerializeTo(buf, opts)	// 如果不需要payload
	gopacket.SerializeLayers(buf, opts, &transLayer, gopacket.Payload("hello"))
	fmt.Println(buf)
	// decode
	var udpLayers *layers.UDP

	// Decode an ethernet packet
	// ethP := gopacket.NewPacket(p1, layers.LayerTypeEthernet, gopacket.Default)
	// Decode an IPv6 header and everything it contains
	// ipP := gopacket.NewPacket(p2, layers.LayerTypeIPv6, gopacket.Default)
	// Decode a TCP header and its payload
	// tcpP := gopacket.NewPacket(p3, layers.LayerTypeTCP, gopacket.Default)
	//
	// tcp4or6P := gopacket.NewPacket(p3, layers.LinkTypeRaw, gopacket.Default)
	// LinkTypeMetadata[LinkTypeRaw] = EnumMetadata{DecodeWith: gopacket.DecodeFunc(decodeIPv4or6), Name: "Raw"}
	p := gopacket.NewPacket(buf.Bytes(), layers.LayerTypeUDP, gopacket.Default)
	if p.ErrorLayer() != nil {
		log.Fatal("Failed to decode packet:", p.ErrorLayer().Error())
	}
	// checkLayers(p, []gopacket.LayerType{LayerTypeIPv6, LayerTypeIPv6Destination, LayerTypeUDP}, t)
	if l, ok := p.Layer(layers.LayerTypeUDP).(*layers.UDP); !ok {
		log.Fatal("No UDP layer type found in packet")
	} else {
		udpLayers = l
	}
	fmt.Println(udpLayers.Length)
}
