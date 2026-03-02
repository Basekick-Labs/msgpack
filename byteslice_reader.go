package msgpack

import (
	"errors"
	"io"
)

// byteSliceReader implements bufReader (io.Reader + io.ByteScanner) for
// zero-allocation decoding from a []byte. When Unmarshal is called with
// a complete byte slice, this avoids the overhead of bytes.NewReader +
// bufio.NewReader (2 allocations + 4KB buffer + interface dispatch per byte).
type byteSliceReader struct {
	data []byte
	pos  int
}

func (r *byteSliceReader) reset(data []byte) {
	r.data = data
	r.pos = 0
}

func (r *byteSliceReader) ReadByte() (byte, error) {
	if r.pos >= len(r.data) {
		return 0, io.EOF
	}
	c := r.data[r.pos]
	r.pos++
	return c, nil
}

func (r *byteSliceReader) UnreadByte() error {
	if r.pos <= 0 {
		return errors.New("msgpack: at beginning of input")
	}
	r.pos--
	return nil
}

func (r *byteSliceReader) Read(p []byte) (int, error) {
	if r.pos >= len(r.data) {
		return 0, io.EOF
	}
	n := copy(p, r.data[r.pos:])
	r.pos += n
	return n, nil
}
