package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url" // TODO: replace for good performance
	"strings"

	"github.com/omise/omise-go"
)

type Client struct {
	pkey string
	skey string

	OmiseClient omise.Client

	Conns map[string]*persistConn
}

var schemas = map[string]string{
	"http":  "80",
	"https": "443",
}

func NewClient(pkey, skey string) (*Client, error) {
	switch {
	case pkey == "" && skey == "":
		return nil, omise.ErrInvalidKey
	case pkey != "" && !strings.HasPrefix(pkey, "pkey_"):
		return nil, omise.ErrInvalidKey
	case skey != "" && !strings.HasPrefix(skey, "skey_"):
		return nil, omise.ErrInvalidKey
	}

	return &Client{
		pkey:        pkey,
		skey:        skey,
		OmiseClient: omise.Client{},
		Conns:       make(map[string]*persistConn, 2),
	}, nil
}

func (c *Client) Do(ctx context.Context, req *http.Request, result interface{}) error {
	defer req.Body.Close()

	var (
		host = req.URL.Hostname()
		port = req.URL.Port()
	)
	if _, ok := c.Conns[host]; !ok || c.Conns[host] == nil {
		conn, err := NewConn(ctx, host, port)
		if err != nil {
			return err
		}
		c.Conns[host] = conn
	}

	errCh := make(chan error, 1)

	go func() {
		err := req.Write(c.Conns[host].bw)
		c.Conns[host].bw.Flush()
		errCh <- err
	}()

	if err := <-errCh; err != nil {
		return err
	}

	resp, err := http.ReadResponse(c.Conns[host].br, req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	buffer, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		err = &omise.ErrTransport{Err: err, Buffer: buffer}
	}

	switch {
	case resp.StatusCode != 200:
		oErr := &omise.Error{StatusCode: resp.StatusCode}
		err = json.Unmarshal(buffer, oErr)
		return err
	}

	return json.Unmarshal(buffer, result)
}

func buildAddres(endpoint string) string {
	u, err := url.Parse(endpoint)
	if err != nil {
		log.Fatalf("[FATAL] failed to parse endpoint(%s)", endpoint)
	}
	return fmt.Sprintf("%s:%s", u.Hostname(), schemas[u.Scheme])
}
