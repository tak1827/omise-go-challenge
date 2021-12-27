package main

import (
	"fmt"
	"sync"
	"testing"

	r "github.com/stretchr/testify/require"
)

func TestSummary(t *testing.T) {
	var (
		sum      = NewSummary()
		donators = []string{
			"account-7",
			"account-1",
			"account-3",
			"account-5",
			"account-2",
			"account-8",
			"account-4",
			"account-6",
			"account-9",
		}
		donatorAmounts = map[string]int64{
			"account-1": int64(100),
			"account-2": int64(200),
			"account-3": int64(300),
			"account-4": int64(400),
			"account-5": int64(500),
			"account-6": int64(600),
			"account-7": int64(700),
			"account-8": int64(800),
			"account-9": int64(900),
		}
		expectedTop = [3]string{
			"account-9",
			"account-8",
			"account-7",
		}
		expectedTotal = int64(0)
		wg            sync.WaitGroup
	)

	for _, amount := range donatorAmounts {
		expectedTotal += amount
	}

	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			counter := index
			for counter < len(donators) {
				name := donators[counter]
				amount := donatorAmounts[name]
				sum.IncrementNum(1)
				sum.IncrementReceived(amount)
				sum.UpdateTop(name, amount)
				counter += 3
			}
		}(i)
	}

	wg.Wait()

	fmt.Printf("sum: %v\n", sum)

	r.EqualValues(t, expectedTotal, sum.Received)
	r.EqualValues(t, expectedTop, sum.Top)
	r.EqualValues(t, donatorAmounts[expectedTop[0]], sum.TopAmount[sum.Top[0]])
	r.EqualValues(t, donatorAmounts[expectedTop[1]], sum.TopAmount[sum.Top[1]])
	r.EqualValues(t, donatorAmounts[expectedTop[2]], sum.TopAmount[sum.Top[2]])
}
