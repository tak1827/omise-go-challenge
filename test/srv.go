package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/omise/omise-go"
)

func logging(url string) {
	log.Printf("[INFO] accessed: %s", url)
}

func server() *http.Server {
	srv := &http.Server{Addr: "localhost:80"}
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		logging(r.URL.Path)
		fmt.Fprintf(w, "Hello, %s", r.URL.Path)
	})
	http.HandleFunc("/tokens", func(w http.ResponseWriter, r *http.Request) {
		logging(r.URL.Path)
		if time.Now().Unix()%5 == 0 {
			http.Error(w, "Rate Limit", 429)
		} else {
			obj := omise.Token{}
			json.NewEncoder(w).Encode(obj)
		}
	})
	http.HandleFunc("/charges", func(w http.ResponseWriter, r *http.Request) {
		logging(r.URL.Path)
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
