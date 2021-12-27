package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/omise/omise-go"
	"github.com/omise/omise-go/operations"
	r "github.com/stretchr/testify/require"
)

const (
	TestPublicKey = "pkey_test_12345"
	TestSecretKey = "skey_test_12345"
	TestEndpoint  = "localhost:80"
)

func listenTestServer() *http.Server {
	srv := &http.Server{Addr: TestEndpoint}
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, %s", r.URL.Path)
	})
	http.HandleFunc("/charges", func(w http.ResponseWriter, r *http.Request) {
		obj := omise.Charge{}
		json.NewEncoder(w).Encode(obj)
	})

	go func() {
		fmt.Print("test server is starting\n")
		if err := srv.ListenAndServe(); err != nil {
			panic(err)
		}

	}()

	return srv
}

func TestDo(t *testing.T) {
	// src := listenTestServer()
	// defer srv.Close()

	client, err := NewClient(TestPublicKey, TestSecretKey)
	r.NoError(t, err)
	defer client.Close()

	var (
		ctx    = context.Background()
		result = &omise.Charge{}
		msg    = &operations.CreateCharge{
			Amount:   100000, // ฿ 1,000.00
			Currency: "thb",
			Card:     "tokn_test_123",
		}
	)

	req, err := client.OmiseClient.Request(msg)
	r.NoError(t, err)

	// overwrite test endpoint
	req.URL.Host = TestEndpoint

	err = client.Do(ctx, req, result)
	r.NoError(t, err)
}

func BenchmarkClient(b *testing.B) {
	// src := listenTestServer()
	// defer srv.Close()

	client, _ := NewClient(TestPublicKey, TestSecretKey)
	defer client.Close()

	var size = 100

	b.ReportAllocs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		for i := 0; i < size; i++ {
			var (
				ctx    = context.Background()
				result = &omise.Charge{}
				msg    = &operations.CreateCharge{
					Amount:   100000, // ฿ 1,000.00
					Currency: "thb",
					Card:     "tokn_test_123",
				}
			)

			req, err := client.OmiseClient.Request(msg)
			if err != nil {
				b.Fatal(err)
			}

			// overwrite test endpoint
			req.URL.Host = TestEndpoint

			if err = client.Do(ctx, req, result); err != nil {
				b.Fatal(err)
			}
		}
	}
}

func BenchmarkOmiseClient(b *testing.B) {
	// src := listenTestServer()
	// defer srv.Close()

	client, err := omise.NewClient(TestPublicKey, TestSecretKey)
	if err != nil {
		b.Fatal(err)
	}

	var size = 100

	b.ReportAllocs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		for i := 0; i < size; i++ {
			var (
				result = &omise.Charge{}
				msg    = &operations.CreateCharge{
					Amount:   100000, // ฿ 1,000.00
					Currency: "thb",
					Card:     "tokn_test_123",
				}
			)

			req, err := client.Request(msg)
			if err != nil {
				b.Fatal(err)
			}

			// overwrite test endpoint
			req.URL.Host = TestEndpoint
			req.URL.Scheme = "http"

			if err = omiseClientDo(client, req, result); err != nil {
				b.Fatal(err)
			}
		}
	}
}

func omiseClientDo(c *omise.Client, req *http.Request, result interface{}) error {
	resp, err := c.Client.Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return err
	}

	buffer, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return &omise.ErrTransport{Err: err, Buffer: buffer}
	}

	switch {
	case resp.StatusCode != 200:
		err := &omise.Error{StatusCode: resp.StatusCode}
		if err := json.Unmarshal(buffer, err); err != nil {
			return &omise.ErrTransport{Err: err, Buffer: buffer}
		}

		return err
	} // status == 200 && e == nil

	if result != nil {
		if err := json.Unmarshal(buffer, result); err != nil {
			return &omise.ErrTransport{Err: err, Buffer: buffer}
		}
	}

	return nil
}
