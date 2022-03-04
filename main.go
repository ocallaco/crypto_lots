package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	flags "github.com/jessevdk/go-flags"
)

const dateFormat = "1/2/2006"

type HistoricalPrice struct {
	Time time.Time
}

type Entry struct {
	ID     string
	Action string
	Date   string
	Pair   string
	Amt1   string
	Amt2   string
	Fee    string
}

type Instrument string

func (e *Entry) ToTrade() *Trade {
	trade := &Trade{}
	if e.Action == "sell" {
		trade.IsSell = true
	} else if e.Action != "buy" {
		panic(fmt.Sprintf("invalid B/S value: %s", e.Action))
	}
	t, err := time.Parse(dateFormat, e.Date)
	if err != nil {
		panic(err)
	}
	trade.Time = t
	trade.TopInst = Instrument(strings.ToLower(e.Pair[0:3]))
	trade.BottomInst = Instrument(strings.ToLower(e.Pair[4:7]))
	trade.TopAmt = ToDotEight(e.Amt1)
	trade.BottomAmt = ToDotEight(e.Amt2)
	return trade
}

type Account map[Instrument]DotEight

func (a Account) String() string {
	str := "{\n"
	for k, v := range a {
		str = str + "\t" + string(k) + ":" + v.ToString() + "\n"
	}
	return str + "}\n"
}

type Args struct {
	TradesCSV        string     `long:"trades" required:"true"`
	BaseInst         Instrument `long:"base" default:"usd"`
	HistoricalPrices string     `long:"prices"`
}

func GetHistoricalPrice(top, bottom Instrument, t time.Time) DotEight {
	// TODO:
	return ToDotEight("3000")
}

func main() {
	a := Args{}
	parser := flags.NewParser(&a, flags.Default)
	_, err := parser.Parse()
	if err != nil {
		if err, ok := err.(*flags.Error); ok && err.Type == flags.ErrHelp {
			os.Exit(0)
		} else {
			log.Fatalln(err)
		}
	}

	f, err := os.Open(a.TradesCSV)
	if err != nil {
		panic(err)
	}

	r := csv.NewReader(f)
	trades := []*Trade{}
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(err)
		}
		entry := Entry{
			ID:     record[0],
			Action: strings.ToLower(record[1]),
			Date:   record[2],
			Pair:   record[3],
			Amt1:   record[4],
			Amt2:   record[5],
			Fee:    record[6],
		}

		t := entry.ToTrade()
		if t.BottomInst != a.BaseInst {
			trades = append(trades, t.Split(a.BaseInst, GetHistoricalPrice(t.TopInst, a.BaseInst, t.Time))...)
		} else {
			trades = append(trades, t)
		}
	}

	accounts := Account{}
	buys := map[Instrument]*LotMatches{}
	sells := map[Instrument]*LotMatches{}

	for _, t := range trades {
		TopBalance, ok := accounts[t.TopInst]
		if !ok {
			TopBalance = DotEight(0)
		}
		BotBalance, ok := accounts[t.BottomInst]
		if !ok {
			BotBalance = DotEight(0)
		}

		if t.IsSell {
			var matches *LotMatches
			if matches, ok = sells[t.TopInst]; !ok {
				matches = NewLotMatches()
				sells[t.TopInst] = matches
			}
			matches.Insert(t)
			TopBalance -= t.TopAmt
			BotBalance += t.BottomAmt
		} else {
			var matches *LotMatches
			if matches, ok = buys[t.TopInst]; !ok {
				matches = NewLotMatches()
				buys[t.TopInst] = matches
			}
			matches.Insert(t)
			TopBalance += t.TopAmt
			BotBalance -= t.BottomAmt
		}
		accounts[t.TopInst] = TopBalance
		accounts[t.BottomInst] = BotBalance
	}

	for inst, b := range buys {
		if string(inst) != "btc" {
			continue
		}
		fmt.Println("currency:", inst)
		s := sells[inst]
		s.Prepare()
		b.Prepare()

		pandl := DotEight(0)

		for _, t := range s.budgets {
			l, err := b.MatchTrade(&t)
			if err != nil {
				panic(err)
			}
			for _, lot := range l {
				fmt.Println("match:", lot.String())
				pandl += lot.PandL
			}
		}
		fmt.Println("PANDL: $", pandl.ToString())
	}

	fmt.Println(accounts.String())
}
