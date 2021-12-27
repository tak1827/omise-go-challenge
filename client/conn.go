package client

import (
	"bufio"
	"context"
	"crypto/tls"
	"io"
	"net"
)

type persistConn struct {
	io.Closer

	conn net.Conn
	br   *bufio.Reader
	bw   *bufio.Writer
}

var (
	_ io.Writer = (*persistConn)(nil)
	_ io.Reader = (*persistConn)(nil)
)

func NewConn(ctx context.Context, host, port string) (pc *persistConn, err error) {
	pc = &persistConn{}

	pc.conn, err = net.Dial("tcp", host+":"+port)
	if err != nil {
		return
	}

	if port == "443" {
		errCh := make(chan error, 1)
		tlsConn := tls.Client(pc.conn, &tls.Config{
			MinVersion: tls.VersionTLS12,
			ServerName: host,
		})
		go func() {
			err := tlsConn.HandshakeContext(ctx)
			errCh <- err
		}()
		pc.conn = tlsConn

		if err = <-errCh; err != nil {
			return
		}
		pc.conn = tlsConn
	}

	pc.br = bufio.NewReaderSize(pc, 1024)
	pc.bw = bufio.NewWriterSize(pc, 1024)
	return
}

func (pc *persistConn) Write(p []byte) (n int, err error) {
	return pc.conn.Write(p)
}

func (pc *persistConn) Read(p []byte) (n int, err error) {
	if n, err = pc.conn.Read(p); err != nil {
		if err == io.EOF {
			return
		}
		return n, err
	}
	return
}

func (pc *persistConn) Close() {
	pc.conn.Close()
}
