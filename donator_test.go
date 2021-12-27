package main

import (
	"io"
	"os"
	"testing"
	"time"

	r "github.com/stretchr/testify/require"
)

func TestDonatorReader(t *testing.T) {
	file, err := os.OpenFile("./test/fng.1000.csv.rot128", os.O_RDONLY, 0644)
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
		got = append(got, donator)
	}

	r.Equal(t, len(expected), len(got))

	for i := range got {
		val, ok := expected[got[i].Name]
		r.Equal(t, true, ok)
		r.Equal(t, val, got[i])
	}
}

func TestOnlyNumber(t *testing.T) {
	var err error

	err = onlyNumber("12345")
	r.NoError(t, err)

	err = onlyNumber("12345a")
	r.EqualError(t, err, ErrInvalidDonatorFormat.Error())
}

func TestValidateMonth(t *testing.T) {
	var err error

	err = validateMonth("1")
	r.NoError(t, err)

	err = validateMonth("01")
	r.NoError(t, err)

	err = validateMonth("11")
	r.NoError(t, err)

	err = validateMonth("13")
	r.EqualError(t, err, ErrInvalidDonatorFormat.Error())
}

func TestValidateExpiry(t *testing.T) {
	var (
		err    error
		yy, mm int64
		now    = time.Date(2021, time.Month(11), 1, 0, 0, 0, 0, &time.Location{})
	)

	mm, yy, err = validateExpiry("12", "2021", now)
	r.NoError(t, err)
	r.Equal(t, int64(12), mm)
	r.Equal(t, int64(2021), yy)

	mm, yy, err = validateExpiry("01", "2022", now)
	r.NoError(t, err)
	r.Equal(t, int64(1), mm)
	r.Equal(t, int64(2022), yy)

	mm, yy, err = validateExpiry("11", "2021", now)
	r.NoError(t, err)
	r.Equal(t, int64(11), mm)
	r.Equal(t, int64(2021), yy)

	mm, yy, err = validateExpiry("10", "2021", now)
	r.EqualError(t, err, ErrExpiredCard.Error())
}
