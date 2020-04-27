package polygonio

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/shopspring/decimal"
)

//https://polygon.io/docs/#get_v2_ticks_stocks_nbbo__ticker___date__anchor
type HistoricQuotesRequest struct {
	Ticker         string
	Date           time.Time
	Timestamp      int
	TimestampLimit int
	Reverse        bool
	Limit          int
}

/*
{
  "results": [
    {
      "t": 1517562000065700400,
      "y": 1517562000065321200,
      "q": 2060,
      "c": [
        1
      ],
      "z": 3,
      "p": 102.7,
      "s": 60,
      "x": 11,
      "P": 0,
      "S": 0,
      "X": 0
    },
...

*/

type HistoricQuotesResponse struct {
	BIDPrice            decimal.Decimal `json:"p"`
	AskPrice            decimal.Decimal `json:"P"`
	SipUnixNano         int64           `json:"t"`
	ParticipantUnixNano int64           `json:"y"`
	TRFUnixNano         int64           `json:"f"`
	AskSize             int64           `json:"S"`
	Tap                 int64           `json:"z"`
}

type HistoricQuotesResponseContainer struct {
	Results []HistoricQuotesResponse `json:"results"`
}

func (pc PolygonioClient) HistoricQuotesRequest(ctx context.Context, request HistoricQuotesRequest) *http.Request {
	base := pc.URL()
	base.Path = fmt.Sprintf("/v2/ticks/stocks/nbbo/%s/%s", request.Ticker, DateFormat(request.Date))
	q := base.Query()
	q.Add("timestamp", strconv.Itoa(request.Timestamp))
	q.Add("timestampLimit", strconv.Itoa(request.TimestampLimit))
	q.Add("reverse", strconv.FormatBool(request.Reverse))
	q.Add("limit", strconv.Itoa(request.Limit))
	base.RawQuery = q.Encode()
	req, err := http.NewRequestWithContext(ctx, "GET", base.String(), nil)
	if err != nil {
		panic(err)
	}
	return req
}

func (pc PolygonioClient) HistoricQuotes(ctx context.Context, request HistoricQuotesRequest) (*HistoricQuotesResponseContainer, error) {
	resp, err := DoCache(pc.HTTPClient, pc.HistoricQuotesRequest(ctx, request), true, pc.Cacher)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		out := &HistoricQuotesResponseContainer{}
		bytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		return out, json.Unmarshal(bytes, out)
	}
	return nil, StatusError(resp.StatusCode)
}
