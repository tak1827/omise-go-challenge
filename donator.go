package main

import (
	"log"
	"os"

	"github.com/tak1827/omise-go-challenge/cipher"
)

// NOTE: all string, cast inside cuncurent writter
type Donator struct {
	Name           string
	AmountSubunits string
	CCNumber       string
	CVV            string
	ExpMonth       string
	ExpYear        string
}

func DonatorReader(f *os.File, ch chan Donator) (*cipher.Rot128Reader, error) {
	var (
		donator = Donator{}
		row     = 0
		// size is optimized for this challenge
		line = make([]byte, 128)
		len  int
		skip = false
	)

	handler := func(b byte) {
		if skip && b != 0x0A {
			return
		}

		line[len] = b
		len += 1

		switch b {
		case 0x2C: // ","
			switch row {
			case 0:
				donator.Name = string(line[:len-1])
			case 1:
				donator.AmountSubunits = string(line[:len-1])
			case 2:
				donator.CCNumber = string(line[:len-1])
			case 3:
				donator.CVV = string(line[:len-1])
			case 4:
				donator.ExpMonth = string(line[:len-1])
			default:
				skip = true
				log.Printf("[WARN] invalid line(%s) in the csv\n", string(line[:len]))
			}
			row += 1
			len = 0
		case 0x0A: // "\n"
			donator.ExpYear = string(line[:len-1])
			ch <- donator
			row = 0
			len = 0
			skip = false
		default:
		}
	}

	return cipher.NewRot128Reader(f, handler)
}
