package polygonio

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/shopspring/decimal"
)

type AggregatesRequest struct {
	Ticker     string
	Multiplier int64
	Timespan   string    //minute,hour,day,week,month,quarter
	From       time.Time //pre-market trading (4am est)
	To         time.Time //end of market (4pm est)
	Unadjusted bool
}

func (ar AggregatesRequest) TimespanDuration() time.Duration {
	return TimespanAsDuration(ar.Timespan) * time.Duration(ar.Multiplier)
}

/*
{
  "ticker": "AAPL",
  "status": "OK",
  "queryCount": 801,
  "resultsCount": 16,
  "adjusted": true,
  "results": [
    {
      "v": 24033,
      "vw": 154.23303,
      "o": 154.4,
      "c": 154.7,
      "h": 154.7,
      "l": 153.01,
      "t": 1546419600000,
      "n": 180
    },
...

*/

type AggregatesResponse struct {
	//for some asinine reason they are using scientific notation on relatively small real numbers (even though the docs say int)
	Volume      decimal.Decimal `json:"v"`
	Open        decimal.Decimal `json:"o"`
	Close       decimal.Decimal `json:"c"`
	High        decimal.Decimal `json:"h"`
	Low         decimal.Decimal `json:"l"`
	UnixMiliSec int64           `json:"t"`
	N           int64           `json:"n"`

	//set by Aggregates()
	timespanDuration time.Duration
}

func (ag AggregatesResponse) String() string {
	return fmt.Sprintf("%s-%s (%s) o:%s c:%s h:%s l:%s v:%s",
		ag.UnixMiliSecInTime().Format(StringFormat),
		ag.ImpliedEnd().Format(StringFormat),
		ag.timespanDuration,
		ag.Open.String(),
		ag.Close.String(),
		ag.High.String(),
		ag.Low.String(),
		ag.Volume.String(),
	)
}

func Merge(in []AggregatesResponse) AggregatesResponse {

	if len(in) == 0 {
		panic("expected len > 0")
	}

	if len(in) == 1 {
		return in[0]
	}

	if len(in) != 2 {
		panic("expected len == 2")
	}

	a := in[0]
	b := in[1]

	out := a
	out.Volume = a.Volume.Add(b.Volume)
	out.Close = b.Close
	if b.High.GreaterThan(out.High) {
		out.High = b.High
	}
	if b.Low.LessThan(out.Low) {
		out.Low = b.Low
	}
	out.N += b.N
	out.timespanDuration = b.ImpliedEnd().Sub(a.UnixMiliSecInTime())
	return out
}

func (ar AggregatesResponse) ImpliedEnd() time.Time {
	return ar.UnixMiliSecInTime().Add(ar.timespanDuration)
}

func (ar AggregatesResponse) Average() decimal.Decimal {
	return ar.Open.Add(ar.Close).Div(decimal.NewFromInt(2))
}

//UnixMiliSecInTime <= search < ImpliedEnd
func (ar AggregatesResponse) Contains(search time.Time) bool {
	//!after is the same as before or equal
	return !ar.UnixMiliSecInTime().After(search) && search.Before(ar.ImpliedEnd())
}

func (ag AggregatesResponse) UnixMiliSecInTime() time.Time {
	return time.Unix(0, ag.UnixMiliSec*(int64(time.Millisecond)/int64(time.Nanosecond)))
}

type AggregatesResponseContainer struct {
	Results []AggregatesResponse `json:"results"`
}

func (arc AggregatesResponseContainer) ClosestAggregate(t time.Time) ([]AggregatesResponse, bool) {
	if len(arc.Results) == 0 {
		return nil, false
	}

	i := sort.Search(len(arc.Results), func(i int) bool {
		return arc.Results[i].Contains(t) || !arc.Results[i].UnixMiliSecInTime().Before(t)
	})

	switch i {
	case len(arc.Results):
		return []AggregatesResponse{arc.Results[len(arc.Results)-1]}, arc.Results[len(arc.Results)-1].Contains(t)
	case 0:
		return []AggregatesResponse{arc.Results[0]}, arc.Results[0].Contains(t)
	default:
		found := arc.Results[i]
		if found.Contains(t) {
			return []AggregatesResponse{found}, true
		} else {
			return []AggregatesResponse{arc.Results[i-1], found}, true
		}
	}
}

//https://api.polygon.io/v2/aggs/ticker/AAPL/range/1/hour/2019-01-01/2019-01-02?apiKey=9Z_aU5TfkO48TEa_ji_fYmNuAml4_QeUkpkaao
func (pc PolygonioClient) AggregatesRequest(ctx context.Context, request AggregatesRequest) *http.Request {
	base := pc.URL()
	base.Path = fmt.Sprintf("/v2/aggs/ticker/%s/range/%d/%s/%s/%s", request.Ticker, request.Multiplier, request.Timespan, DateFormat(request.From), DateFormat(request.To))
	q := base.Query()
	q.Add("unadjusted", strconv.FormatBool(request.Unadjusted))
	base.RawQuery = q.Encode()
	req, err := http.NewRequestWithContext(ctx, "GET", base.String(), nil)
	if err != nil {
		panic(err)
	}
	return req
}

func (pc PolygonioClient) Aggregates(ctx context.Context, request AggregatesRequest) (*AggregatesResponseContainer, error) {
	resp, err := DoCache(pc.HTTPClient, pc.AggregatesRequest(ctx, request), true, pc.Cacher)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		out := &AggregatesResponseContainer{}
		bytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(bytes, out); err != nil {
			return nil, err
		}
		for i := range out.Results {
			out.Results[i].timespanDuration = request.TimespanDuration()
		}
		return out, nil
	}
	return nil, StatusError(resp.StatusCode)
}

var LimitExceededError = fmt.Errorf("Search limit exceeded")
var SearchReturnedNoResults = fmt.Errorf("Search returned no results")

//search API until we have a result less than search and we have a result that contains the search date (or two results on either side of it)
func (pc PolygonioClient) AggregatesSearch(ctx context.Context, request AggregatesRequest, search time.Time) ([]AggregatesResponse, error) {
	request.From = FromDate(search)
	request.To = ToDate(search)
	count := 0

	for {
		results, err := pc.Aggregates(ctx, request)
		if err != nil {
			return nil, err
		}

		//fmt.Println(request.From.Format("2006-01-02 3:04:05 PM MST"), request.To.Format("2006-01-02 3:04:05 PM MST"))
		closest, found := results.ClosestAggregate(search)
		if found {
			return closest, nil
		}

		//randomly extend on one direction (this is kind of stretch)
		if len(closest) == 0 {
			if rand.Intn(2) == 0 {
				request.From = request.From.AddDate(0, 0, -1)
			} else {
				request.To = request.To.AddDate(0, 0, 1)
			}
		} else if search.Before(closest[0].UnixMiliSecInTime()) {
			request.From = request.From.AddDate(0, 0, -1)
		} else if search.After(closest[0].UnixMiliSecInTime()) {
			request.To = request.To.AddDate(0, 0, 1)
		}

		count++
		if count > 5 {
			return nil, LimitExceededError
		}
	}
}
