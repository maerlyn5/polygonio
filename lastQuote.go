package polygonio

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/shopspring/decimal"
)

//https://polygon.io/docs/#get_v1_last_quote_stocks__symbol__anchor
type LastQuoteRequest struct {
	Ticker string
}

/*
{
  "status": "success",
  "symbol": "AAPL",
  "last": {
    "askprice": 159.59,
    "asksize": 2,
    "askexchange": 11,
    "bidprice": 159.45,
    "bidsize": 20,
    "bidexchange": 12,
    "timestamp": 1518086601843
  }
}
*/

type LastQuoteResponse struct {
	AskPrice    decimal.Decimal `json:"askprice"`
	AskSize     int64           `json:"asksize"`
	BidPrice    decimal.Decimal `json:"bidprice"`
	BidSize     int64           `json:"bidsize"`
	UnixMiliSec int64           `json:"timestamp"`
}

func (lr LastQuoteResponse) UnixMiliSecInTime() time.Time {
	return time.Unix(0, lr.UnixMiliSec*(int64(time.Millisecond)/int64(time.Nanosecond)))
}

func (lr LastQuoteResponse) Market() decimal.Decimal {
	return lr.BidPrice.Add(lr.AskPrice).Div(decimal.NewFromInt(2))
}

type LastQuoteResponseContainer struct {
	Last LastQuoteResponse `json:"last"`
}

///v1/last_quote/stocks/{symbol}
func (pc PolygonioClient) LastQuoteRequest(ctx context.Context, request LastQuoteRequest) *http.Request {
	base := pc.URL()
	base.Path = fmt.Sprintf("/v1/last_quote/stocks/%s", request.Ticker)
	req, err := http.NewRequestWithContext(ctx, "GET", base.String(), nil)
	if err != nil {
		panic(err)
	}
	return req
}

func (pc PolygonioClient) LastQuote(ctx context.Context, request LastQuoteRequest) (*LastQuoteResponseContainer, error) {
	resp, err := DoCache(pc.HTTPClient, pc.LastQuoteRequest(ctx, request), false, pc.Cacher)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		out := &LastQuoteResponseContainer{}
		bytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		return out, json.Unmarshal(bytes, out)
	}
	return nil, StatusError(resp.StatusCode)
}
