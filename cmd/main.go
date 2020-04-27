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
	ticker     = flag.String("ticker", "", "")
	timespan   = flag.String("timespan", "hour", "")
	multiplier = flag.Int64("multiplier", 1, "")
	search     = flag.String("search", "", "")
)

func main() {

	flag.Parse()
	_search, err := time.Parse("2006-01-02 3:04:05 PM MST", *search)
	if err != nil {
		panic(err)
	}

	client := polygonio.NewPolygonioClient(os.Getenv("apiKey"), http.DefaultClient)

	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	client.Cacher = polygonio.FileCacher{Dir: filepath.Join(wd, "cache"), FileCacherIo: polygonio.OsFileCacherIo{}}

	resp, err := client.AggregatesSearch(context.Background(), polygonio.AggregatesRequest{
		Ticker:     *ticker,
		Multiplier: *multiplier,
		Timespan:   *timespan,
		Unadjusted: false,
	}, _search)
	if err != nil {
		panic(err)
	}

	for _, tr := range resp {
		fmt.Println(tr.UnixMiliSec, tr.UnixMiliSecInTime().Format("2006-01-02 3:04:05 PM MST"))
	}
}
