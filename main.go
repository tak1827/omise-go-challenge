package main

import (
	"fmt"
	"io"
	"os"

	"github.com/tak1827/go-queue/queue"
)

const (
	// Read these from environment variables or configuration files!
	OmisePublicKey = "pkey_test_5q93kauv5ilajuv91ry"
	OmiseSecretKey = "skey_test_5q93kauvf2839xezykz"
)

func filePath() string {
	if len(os.Args) != 1 {
		panic("pass the donation csv path as an argument")
	}
	path := os.Args[0]
	if _, err := os.Stat(path); err != nil {
		panic(fmt.Sprintf("invalid argument: %s", err.Error()))
	}
	return path
}

func main() {
	var (
		path      = filePath()
		bsize     = int64(80)
		buffer    = make([]byte, bsize)
		donatorCh = make(chan Donator, 1)
		qsize     = 128
		offset    = int64(50) // omit header, optimized for this challenge
	)

	file, err := os.OpenFile(path, os.O_RDONLY, 0644)
	if err != nil {
		panic(err)
	}

	rotReader, err := DonatorReader(file, donatorCh)
	if err != nil {
		panic(err)
	}

	q := queue.NewQueue(qsize, false)

	go func() {
		defer close(donatorCh)
		lineCounter := 0
		for {
			_, err = rotReader.ReadAt(buffer, offset)
			if err == io.EOF {
				break
			}

			offset += bsize

			lineCounter += 1
			if lineCounter >= 2 {
				break
			}
		}
	}()

	for donator := range donatorCh {
		q.Enqueue(donator)
	}
}
