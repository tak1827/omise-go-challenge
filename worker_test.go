package main

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"

	r "github.com/stretchr/testify/require"
	"github.com/tak1827/go-queue/queue"
)

const (
	TestPublicKey = "pkey_test_12345"
	TestSecretKey = "skey_test_12345"
)

func TestWorker(t *testing.T) {
	var (
		q             = queue.NewQueue(16, false)
		interval      = int64(100)
		expectedCount = uint32(5)
		counter       = uint32(0)
	)

	ctx, cancel := context.WithCancel(context.Background())

	callback := func(d Donator, succeeded bool) {
		if succeeded {
			atomic.AddUint32(&counter, 1)
		}
		fmt.Printf("donator: %v\n", d)
	}

	worker := NewWorker(&q, interval, TestPublicKey, TestSecretKey, callback)

	go func() {
		worker.Run(ctx, true)
	}()

	donator := Donator{
		Name:           "account",
		AmountSubunits: "0",
		CCNumber:       "1234567890123456",
		CVV:            "123",
		ExpMonth:       "1",
		ExpYear:        "2000",
	}
	// enqueue wrong data
	err := q.Enqueue(donator)
	r.NoError(t, err)
	// enqueue collect data
	for i := uint32(0); i < expectedCount; i++ {
		donator.Name = fmt.Sprintf("acount-%d", i)
		donator.AmountSubunits = fmt.Sprintf("%d00", i)
		donator.ExpYear = "3000"
		err = q.Enqueue(donator)
		r.NoError(t, err)
	}

	for {
		if atomic.LoadUint32(&counter) >= expectedCount {
			cancel()
			break
		}
	}

	r.Equal(t, expectedCount, counter)
	r.Equal(t, true, q.IsEmpty())
}
