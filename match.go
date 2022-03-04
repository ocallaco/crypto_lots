package main

import (
	"fmt"
	"sort"
	"time"
)

type LotBudget struct {
	Trade     *Trade
	Remaining DotEight
}

type LotMatches struct {
	index   int
	budgets []LotBudget
}

func (lm *LotMatches) Len() int {
	return len(lm.budgets)
}

func (lm *LotMatches) Less(i, j int) bool {
	return lm.budgets[i].Trade.Time.Before(lm.budgets[j].Trade.Time)
}

func (lm *LotMatches) Swap(i, j int) {
	lm.budgets[i], lm.budgets[j] = lm.budgets[j], lm.budgets[i]
}

func (lm *LotMatches) Insert(trade *Trade) {
	lb := LotBudget{
		Trade:     trade,
		Remaining: trade.TopAmt,
	}
	lm.budgets = append(lm.budgets, lb)
}

func (lm *LotMatches) Prepare() {
	sort.Sort(lm)
	lm.index = 0
}

type Lot struct {
	Start time.Time
	End   time.Time
	PandL DotEight
}

func (l Lot) String() string {
	return fmt.Sprintf("%+v-%+v : $%s", l.Start, l.End, l.PandL.ToString())
}

func (lm *LotMatches) MatchTrade(b *LotBudget) ([]Lot, error) {
	// TODO: don't allow matching against new trades
	if lm.index < 0 {
		return nil, fmt.Errorf("not prepared")
	}
	i := lm.index
	res := []Lot{}
	unitBasis := b.Trade.BottomAmt.Div(b.Trade.TopAmt)
	for b.Remaining > 0 {
		if len(lm.budgets) < i {
			return nil, fmt.Errorf("ran out of trades to match")
		}
		nextLot := lm.budgets[i]
		nextLotBasis := nextLot.Trade.BottomAmt.Div(nextLot.Trade.TopAmt)

		lotAmt := b.Remaining
		if nextLot.Remaining <= b.Remaining {
			lotAmt = nextLot.Remaining
			i++
		}
		nextLot.Remaining -= lotAmt
		b.Remaining -= lotAmt
		res = append(res, Lot{
			Start: nextLot.Trade.Time,
			End:   b.Trade.Time,
			PandL: unitBasis.Mul(lotAmt) - nextLotBasis.Mul(lotAmt),
		})
	}
	lm.index = i
	return res, nil
}

func NewLotMatches() *LotMatches {
	return &LotMatches{
		index:   -1,
		budgets: make([]LotBudget, 0),
	}
}
