package buf

import (
	"errors"
	"io"
)

// PacketReader is a Reader that read one Buffer every time.
type PacketReader struct {
	io.Reader
}

// NewPacketReader creates a new PacketReader based on the given reader.
func NewPacketReader(reader io.Reader) *PacketReader {
	return &PacketReader{
		Reader: reader,
	}
}

// ReadMultiBuffer implements Reader.
func (r *PacketReader) ReadUdp() (*Buffer, error) {
	b, err := readOneUDP(r.Reader)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func readOneUDP(r io.Reader) (*Buffer, error) {
	b := New()
	_, err := b.ReadFrom(r)
	if !b.IsEmpty() {
		return b, nil
	}

	// 失败，释放
	defer b.Release()
	if err != nil {
		return nil, err
	}

	return nil, errors.New("reader returns too many empty payloads")
}
