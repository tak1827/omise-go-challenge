package cipher

import (
	"bytes"
	"io"
	"testing"

	r "github.com/stretchr/testify/require"
)

var (
	TestBuffer        = []byte{128, 129, 130}
	ReverseTestBuffer = []byte{0, 1, 2}
)

type MockReader struct {
	r io.Reader
}

func (r *MockReader) Read(b []byte) (n int, err error) {
	return r.r.Read(b)
}

func (r *MockReader) ReadAt(b []byte, off int64) (n int, err error) {
	return r.r.Read(b)
}

func TestRot128Reader_Read(t *testing.T) {
	reader, err := NewRot128Reader(&MockReader{r: bytes.NewBuffer(TestBuffer)}, nil)
	r.NoError(t, err)
	r.NotNil(t, reader)

	buf := make([]byte, 3, 3)
	n, err := reader.ReadAt(buf, 0)
	r.NoError(t, err)
	r.Equal(t, 3, n)
	r.Equal(t, ReverseTestBuffer, buf)
}

func TestRot128Reader_Reversible(t *testing.T) {
	reader, err := NewRot128Reader(&MockReader{r: bytes.NewBuffer(TestBuffer)}, nil)
	r.NoError(t, err)
	r.NotNil(t, reader)

	reader, err = NewRot128Reader(reader, nil)
	r.NoError(t, err)
	r.NotNil(t, reader)

	buf := make([]byte, 3, 3)
	n, err := reader.ReadAt(buf, 0)
	r.NoError(t, err)
	r.Equal(t, 3, n)
	r.Equal(t, TestBuffer, buf)
}

func TestRot128Writer_Write(t *testing.T) {
	buf := &bytes.Buffer{}
	writer, err := NewRot128Writer(buf)
	r.NoError(t, err)
	r.NotNil(t, writer)

	n, err := writer.Write(TestBuffer)
	r.NoError(t, err)
	r.Equal(t, 3, n)
	r.Equal(t, ReverseTestBuffer, buf.Bytes())
}

func TestRot128Writer_Reversible(t *testing.T) {
	buf := &bytes.Buffer{}
	writer, err := NewRot128Writer(buf)
	r.NoError(t, err)
	r.NotNil(t, writer)

	writer, err = NewRot128Writer(writer)
	r.NoError(t, err)
	r.NotNil(t, writer)

	n, err := writer.Write(TestBuffer)
	r.NoError(t, err)
	r.Equal(t, 3, n)
	r.Equal(t, TestBuffer, buf.Bytes())
}
