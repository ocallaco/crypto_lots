package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path"
	"time"
)

type PairHist map[string]DotEight

type PriceService struct {
	data map[Instrument]PairHist
}

func BuildPriceService(pth string, base Instrument) (*PriceService, error) {
	ps := &PriceService{
		data: map[Instrument]PairHist{},
	}

	files, err := os.ReadDir(pth)
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		top, bot, err := ReadPair(file.Name())
		if err != nil {
			continue
		}
		invert := false
		if bot != base {
			if top != base {
				continue
			}
			top = bot
			invert = true
		}

		f, err := os.Open(path.Join(pth, file.Name()))
		if err != nil {
			return nil, err
		}
		r := csv.NewReader(f)
		prices := map[string]DotEight{}

		for {
			record, err := r.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				break
			}
			if len(record) != 2 {
				return nil, fmt.Errorf("bad entry: %+v", record)
			}

			_, err = time.Parse(dateFormat, record[0])
			if err != nil {
				return nil, fmt.Errorf("bad date entry: %s, %w", record[0], err)
			}

			px := ToDotEight(record[1])
			if invert {
				px = px.Recip()
			}
			prices[record[0]] = px
		}

		ps.data[top] = prices
	}

	return ps, nil
}

func (ps *PriceService) GetHistoricalPrice(top Instrument, t time.Time) (DotEight, error) {
	tS := t.Format(dateFormat)
	prices, ok := ps.data[top]
	if !ok {
		return DotEight(0), fmt.Errorf("no historical prices for %s", top)
	}

	price, ok := prices[tS]
	if !ok {
		return DotEight(0), fmt.Errorf("no historical price for %s on  date %s", top, tS)
	}
	return price, nil
}
