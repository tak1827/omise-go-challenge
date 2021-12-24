package main

import (
	"fmt"
	"io"
	"os"
	"testing"

	r "github.com/stretchr/testify/require"
)

func TestDonatorReader(t *testing.T) {
	file, err := os.OpenFile("./data-test/fng.1000.csv.rot128", os.O_RDONLY, 0644)
	r.NoError(t, err)

	var (
		bsize       = int64(80)
		buffer      = make([]byte, bsize)
		donatorCh   = make(chan Donator, 10)
		lineCounter = 0
		// omit header, optimized for this challenge
		offset = int64(50)
	)

	var (
		expected = map[string]Donator{
			"Mr. Grossman R Oldbuck": Donator{
				Name:           "Mr. Grossman R Oldbuck",
				AmountSubunits: "2879410",
				CCNumber:       "5375543637862918",
				CVV:            "488",
				ExpMonth:       "11",
				ExpYear:        "2021",
			},
			"Mr. Ferdinand H Took-Brandybuck": Donator{
				Name:           "Mr. Ferdinand H Took-Brandybuck",
				AmountSubunits: "2253551",
				CCNumber:       "5238569266360327",
				CVV:            "052",
				ExpMonth:       "11",
				ExpYear:        "2019",
			},
		}
	)

	rotReader, err := DonatorReader(file, donatorCh)
	r.NoError(t, err)

	go func() {
		defer close(donatorCh)

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

	var got []Donator
	for donator := range donatorCh {
		fmt.Printf("got: %v\n", donator)
		got = append(got, donator)
	}

	r.Equal(t, len(expected), len(got))

	for i := range got {
		val, ok := expected[got[i].Name]
		r.Equal(t, true, ok)
		r.Equal(t, val, got[i])
	}
}
