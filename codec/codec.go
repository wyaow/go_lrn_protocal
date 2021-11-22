// Package customCodec
// @Description: 此编解码基于 LengthFieldBasedFrameCodec
// more: https://github.com/smallnest/goframe
// netty codec quick use: https://netty.io/4.1/api/io/netty/handler/codec/LengthFieldBasedFrameDecoder.html
package codec

import (
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/panjf2000/gnet"
)

const (
	DefaultHeadLength uint32 = 24         // 头部长度(除内容外长度)
	MagicNum          uint32 = 0x80635314 // 魔数

	// UDP最大内容长度
	MaxUdpPacketSize uint32 = 2<<16 - 1 - 8 - 20
)

/*
The following message is a simple variation of the first example. An extra header value is
prepended to the message. lengthAdjustment is zero again because the decoder always takes
the length of the prepended data into account during frame length calculation.
 lengthFieldOffset   = 0
 lengthFieldLength   = 4
 lengthAdjustment    = 20 (= the length of Header 1)
 initialBytesToStrip = 0

 BEFORE DECODE (18 bytes)                           AFTER DECODE (18 bytes)
 +------------+------------+----------------+      +------------+------------+----------------+
 |    Length  |   Header 1 | Actual Content |----->|    Length  |   Header 1 | Actual Content |
 | 0x0000000C |  16byte+4  | "HELLO, WORLD" |      | 0x0000000C |  16byte +4 | "HELLO, WORLD" |
 +------------+------------+----------------+      +------------+-----------+-----------------+

这里采取LengthFieldBasedFrameCodec 作为基础编解码器，对其进行封装
1.其中Length 采用4字节，最大2^32-1，支持G级别
2.Header1为md5对应checksum, 16字节, 魔数4字节
3.Content为内容
*/
type customCodec struct {
	ec    gnet.EncoderConfig
	dc    gnet.DecoderConfig
	codec *gnet.LengthFieldBasedFrameCodec
}

func NewCustomCodec() *customCodec {
	customCodec := &customCodec{}
	customCodec.ec = gnet.EncoderConfig{
		ByteOrder:                       binary.BigEndian,
		LengthFieldLength:               4,
		LengthAdjustment:                -20,
		LengthIncludesLengthFieldLength: false,
	}
	customCodec.dc = gnet.DecoderConfig{
		ByteOrder:           binary.BigEndian,
		LengthFieldOffset:   0,
		LengthFieldLength:   4,
		LengthAdjustment:    20,
		InitialBytesToStrip: 0,
	}

	customCodec.codec = gnet.NewLengthFieldBasedFrameCodec(customCodec.ec, customCodec.dc)
	return customCodec
}

// Encode ...
func (cc *customCodec) Encode(c gnet.Conn, buf []byte) (out []byte, err error) {
	if cc.codec == nil {
		return nil, errors.New("encode codec is not init")
	}

	// 加上16字节checkSum
	newBuf := make([]byte, 0, int(DefaultHeadLength)+len(buf))
	sum := md5.Sum(buf)
	newBuf = append(newBuf, sum[:]...)

	// 加上4字节的magicNum
	magic := make([]byte, 4)
	cc.ec.ByteOrder.PutUint32(magic, MagicNum)
	newBuf = append(newBuf, magic...)

	newBuf = append(newBuf, buf...)
	return cc.codec.Encode(c, newBuf)
}

// Decode ...
func (cc *customCodec) Decode(c gnet.Conn) ([]byte, error) {
	if cc.codec == nil {
		return nil, errors.New("decode codec is not init")
	}

	// after decode, Conn will be shifted
	buf, err := cc.codec.Decode(c)
	if err != nil {
		return buf, fmt.Errorf("inner decode err:%v", err)
	}
	byteBuffer := bytes.NewBuffer(buf)

	var dataLength uint32
	var expectSum [md5.Size]byte
	var gotMagic uint32
	_ = binary.Read(byteBuffer, binary.BigEndian, &dataLength)
	_ = binary.Read(byteBuffer, binary.BigEndian, &expectSum)
	_ = binary.Read(byteBuffer, binary.BigEndian, &gotMagic)

	// 计算md5比较
	checkSum := md5.Sum(byteBuffer.Bytes())
	if expectSum != checkSum {
		return nil, fmt.Errorf("md5 check sum fail, expected:%s, got:%s", expectSum, checkSum)
	}

	// 比较魔数
	if gotMagic != MagicNum {
		return nil, fmt.Errorf("magicnum check fail, expected:%x, got:%x", MagicNum, gotMagic)
	}

	return byteBuffer.Bytes(), nil
}
