package client

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/omise/omise-go"
)

type Client struct {
	pkey string
	skey string

	OmiseClient omise.Client

	conns map[string]*persistConn
}

var (
	ErrRateLimit = errors.New("api rate limit")

	schemas = map[string]string{
		"http":  "80",
		"https": "443",
	}
)

func NewClient(pkey, skey string) (Client, error) {
	if pkey == "" || skey == "" {
		return Client{}, omise.ErrInvalidKey
	}

	return Client{
		pkey:        pkey,
		skey:        skey,
		OmiseClient: omise.Client{},
		conns:       make(map[string]*persistConn, 2),
	}, nil
}

func (c *Client) Do(ctx context.Context, req *http.Request, result interface{}) error {
	defer req.Body.Close()

	var (
		host = req.URL.Hostname()
		port = req.URL.Port()
	)
	if _, ok := c.conns[host]; !ok || c.conns[host] == nil {
		conn, err := NewConn(ctx, host, port)
		if err != nil {
			return err
		}
		c.conns[host] = conn
	}

	errCh := make(chan error, 1)

	go func() {
		err := req.Write(c.conns[host].bw)
		c.conns[host].bw.Flush()
		errCh <- err
	}()

	if err := <-errCh; err != nil {
		return err
	}

	resp, err := http.ReadResponse(c.conns[host].br, req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	buffer, err := io.ReadAll(resp.Body)
	if err != nil {
		err = &omise.ErrTransport{Err: err, Buffer: buffer}
	}

	switch {
	case resp.StatusCode == 429:
		return ErrRateLimit
	case resp.StatusCode != 200:
		oErr := &omise.Error{StatusCode: resp.StatusCode}
		err = json.Unmarshal(buffer, oErr)
		return err
	}

	return json.Unmarshal(buffer, result)
}

func (c *Client) Close() {
	for key, val := range c.conns {
		if val != nil {
			c.conns[key].Close()
		}
	}
}
