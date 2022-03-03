package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	flags "github.com/jessevdk/go-flags"
)

const dateFormat = "1/2/2006"

type Trade struct {
	Time       time.Time
	TopInst    Instrument
	BottomInst Instrument
	TopAmt     int64
	BottomAmt  int64
}

type HistoricalPrice struct {
	Time           time.Time
	TopAmtDotEight int64
}

type Entry struct {
	ID     int
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
	if len(comps == 2) {
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

func (e *Entry) ToTrade() (*Trade, err) {
	trade := &Trade{}
	t, err := time.Parse(dateFormat, e.Date)
	trade.Time = t
	trade.TopInst = Instrument(strings.ToLower(c.Pair[0:3]))
	trade.BottomInst = Instrument(strings.ToLower(c.Pair[3:3]))
	trade.TopAmt = ToDotEight(e.Amt1)
	trade.BotAmt = ToDotEight(e.Amt2)
	return trade
}

func main() {

}
