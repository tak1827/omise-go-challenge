package main

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/omise/omise-go/operations"
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

func (d Donator) GenCreateTokenMsg() (token operations.CreateToken, err error) {
	if d.Name == "" || d.CCNumber == "" || d.CVV == "" || d.ExpMonth == "" || d.ExpYear == "" {
		err = ErrInvalidDonatorFormat
		return
	}

	token.Name = d.Name

	// validate `CCNumber`
	if err = onlyNumber(d.CCNumber); err != nil {
		return
	}
	if len(d.CCNumber) != 16 {
		err = ErrInvalidDonatorFormat
		return
	}
	token.Number = d.CCNumber

	// validate `CVV`
	if err = onlyNumber(d.CVV); err != nil {
		return
	}
	if len(d.CVV) != 3 {
		err = ErrInvalidDonatorFormat
		return
	}
	token.SecurityCode = d.CVV

	// validate `ExpMonth` and `ExpYear`
	if err = onlyNumber(d.ExpMonth); err != nil {
		return
	}
	if err = validateMonth(d.ExpMonth); err != nil {
		return
	}
	if err = onlyNumber(d.ExpYear); err != nil {
		return
	}
	mm, yy, err := validateExpiry(d.ExpMonth, d.ExpYear, time.Now())
	if err != nil {
		return
	}
	token.ExpirationMonth = time.Month(mm)
	token.ExpirationYear = int(yy)

	token.City = "Bangkok"
	token.PostalCode = "10240"

	return
}

func (d Donator) GenCreateChargeMsg(token string) (charge operations.CreateCharge, err error) {
	if d.AmountSubunits == "" {
		err = ErrInvalidDonatorFormat
		return
	}

	// validate `AmountSubunits`
	if err = onlyNumber(d.AmountSubunits); err != nil {
		return
	}

	if charge.Amount, err = strconv.ParseInt(d.AmountSubunits, 10, 64); err != nil {
		err = ErrInvalidDonatorFormat
		return
	}

	charge.Currency = "thb"
	charge.Card = token

	return
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

func onlyNumber(str string) error {
	for i := range str {
		if !('0' <= str[i] && str[i] <= '9') {
			return ErrInvalidDonatorFormat
		}
	}
	return nil
}

func validateMonth(m string) (err error) {
	switch m {
	case "01":
		return
	case "02":
		return
	case "03":
		return
	case "04":
		return
	case "05":
		return
	case "06":
		return
	case "07":
		return
	case "08":
		return
	case "09":
		return
	case "10":
		return
	case "11":
		return
	case "12":
		return
	case "1":
		return
	case "2":
		return
	case "3":
		return
	case "4":
		return
	case "5":
		return
	case "6":
		return
	case "7":
		return
	case "8":
		return
	case "9":
		return
	default:
		err = ErrInvalidDonatorFormat
		return
	}
}

func validateExpiry(m, y string, now time.Time) (mm, yy int64, err error) {
	if mm, err = strconv.ParseInt(m, 10, 8); err != nil {
		err = ErrInvalidDonatorFormat
		return
	}
	if yy, err = strconv.ParseInt(y, 10, 16); err != nil {
		err = ErrInvalidDonatorFormat
		return
	}

	if yy*100+mm < int64(now.Year()*100)+int64(now.Month()) {
		err = ErrExpiredCard
		return
	}
	return
}
