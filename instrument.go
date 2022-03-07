package main

import (
	"fmt"
	"strings"
)

var ErrNotAPair = fmt.Errorf("not a pair")

type Instrument string

func ReadPair(name string) (Instrument, Instrument, error) {
	var left, right string
	if strings.Contains(name, "-") {
		s := strings.Split(name, "-")
		left, right = s[0], s[1]
	} else if strings.Contains(name, "/") {
		s := strings.Split(name, "/")
		left, right = s[0], s[1]
	} else {
		return Instrument(""), Instrument(""), ErrNotAPair
	}

	return Instrument(strings.ToLower(left)), Instrument(strings.ToLower(right)), nil
}
