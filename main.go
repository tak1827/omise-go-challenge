package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"sync/atomic"

	"github.com/tak1827/go-queue/queue"
)

const (
	// NOTE: Read these from environment variables or configuration files!
	OmisePublicKey  = "pkey_test_5q93kauv5ilajuv91ry"
	OmiseSecretKey  = "skey_test_5q93kauvf2839xezykz"
	OmisePublicKey2 = "pkey_test_5qb85uplgtj83tmb5y1"
	OmiseSecretKey2 = "skey_test_5qb85uplr54fv4ip3f3"
)

func main() {
	var (
		path       = filePath()
		offset     = int64(50) // omit header, optimized for this challenge
		bsize      = int64(256)
		buffer     = make([]byte, bsize)
		donatorCh  = make(chan Donator, 1)
		qsize      = 1024 // donators are less than this
		q          = queue.NewQueue(qsize, false)
		sum        = NewSummary()
		donatorNum = uint32(0)
		finishedCh = make(chan struct{}, 1)
	)

	// open csv
	file, err := os.OpenFile(path, os.O_RDONLY, 0644)
	if err != nil {
		panic(err)
	}

	// create cipher reader
	rotReader, err := DonatorReader(file, donatorCh)
	if err != nil {
		panic(err)
	}

	// prepare workers
	var (
		ctx, stopWorkers = context.WithCancel(context.Background())
		callback         = func(d Donator, succeeded bool) {
			amount := d.Amount()

			sum.IncrementNum(1)
			sum.IncrementReceived(amount)

			if succeeded {
				sum.IncrementDonated(amount)
				sum.UpdateTop(d.Name, amount)
			} else {
				sum.IncrementFaulty(amount)
			}

			if atomic.LoadUint32(&sum.Num) >= donatorNum {
				close(finishedCh)
			}
		}
		workers = []*Worker{
			NewWorker(&q, OmisePublicKey, OmiseSecretKey, callback),
			NewWorker(&q, OmisePublicKey2, OmiseSecretKey2, callback),
		}
	)

	// run workers
	for i := range workers {
		go workers[i].Run(ctx, false)
	}

	fmt.Println("performing donations...")

	// read encripted csv
	go func() {
		defer close(donatorCh)

		for {
			if _, err = rotReader.ReadAt(buffer, offset); err != nil {
				if err == io.EOF {
					break
				}
				panic(err)
			}

			offset += bsize
		}
	}()

	// enqueue donation tasks
	for donator := range donatorCh {
		donatorNum += 1
		if err = q.Enqueue(donator); err != nil {
			panic(fmt.Sprintf("queue size is too small, err: %s", err.Error()))
		}
	}

	// wait for all tasks completion
	<-finishedCh

	// stop workers
	stopWorkers()
	for i := range workers {
		workers[i].Close()
	}

	fmt.Println("done.")

	// print the result
	sum.Print()
}

func filePath() string {
	if len(os.Args) != 2 {
		panic("pass the donation csv path as an argument")
	}
	path := os.Args[1]
	if _, err := os.Stat(path); err != nil {
		panic(fmt.Sprintf("invalid argument: %s milisec", err.Error()))
	}
	return path
}
