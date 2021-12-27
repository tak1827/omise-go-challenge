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
	// Read these from environment variables or configuration files!
	OmisePublicKey = "pkey_test_5q93kauv5ilajuv91ry"
	OmiseSecretKey = "skey_test_5q93kauvf2839xezykz"
)

func filePath() string {
	if len(os.Args) != 2 {
		panic("pass the donation csv path as an argument")
	}
	path := os.Args[1]
	if _, err := os.Stat(path); err != nil {
		panic(fmt.Sprintf("invalid argument: %s", err.Error()))
	}
	return path
}

func main() {
	var (
		path       = filePath()
		offset     = int64(50) // omit header, optimized for this challenge
		bsize      = int64(1024)
		buffer     = make([]byte, bsize)
		donatorCh  = make(chan Donator, 1)
		qsize      = 128
		q          = queue.NewQueue(qsize, false)
		interval   = int64(1)
		sum        = NewSummary()
		donatorNum = uint32(0)
		finishedCh = make(chan struct{}, 1)
	)

	file, err := os.OpenFile(path, os.O_RDONLY, 0644)
	if err != nil {
		panic(err)
	}

	rotReader, err := DonatorReader(file, donatorCh)
	if err != nil {
		panic(err)
	}

	callback := func(d Donator, succeeded bool) {
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

		fmt.Printf("donator: %v\n", d)
	}

	workers := []*Worker{
		NewWorker(&q, interval, OmisePublicKey, OmiseSecretKey, callback),
	}

	ctx, cancel := context.WithCancel(context.Background())

	for i := range workers {
		go workers[i].Run(ctx, false)
	}

	go func() {
		defer close(donatorCh)
		lineCounter := 0
		for {
			if _, err = rotReader.ReadAt(buffer, offset); err != nil {
				if err == io.EOF {
					break
				}
				panic(err)
			}

			offset += bsize

			lineCounter += 1
			if lineCounter >= 2 {
				break
			}
		}
	}()

	for donator := range donatorCh {
		donatorNum += 1
		q.Enqueue(donator)
	}

	<-finishedCh
	cancel()
	fmt.Printf("sum: %v\n", sum)
}
