package main

import (
	"crypto/sha1"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"sort"
)

type Lot struct {
	Buy            *Trade
	Sell           *Trade
	BuyPx          DotEight
	SellPx         DotEight
	Amt            DotEight
	PandL          DotEight
	PreviousReport bool
}

func (l *Lot) String() string {
	return fmt.Sprintf("%+v\t%+v\t%s\t%s\t%s\t$%s\t%s", l.Buy.Time.Format(dateFormat), l.Sell.Time.Format(dateFormat), l.Amt.ToString(), l.BuyPx.ToString(), l.SellPx.ToString(), l.PandL.ToString())
}

func (l *Lot) CSV() []string {
	return []string{l.Buy.Time.Format(dateFormat), l.Sell.Time.Format(dateFormat), l.Amt.Mul(l.SellPx).ToString(), l.Amt.Mul(l.BuyPx).ToString(), string(l.Buy.TopInst), l.SHA()}
}

func (l *Lot) SHA() string {
	w := sha1.New()
	w.Write([]byte(l.String()))
	return fmt.Sprintf("%x", w.Sum(nil))
}

func (l *Lot) ReportedLot() []string {
	return []string{l.Buy.ID, l.Sell.ID, l.Amt.ToString()}
}

type LotBudget struct {
	Trade     *Trade
	Remaining DotEight
}

type LotReport struct {
	Buy  string
	Sell string
	Amt  DotEight
}

func ReadLotReports(pth string) ([]LotReport, error) {
	reports := make([]LotReport, 0)
	f, err := os.Open(pth)
	if err != nil {
		return nil, err
	}
	r := csv.NewReader(f)
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if len(record) < 3 {
			return nil, fmt.Errorf("bad lot report record: %+v", record)
		}
		reports = append(reports, LotReport{
			Buy:  record[0],
			Sell: record[1],
			Amt:  ToDotEight(record[2]),
		})
	}
	return reports, nil
}

func MatchTrades(trades []*Trade, LIFO bool, reportedLots []LotReport) ([]*Lot, error) {
	buys := []*LotBudget{}
	sells := []*LotBudget{}

	idToBudget := map[string]*LotBudget{}

	for _, t := range trades {
		b := &LotBudget{
			Trade:     t,
			Remaining: t.TopAmt,
		}
		if t.IsSell {
			sells = append(sells, b)
		} else {
			buys = append(buys, b)
		}
		idToBudget[t.ID] = b
	}

	sort.Slice(sells, func(i, j int) bool {
		return sells[i].Trade.Time.Before(sells[j].Trade.Time)
	})

	if LIFO {
		sort.Slice(buys, func(i, j int) bool {
			return buys[j].Trade.Time.Before(buys[i].Trade.Time)
		})
	} else {
		sort.Slice(buys, func(i, j int) bool {
			return buys[i].Trade.Time.Before(buys[j].Trade.Time)
		})
	}

	lots := []*Lot{}

	for _, rep := range reportedLots {
		buy, ok := idToBudget[rep.Buy]
		if !ok {
			if _, ok := idToBudget[rep.Sell]; ok {
				return nil, fmt.Errorf("invalid reported lot Buy ID %s not present, but Sell is", rep.Buy)
			}
			// doesn't apply to this instrument, skip
			continue
		}
		sell, ok := idToBudget[rep.Sell]
		if !ok {
			return nil, fmt.Errorf("invalid reported lot Sell ID %s not present but Buy is", rep.Sell)
		}
		newL := MatchBuySell(buy, sell, rep.Amt)
		newL.PreviousReport = true
		lots = append(lots, newL)
	}

	for _, sell := range sells {
		if sell.Remaining == 0 {
			continue
		}
		for _, buy := range buys {
			if buy.Remaining == 0 || sell.Trade.Time.Before(buy.Trade.Time) {
				continue
			}
			lots = append(lots, MatchBuySell(buy, sell, DotEight(0)))
			if sell.Remaining == 0 {
				break
			}
		}
	}
	sort.Slice(lots, func(i, j int) bool {
		return lots[i].Sell.Time.Before(lots[j].Sell.Time)
	})
	return lots, nil
}

func MatchBuySell(buy, sell *LotBudget, amt DotEight) *Lot {
	soldPx := sell.Trade.Price()
	costBasisPx := buy.Trade.Price()
	lotAmt := sell.Remaining
	if amt > 0 {
		lotAmt = amt
	}
	if buy.Remaining < lotAmt {
		lotAmt = buy.Remaining
	}
	sell.Remaining -= lotAmt
	buy.Remaining -= lotAmt
	return &Lot{
		Buy:    buy.Trade,
		Sell:   sell.Trade,
		Amt:    lotAmt,
		PandL:  soldPx.Mul(lotAmt) - costBasisPx.Mul(lotAmt),
		BuyPx:  costBasisPx,
		SellPx: soldPx,
	}
}
