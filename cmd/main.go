package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/maerlyn5/polygonio"
)

var (
	cmd        = flag.String("cmd", "last", "")
	ticker     = flag.String("ticker", "", "")
	timespan   = flag.String("timespan", "hour", "")
	multiplier = flag.Int64("multiplier", 1, "")
	search     = flag.String("search", "", "")
)

func PolygonClient() polygonio.PolygonioClient {
	client := polygonio.NewPolygonioClient(os.Getenv("apiKey"), http.DefaultClient)
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	client.Cacher = polygonio.FileCacher{Dir: filepath.Join(wd, "cache"), FileCacherIo: polygonio.OsFileCacherIo{}}
	return client
}

func main() {

	flag.Parse()
	switch *cmd {
	case "last":
		cmdLast()
	case "search":
		cmdSearch()
	default:
		flag.Usage()
	}
}

func cmdLast() {
	client := PolygonClient()
	resp, err := client.LastQuote(context.Background(), polygonio.LastQuoteRequest{
		Ticker: *ticker,
	})
	if err != nil {
		panic(err)
	}

	fmt.Println(resp)
}

func cmdSearch() {

	_search, err := time.Parse(polygonio.StringFormat, *search)
	if err != nil {
		panic(err)
	}

	client := PolygonClient()

	resp, err := client.AggregatesSearch(context.Background(), polygonio.AggregatesRequest{
		Ticker:     *ticker,
		Multiplier: *multiplier,
		Timespan:   *timespan,
		Unadjusted: false,
	}, _search)
	if err != nil {
		panic(err)
	}

	merged := polygonio.Merge(resp)
	fmt.Println(merged.String())
}
