package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"strings"
	"time"

	flags "github.com/jessevdk/go-flags"
)

const dateFormat = "1/2/2006"

type HistoricalPrice struct {
	Time time.Time
}

type Entry struct {
	ID       string
	Action   string
	Date     string
	Pair     string
	Amt1     string
	Amt2     string
	Fee      string
	Exchange string
}

func (e *Entry) ToTrade() *Trade {
	trade := &Trade{}
	if e.Action == "sell" {
		trade.IsSell = true
	} else if e.Action != "buy" {
		return nil
	}
	t, err := time.Parse(dateFormat, e.Date)
	if err != nil {
		panic(err)
	}
	trade.Time = t
	left, right, err := ReadPair(e.Pair)
	if err != nil {
		panic(err)
	}

	fee := e.Fee
	if fee == "" {
		fee = "0"
	}
	trade.ID = e.ID
	trade.TopInst = left
	trade.BottomInst = right
	trade.TopAmt = ToDotEight(e.Amt1)
	trade.BottomAmt = ToDotEight(e.Amt2)
	trade.FeeAmt = ToDotEight(fee)
	trade.Exchange = e.Exchange

	if trade.TopAmt < 0 || trade.BottomAmt < 0 || trade.FeeAmt < 0 || trade.BottomAmt+trade.FeeAmt == 0 {
		panic("illegal trade entry.  fees/amounts can't be negative (and fee + bottom can't be 0) -- Buy/Sell is sufficient")
	}
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
	Input             string     `long:"input" required:"true"`
	BaseInst          Instrument `long:"base" default:"usd"`
	FIFO              bool       `long:"fifo"`
	Verbose           bool       `long:"verbose" short:"v"`
	OutDir            string     `long:"output" default:"/tmp"`
	StopDate          string     `long:"stop-date"`
	WriteReportedLots bool       `long:"write-reported-lots"`
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

	//make sure base inst is lowercase
	a.BaseInst = Instrument(strings.ToLower(string(a.BaseInst)))

	tradesCsv := path.Join(a.Input, "trades.csv")
	historicalPrices := path.Join(a.Input, "prices")

	ps, err := BuildPriceService(historicalPrices, a.BaseInst)
	if err != nil {
		panic(err)
	}

	f, err := os.Open(tradesCsv)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	var reportedLots []LotReport
	reportedLots, err = ReadLotReports(path.Join(a.Input, "reported_lots.csv"))
	if err != nil {
		if !os.IsNotExist(err) {
			panic(err)
		}
		log.Println("no reported lots")
	}

	stopDateStr := a.StopDate
	if stopDateStr == "" {
		stopDateStr = time.Now().Format(dateFormat)
	}
	stopDate, err := time.Parse(dateFormat, stopDateStr)
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
		exch := ""
		if len(record) >= 8 {
			exch = record[7]
		}
		entry := Entry{
			ID:       record[0],
			Action:   strings.ToLower(record[1]),
			Date:     record[2],
			Pair:     record[3],
			Amt1:     record[4],
			Amt2:     record[5],
			Fee:      record[6],
			Exchange: exch,
		}

		t := entry.ToTrade()
		if t == nil {
			continue
		}
		if t.BottomInst != a.BaseInst {
			useBottom := false
			px, err := ps.GetHistoricalPrice(t.TopInst, t.Time)
			if err != nil {
				px, err = ps.GetHistoricalPrice(t.BottomInst, t.Time)
				if err != nil {
					fmt.Printf("skipping trade: %s\n", err.Error())
					continue
				}
				useBottom = true
			}
			if !useBottom && t.BottomAmt == 0 {
				fmt.Printf("cannot split trade with 0 bottom amount %s\n", t.ID)
				continue
			}
			trades = append(trades, t.Split(a.BaseInst, px, useBottom)...)
		} else {
			trades = append(trades, t)
		}
	}
	if a.Verbose {
		for _, t := range trades {
			fmt.Println(t.String())
		}
	}

	subLists := map[Instrument][]*Trade{}
	for _, t := range trades {
		inst := t.TopInst
		list, ok := subLists[inst]
		if !ok {
			list = make([]*Trade, 0)
		}
		if !t.Time.After(stopDate) {
			subLists[inst] = append(list, t)
		}
	}

	var newReportedLots *csv.Writer
	if a.WriteReportedLots {
		reportFile, err := os.Create(path.Join(a.OutDir, "reported_lots.csv"))
		if err != nil {
			panic(err)
		}
		defer reportFile.Close()
		newReportedLots = csv.NewWriter(reportFile)
		defer newReportedLots.Flush()
	}

	for inst, subTrades := range subLists {
		func() {
			fmt.Println("INST:", inst, len(subTrades))
			lots, err := MatchTrades(subTrades, !a.FIFO, reportedLots)
			if err != nil {
				panic(err)
			}
			f, err := os.Create(path.Join(a.OutDir, string(inst)+".csv"))
			if err != nil {
				panic(err)
			}
			w := csv.NewWriter(f)
			defer w.Flush()
			w.Write([]string{
				"Purchase Date",
				"Date Sold",
				"Proceeds",
				"Cost Basis",
				"Currency",
				"Description",
			})
			for _, lot := range lots {
				w.Write(lot.CSV())
			}
			if a.WriteReportedLots {
				for _, lot := range lots {
					newReportedLots.Write(lot.ReportedLot())
				}
			}
		}()
	}
}
