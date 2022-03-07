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
	var baseAmt DotEight
	fee := t.FeeAmt // fee is additional cost in Bottom Inst

	if useBottom {
		baseAmt = t.BottomAmt.Mul(basePrice)
	} else {
		baseAmt = t.TopAmt.Mul(basePrice)
	}

	if t.IsSell {
		fee = -fee // if we're selling, then we're receiving less of the bottom inst, if we're buying, we're spending more
	}

	return []*Trade{
		&Trade{
			ID:         t.ID + "_left",
			Time:       t.Time,
			IsSell:     t.IsSell,
			TopInst:    t.TopInst,
			BottomInst: baseInst,
			TopAmt:     t.TopAmt,
			BottomAmt:  baseAmt,
		},
		&Trade{
			ID:         t.ID + "_right",
			Time:       t.Time,
			IsSell:     !t.IsSell,
			TopInst:    t.BottomInst,
			BottomInst: baseInst,
			TopAmt:     t.BottomAmt + fee, // fee only attached to right-side trade (not as FeeAmt since that's in the bottom inst)
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
