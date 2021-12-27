package main

import (
	"context"
	"errors"
	"fmt"
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
)

type CallbackFunc func(d Donator, succeeded bool)

type Worker struct {
	q        *queue.Queue
	cli      client.Client
	interval int64 // sec

	callback CallbackFunc
}

var DefaultCallback = func(d Donator, succeeded bool) {}

func NewWorker(q *queue.Queue, interval int64, pkey, skey string, callback CallbackFunc) Worker {
	cli, err := client.NewClient(pkey, skey)
	if err != nil {
		panic(fmt.Sprintf("failed to create client, err: %v", err))
	}

	if interval == 0 {
		interval = 1
	}

	if callback == nil {
		callback = DefaultCallback
	}

	return Worker{
		q:        q,
		cli:      cli,
		interval: interval,
		callback: callback,
	}
}

func (w *Worker) Run(ctx context.Context, isTest bool) {
	var (
		timer     = time.NewTicker(time.Duration(w.interval) * time.Second)
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
				if errors.Is(err, ErrExpiredCard) {
					log.Printf("[WARN] expired card(%v)\n", elm)
				}
				w.handleErr(d)
				continue
			}

			if req, err = w.cli.OmiseClient.Request(&tokenMsg); err != nil {
				w.handleErr(d)
				continue
			}

			if isTest { // overwrite url for test
				req.URL.Host = TestEndpoint
			}

			if err = w.cli.Do(c, req, &token); err != nil {
				if errors.Is(err, client.ErrRateLimit) {
					w.handleRatelimit(timer, d)
				} else {
					w.handleErr(d)
				}
				continue
			}

			if chargeMsg, err = d.GenCreateChargeMsg(token.Base.ID); err != nil {
				w.handleErr(d)
				continue
			}

			if req, err = w.cli.OmiseClient.Request(&chargeMsg); err != nil {
				w.handleErr(d)
				continue
			}

			if isTest { // overwrite url for test
				req.URL.Host = TestEndpoint
			}

			if err = w.cli.Do(c, req, &charge); err != nil {
				if errors.Is(err, client.ErrRateLimit) {
					w.handleRatelimit(timer, d)
				} else {
					w.handleErr(d)
				}
				continue
			}

			w.callback(d, true)
			w.resetTimer(timer, true)
		}
	}
}

func (w *Worker) resetTimer(timer *time.Ticker, succeeded bool) {
	if succeeded {
		w.interval = w.interval / 2
	} else {
		w.interval = w.interval * 4
	}

	if w.interval == 0 {
		w.interval = 1
	}

	timer.Reset(time.Duration(w.interval) * time.Second)
}

func (w *Worker) handleErr(d Donator) {
	w.callback(d, false)
}

func (w *Worker) handleRatelimit(timer *time.Ticker, d Donator) {
	w.handleErr(d)
	w.resetTimer(timer, false)
	if err := w.q.Enqueue(d); err != nil {
		panic(fmt.Sprintf("failed to enqueue(%v), err: %s", d, err.Error()))
	}
}
