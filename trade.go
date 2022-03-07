package main

import (
	"strings"
	"time"
)

type Trade struct {
	ID         string
	Time       time.Time
	IsSell     bool
	TopInst    Instrument
	BottomInst Instrument
	TopAmt     DotEight
	BottomAmt  DotEight
	FeeAmt     DotEight
	Exchange   string
}

func (t *Trade) Price() DotEight {
	fee := t.FeeAmt
	if t.IsSell {
		fee = -fee
	}
	return (t.BottomAmt + fee).Div(t.TopAmt)
}

// if we swapped 2 non-USD instruments directly, we need to translate into a trade of Top for USD and USD for Bot
func (t *Trade) Split(baseInst Instrument, basePrice DotEight, useBottom bool) []*Trade {
	var fee, baseAmt DotEight

	if useBottom {
		baseAmt = t.BottomAmt.Mul(basePrice)
		fee = t.FeeAmt.Mul(basePrice)
	} else {
		baseAmt = t.TopAmt.Mul(basePrice)
		// if using price of top inst, we need to translate the fee into the base Inst via the top inst
		fee = t.FeeAmt.Mul(t.TopAmt.Div(t.BottomAmt)).Mul(basePrice) // TODO: test for correctness
	}

	if t.IsSell {
		fee = -fee
	}

	return []*Trade{
		&Trade{
			ID:         t.ID + "_left",
			Time:       t.Time,
			IsSell:     t.IsSell,
			TopInst:    t.TopInst,
			BottomInst: baseInst,
			TopAmt:     t.TopAmt,
			BottomAmt:  baseAmt + fee,
		},
		&Trade{
			ID:         t.ID + "_right",
			Time:       t.Time,
			IsSell:     !t.IsSell,
			TopInst:    t.BottomInst,
			BottomInst: baseInst,
			TopAmt:     t.BottomAmt,
			BottomAmt:  baseAmt,
		},
	}
}

func (t *Trade) String() string {
	bs := "Buy"
	if t.IsSell {
		bs = "Sell"
	}
	resSlice := []string{
		t.ID,
		t.Time.Format(dateFormat),
		bs,
		string(t.TopInst),
		string(t.BottomInst),
		t.TopAmt.ToString(),
		t.BottomAmt.ToString(),
		t.FeeAmt.ToString(),
		t.Exchange,
	}
	return strings.Join(resSlice, "\t")
}
