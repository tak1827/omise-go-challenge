package cipher

import (
	"io"
)

type Rot128Handler func(b byte)

var DefaultRot128Handler = func(b byte) {}

// Rot128Reader implements io.Reader that transforms
type Rot128Reader struct {
	reader  Reader
	handler Rot128Handler
}

func NewRot128Reader(r Reader, h Rot128Handler) (*Rot128Reader, error) {
	if h == nil {
		h = DefaultRot128Handler
	}
	return &Rot128Reader{
		reader:  r,
		handler: h,
	}, nil
}

func (r *Rot128Reader) ReadAt(p []byte, off int64) (int, error) {
	if n, err := r.reader.ReadAt(p, off); err != nil {
		return n, err
	} else {
		r.rot128(p[:n])
		return n, nil
	}
}

func (r *Rot128Reader) Read(p []byte) (int, error) {
	return r.ReadAt(p, 0)
}

func (r *Rot128Reader) rot128(buf []byte) {
	for idx := range buf {
		buf[idx] += 128
		r.handler(buf[idx])
	}
}

type Rot128Writer struct {
	writer io.Writer
	buffer []byte // not thread-safe
}

func NewRot128Writer(w io.Writer) (*Rot128Writer, error) {
	return &Rot128Writer{
		writer: w,
		buffer: make([]byte, 4096, 4096),
	}, nil
}

func (w *Rot128Writer) Write(p []byte) (int, error) {
	n := copy(w.buffer, p)
	rot128(w.buffer[:n])
	return w.writer.Write(w.buffer[:n])
}

func rot128(buf []byte) {
	for idx := range buf {
		buf[idx] += 128
	}
}
