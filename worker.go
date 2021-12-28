package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/omise/omise-go"
	"github.com/omise/omise-go/operations"
	"github.com/tak1827/go-queue/queue"
	"github.com/tak1827/omise-go-challenge/client"
)

const (
	TestEndpoint = "localhost:80"

	FastInterval = int64(10)    // 100 milsec
	SlowInterval = int64(10000) // 1s
)

type CallbackFunc func(d Donator, succeeded bool)

type Worker struct {
	q   *queue.Queue
	cli *client.Client

	interval int64 // milsec

	callback CallbackFunc
}

var DefaultCallback = func(d Donator, succeeded bool) {}

func NewWorker(q *queue.Queue, pkey, skey string, callback CallbackFunc) *Worker {
	cli, err := client.NewClient(pkey, skey)
	if err != nil {
		panic(fmt.Sprintf("failed to create client, err: %v", err))
	}

	if callback == nil {
		callback = DefaultCallback
	}

	return &Worker{
		q:        q,
		cli:      &cli,
		interval: FastInterval,
		callback: callback,
	}
}

func (w *Worker) Run(ctx context.Context, isTest bool) {
	var (
		timer     = time.NewTicker(time.Duration(w.interval) * time.Millisecond)
		d         Donator
		tokenMsg  operations.CreateToken
		chargeMsg operations.CreateCharge
		req       *http.Request
		token     omise.Token
		charge    omise.Charge
		c         = context.Background()
		err       error
	)

	for {
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
			elm, empty := w.q.Dequeue()
			if empty {
				continue
			}

			d = elm.(Donator)
			if tokenMsg, err = d.GenCreateTokenMsg(); err != nil {
				if !errors.Is(err, ErrExpiredCard) {
					log.Printf("[WARN] invalid donator format(%v)\n", elm)
				}
				w.handleErr(d, err)
				continue
			}

			if req, err = w.cli.OmiseClient.Request(&tokenMsg); err != nil {
				w.handleErr(d, err)
				continue
			}

			if isTest { // overwrite url for test
				req.URL.Host = TestEndpoint
			}

			if err = w.cli.Do(c, req, &token); err != nil {
				if errors.Is(err, client.ErrRateLimit) {
					w.handleRatelimit(timer, d)
				} else if errors.Is(err, io.ErrUnexpectedEOF) {
					w.handleEOF(timer, d)
				} else {
					w.handleErr(d, err)
				}
				continue
			}

			if chargeMsg, err = d.GenCreateChargeMsg(token.Base.ID); err != nil {
				w.handleErr(d, err)
				continue
			}

			if req, err = w.cli.OmiseClient.Request(&chargeMsg); err != nil {
				w.handleErr(d, err)
				continue
			}

			if isTest { // overwrite url for test
				req.URL.Host = TestEndpoint
			}

			if err = w.cli.Do(c, req, &charge); err != nil {
				if errors.Is(err, client.ErrRateLimit) {
					w.handleRatelimit(timer, d)
				} else if errors.Is(err, io.ErrUnexpectedEOF) {
					w.handleEOF(timer, d)
				} else {
					w.handleErr(d, err)
				}
				continue
			}

			w.callback(d, true)
			w.resetTimer(timer, FastInterval)
		}
	}
}

func (w *Worker) resetTimer(timer *time.Ticker, interval int64) {
	w.interval = interval
	timer.Reset(time.Duration(w.interval) * time.Millisecond)
}

func (w *Worker) handleErr(d Donator, err error) {
	// log.Printf("[WARN] err: %s, interval: %d milisec\n", err.Error(), w.interval)
	w.callback(d, false)
}

func (w *Worker) handleRatelimit(timer *time.Ticker, d Donator) {
	log.Printf("[WARN] ratelimit, interval: %d milisec\n", w.interval)
	if err := w.q.Enqueue(d); err != nil {
		panic(fmt.Sprintf("failed to enqueue(%v), err: %s", d, err.Error()))
	}
	w.resetTimer(timer, SlowInterval)
}

func (w *Worker) handleEOF(timer *time.Ticker, d Donator) {
	log.Printf("[WARN] unexpected EOF, interval: %d milisec\n", w.interval)
	if err := w.q.Enqueue(d); err != nil {
		panic(fmt.Sprintf("failed to enqueue(%v), err: %s", d, err.Error()))
	}
	w.resetTimer(timer, FastInterval)
	w.cli.ResetConn()
}

func (w *Worker) Close() {
	w.cli.Close()
}
