package msgpack

// byteSliceWriter is a zero-allocation writer that appends directly to a
// []byte buffer. It implements the writer interface (io.Writer + WriteByte).
// Used by Marshal() to avoid allocating a bytes.Buffer on every call.
type byteSliceWriter struct {
	buf *[]byte
}

func (w *byteSliceWriter) Write(p []byte) (int, error) {
	*w.buf = append(*w.buf, p...)
	return len(p), nil
}

func (w *byteSliceWriter) WriteByte(c byte) error {
	*w.buf = append(*w.buf, c)
	return nil
}
