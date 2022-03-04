package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSplit(t *testing.T) {
	topInst := Instrument("eth")
	bottomInst := Instrument("btc")
	baseInst := Instrument("usd")

	trade := &Trade{
		IsSell:     false,
		TopInst:    topInst,
		BottomInst: bottomInst,
		TopAmt:     ToDotEight("30"),
		BottomAmt:  ToDotEight("1.5"),
	}

	trades := trade.Split(baseInst, ToDotEight("3000"))

	// TODO: test that trade and split trades result in same effect on account balances -- ie, ETH, USD, and BTC all end up the same
	assert.Equal(t, "90000.00000000", trades[0].BottomAmt.ToString())
	assert.Equal(t, "90000.00000000", trades[1].BottomAmt.ToString())
}
