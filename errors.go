package main

import (
	"errors"
)

var ErrInvalidArgs = errors.New("invalid args")
var ErrInvalidDonatorFormat = errors.New("invalid donator format")
var ErrExpiredCard = errors.New("expired card")
