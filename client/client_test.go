package client

import (
	"context"
	"encoding/json"
	"fmt"
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

func TestSend(t *testing.T) {
	listenTestServer()
	// defer srv.Close()

	var (
		ctx = context.Background()
	)

	client, err := NewClient(TestPublicKey, TestSecretKey)
	r.NoError(t, err)

	// Creates a charge from the token
	charge := &omise.Charge{}
	createCharge := &operations.CreateCharge{
		Amount:   100000, // ฿ 1,000.00
		Currency: "thb",
		Card:     "tokn_test_5q97nvva6qxx0tvl5q9",
	}

	req, err := client.OmiseClient.Request(createCharge)
	r.NoError(t, err)

	// overwrite test endpoint
	req.URL.Host = TestEndpoint

	err = client.Do(ctx, req, charge)
	r.NoError(t, err)
}
