package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	flags "github.com/jessevdk/go-flags"
)

const dateFormat = "1/2/2006"

type Trade struct {
	Time       time.Time
	IsSell     bool
	TopInst    Instrument
	BottomInst Instrument
	TopAmt     DotEight
	BottomAmt  DotEight
}

type HistoricalPrice struct {
	Time           time.Time
	TopAmtDotEight int64
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

type DotEight int64
type Instrument string

func (de DotEight) ToString() string {
	left := de / 1e8
	right := de % 1e8
	if right < 0 {
		right = -right
	}
	return fmt.Sprintf("%d.%d", left, right)
}

func (de DotEight) ToFloat64() float64 {
	return float64(de) / 1e8
}

func ToDotEight(s string) DotEight {
	comps := strings.Split(s, ".")
	if len(comps) > 2 {
		panic(fmt.Sprintf("invalid string for DotEight: %s\n", s))
	}

	left, err := strconv.Atoi(comps[0])
	if err != nil {
		panic(fmt.Sprintf("invalid string for DotEight: %s\n", s))
	}
	right := 0
	if len(comps) == 2 {
		rightStr := (comps[1] + "00000000")[0:8]
		right, err = strconv.Atoi(rightStr)
		if err != nil {
			panic(fmt.Sprintf("invalid string for DotEight: %s\n", s))
		}
	}
	dotEight := int64(left) * int64(1e8)
	dotEight = dotEight + int64(right)
	return DotEight(dotEight)
}

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
	TradesCSV string `long:"trades" required:"true"`
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

		trades = append(trades, entry.ToTrade())
	}

	accounts := Account{}
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
			TopBalance -= t.TopAmt
			BotBalance += t.BottomAmt
		} else {
			TopBalance += t.TopAmt
			BotBalance -= t.BottomAmt
		}
		accounts[t.TopInst] = TopBalance
		accounts[t.BottomInst] = BotBalance
	}

	fmt.Println(accounts.String())
}
