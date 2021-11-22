package codec

import (
	"GnetLrn/common/bytespool"
	"GnetLrn/log"
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
)

var nilMd5Sum = md5.Sum(nil)

// ClientEncode :
func ClientEncode(data []byte) ([]byte, error) {
	result := make([]byte, 0, len(data))
	buffer := bytes.NewBuffer(result)

	// 写数据长度
	dataLen := uint32(len(data))
	if err := binary.Write(buffer, binary.BigEndian, dataLen); err != nil {
		return nil, fmt.Errorf("encode pack datalength error, %v", err)
	}

	// 写md5 checksum
	checksum := md5.Sum(data)
	if err := binary.Write(buffer, binary.BigEndian, checksum); err != nil {
		return nil, fmt.Errorf("encode pack md5 checksum error, %v", err)
	}

	// 写魔数
	if err := binary.Write(buffer, binary.BigEndian, MagicNum); err != nil {
		return nil, fmt.Errorf("encode pack magicnum error, %v", err)
	}

	// 写数据
	if dataLen > 0 {
		if err := binary.Write(buffer, binary.BigEndian, data); err != nil {
			return nil, fmt.Errorf("encode pack data error, %v", err)
		}
	}

	return buffer.Bytes(), nil
}

// ClientDecode :
func ClientDecode(conn net.Conn) ([]byte, error) {
	if conn.RemoteAddr().Network() != "tcp" {
		packetConn, ok := conn.(net.PacketConn)
		if !ok {
			return nil, errors.New("conn convert to PacketConn conn fail")
		}

		return packetDecode(packetConn)
	}

	headData := make([]byte, DefaultHeadLength)
	n, err := io.ReadFull(conn, headData)
	if uint32(n) != DefaultHeadLength {
		return nil, fmt.Errorf("read header len error, expect:%d, got:%d, err:%v", DefaultHeadLength, n, err)
	}

	// parse protocol header
	var dataLength uint32
	var checksum [md5.Size]byte
	var gotMagic uint32
	bytesBuffer := bytes.NewBuffer(headData)
	_ = binary.Read(bytesBuffer, binary.BigEndian, &dataLength)
	_ = binary.Read(bytesBuffer, binary.BigEndian, &checksum)
	_ = binary.Read(bytesBuffer, binary.BigEndian, &gotMagic)

	// check logic
	if gotMagic != MagicNum {
		return nil, fmt.Errorf("read data error, magic expect:%x, got:%x, err:%v", MagicNum, gotMagic, err)
	}

	if dataLength < 1 {
		return nil, nil
	}

	data := make([]byte, dataLength)
	n, err = io.ReadFull(conn, data)
	if uint32(n) != dataLength {
		return nil, fmt.Errorf("read data error, length expect:%v, got:%v, err:%v", dataLength, n, err)
	}

	gotSum := md5.Sum(data)
	if gotSum != checksum {
		return nil, fmt.Errorf("read data error, checksum expect:%v, got:%v, err:%v", checksum, gotSum, err)
	}

	return data, nil
}

// 数据报文，需要一次读完，不能多次读
func packetDecode(packetConn net.PacketConn) ([]byte, error) {
	//headData := make([]byte, DefaultHeadLength)
	buf := bytespool.Alloc(int32(2048))
	n, a, err := packetConn.ReadFrom(buf)
	log.Debugf("begin read packet from remote addr: %s", Addr(a))
	if uint32(n) < DefaultHeadLength {
		return nil, fmt.Errorf("read packet header len:%v less than %v, err:%v", n, DefaultHeadLength, err)
	}
	if err != nil {
		return nil, fmt.Errorf("read packet error, %v", err)
	}

	// parse protocol header
	buffer := bytes.NewBuffer(buf)
	defer buffer.Reset()
	var dataLength uint32
	var checksum [md5.Size]byte
	var gotMagic uint32
	_ = binary.Read(buffer, binary.BigEndian, &dataLength)
	_ = binary.Read(buffer, binary.BigEndian, &checksum)
	_ = binary.Read(buffer, binary.BigEndian, &gotMagic)

	// check logic
	if gotMagic != MagicNum {
		return nil, fmt.Errorf("read packet error, magic expect:%x, got:%x, err:%v", MagicNum, gotMagic, err)
	}

	// 无数据
	if dataLength < 1 {
		if checksum != nilMd5Sum {
			return nil, fmt.Errorf("read packet error, checksum expect:%v, got nil sum:%v", checksum, nilMd5Sum)
		}
		return nil, nil
	}

	// 读取数据部分
	data := make([]byte, dataLength)
	n, err = buffer.Read(data)
	if uint32(n) != dataLength {
		return nil, fmt.Errorf("read packet error, length expect:%v, got:%v, err:%v", dataLength, n, err)
	}
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("read packet error, n:%v, err:%v", n, err)
	}

	gotSum := md5.Sum(data)
	if gotSum != checksum {
		return nil, fmt.Errorf("read packet error, checksum expect:%v, got:%v, err:%v", checksum, gotSum, err)
	}

	return data, nil
}

func Addr(a net.Addr) string {
	if a == nil {
		return ""
	}

	return "[" + a.String() + "]"
}
