package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/omise/omise-go"
)

const (
	// Read these from environment variables or configuration files!
	OmisePublicKey = "pkey_test_5q93kauv5ilajuv91ry"
	OmiseSecretKey = "skey_test_5q93kauvf2839xezykz"
)

func server() *http.Server {
	srv := &http.Server{Addr: "localhost:80"}
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, %s", r.URL.Path)
	})
	http.HandleFunc("/charges", func(w http.ResponseWriter, r *http.Request) {
		obj := omise.Charge{}
		json.NewEncoder(w).Encode(obj)
	})

	return srv
}

func main() {
	srv := server()
	if err := srv.ListenAndServe(); err != nil {
		panic(err)
	}
}
