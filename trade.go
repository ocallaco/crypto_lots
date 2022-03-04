package main

import (
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
}

func (t *Trade) Split(baseInst Instrument, basePrice DotEight) []*Trade {
	baseAmt := t.TopAmt.Mul(basePrice)

	leftTrade := &Trade{
		ID:         t.ID + "left",
		Time:       t.Time,
		IsSell:     t.IsSell,
		TopInst:    t.TopInst,
		BottomInst: baseInst,
		TopAmt:     t.TopAmt,
		BottomAmt:  baseAmt,
	}
	rightTrade := &Trade{
		ID:         t.ID + "right",
		Time:       t.Time,
		IsSell:     !t.IsSell,
		TopInst:    t.BottomInst,
		BottomInst: baseInst,
		TopAmt:     t.BottomAmt,
		BottomAmt:  baseAmt,
	}

	return []*Trade{
		leftTrade,
		rightTrade,
	}

}
