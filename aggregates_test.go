package polygonio

import (
	"context"
	"net/http"
	"net/url"
	"reflect"
	"testing"
	"time"

	"github.com/shopspring/decimal"
)

func TestPolygonioClient_AggregatesRequest(t *testing.T) {

	anyc, err := time.LoadLocation("America/New_York")
	if err != nil {
		panic(err)
	}

	type fields struct {
		HTTPClient *http.Client
		APIKey     string
		BaseHost   string
		BaseScheme string
	}
	type args struct {
		ctx     context.Context
		request AggregatesRequest
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			fields: fields{APIKey: "apiKey", BaseHost: "base", BaseScheme: "http"},
			args: args{ctx: context.Background(), request: AggregatesRequest{
				Ticker:     "AAPL",
				Multiplier: 1,
				Timespan:   "day",
				From:       time.Date(2019, 01, 01, 0, 0, 0, 0, anyc),
				To:         time.Date(2019, 01, 02, 0, 0, 0, 0, anyc),
				Unadjusted: false,
			}},
			want: "http://base/v2/aggs/ticker/AAPL/range/1/day/2019-01-01/2019-01-02?apiKey=apiKey&unadjusted=false",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pc := PolygonioClient{
				HTTPClient: tt.fields.HTTPClient,
				APIKey:     tt.fields.APIKey,
				BaseHost:   tt.fields.BaseHost,
				BaseScheme: tt.fields.BaseScheme,
			}

			wantUrl, err := url.Parse(tt.want)
			if err != nil {
				panic(err)
			}

			if got := pc.AggregatesRequest(tt.args.ctx, tt.args.request); !reflect.DeepEqual(got.URL, wantUrl) {
				t.Errorf("PolygonioClient.HistoricQuotes() = %v, want %v", got, wantUrl)
			}
		})
	}
}

func TestAggregatesResponse_Contains(t *testing.T) {
	type fields struct {
		Volume           decimal.Decimal
		Open             decimal.Decimal
		Close            decimal.Decimal
		High             decimal.Decimal
		Low              decimal.Decimal
		UnixMiliSec      int64
		N                int64
		timespanDuration time.Duration
	}
	type args struct {
		search time.Time
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			fields: fields{UnixMiliSec: 1549011600000, timespanDuration: time.Minute},
			args:   args{search: time.Unix(0, 1549011600000*(int64(time.Millisecond)/int64(time.Nanosecond)))},
			want:   true,
		},
		{
			fields: fields{UnixMiliSec: 1549011600000, timespanDuration: time.Minute},
			args:   args{search: time.Unix(0, 1549011600000*(int64(time.Millisecond)/int64(time.Nanosecond))).Add(time.Minute).Add(-time.Second)},
			want:   true,
		},
		{
			fields: fields{UnixMiliSec: 1549011600000, timespanDuration: time.Minute},
			args:   args{search: time.Unix(0, 1549011600000*(int64(time.Millisecond)/int64(time.Nanosecond))).Add(time.Minute)},
			want:   false,
		},
		{
			fields: fields{UnixMiliSec: 1549011600000, timespanDuration: time.Minute},
			args:   args{search: time.Unix(0, 1549011500000*(int64(time.Millisecond)/int64(time.Nanosecond)))},
			want:   false,
		},
		{
			fields: fields{UnixMiliSec: 1549011600000, timespanDuration: time.Minute},
			args:   args{search: time.Unix(0, 1549011600000*(int64(time.Millisecond)/int64(time.Nanosecond))).Add(time.Minute).Add(time.Second)},
			want:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ar := AggregatesResponse{
				Volume:           tt.fields.Volume,
				Open:             tt.fields.Open,
				Close:            tt.fields.Close,
				High:             tt.fields.High,
				Low:              tt.fields.Low,
				UnixMiliSec:      tt.fields.UnixMiliSec,
				N:                tt.fields.N,
				timespanDuration: tt.fields.timespanDuration,
			}
			if got := ar.Contains(tt.args.search); got != tt.want {
				t.Errorf("AggregatesResponse.Contains() = %v, want %v", got, tt.want)
			}
		})
	}
}

/*

 */

func TestAggregatesResponseContainer_ClosestAggregate(t *testing.T) {

	results := []AggregatesResponse{
		{timespanDuration: time.Hour, UnixMiliSec: 1546297200000}, //2018-12-31 3:00:00 PM PST
		{timespanDuration: time.Hour, UnixMiliSec: 1546300800000}, //2018-12-31 4:00:00 PM PST
		{timespanDuration: time.Hour, UnixMiliSec: 1546419600000}, //2019-01-02 1:00:00 AM PST
		{timespanDuration: time.Hour, UnixMiliSec: 1546423200000}, //2019-01-02 2:00:00 AM PST
	}

	mustTime := func(f string) time.Time {
		a, e := time.Parse("2006-01-02 3:04:05 PM MST", f)
		if e != nil {
			panic(e)
		}
		return a
	}

	type fields struct {
		Results []AggregatesResponse
	}
	type args struct {
		t time.Time
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []AggregatesResponse
		want1  bool
	}{
		{
			name:   "before",
			fields: fields{Results: results},
			args:   args{t: mustTime("2018-12-31 2:00:00 PM PST")},
			want:   []AggregatesResponse{{timespanDuration: time.Hour, UnixMiliSec: 1546297200000}},
			want1:  false,
		},
		{
			name:   "start inclusive",
			fields: fields{Results: results},
			args:   args{t: mustTime("2018-12-31 3:00:00 PM PST")},
			want:   []AggregatesResponse{{timespanDuration: time.Hour, UnixMiliSec: 1546297200000}},
			want1:  true,
		},
		{
			name:   "start contains",
			fields: fields{Results: results},
			args:   args{t: mustTime("2018-12-31 3:00:05 PM PST")},
			want:   []AggregatesResponse{{timespanDuration: time.Hour, UnixMiliSec: 1546297200000}},
			want1:  true,
		},
		{
			name:   "border of two slots",
			fields: fields{Results: results},
			args:   args{t: mustTime("2018-12-31 4:00:00 PM PST")},
			want:   []AggregatesResponse{{timespanDuration: time.Hour, UnixMiliSec: 1546300800000}},
			want1:  true,
		},
		{
			name:   "in between",
			fields: fields{Results: results},
			args:   args{t: mustTime("2018-12-31 5:00:00 PM PST")},
			want:   []AggregatesResponse{{timespanDuration: time.Hour, UnixMiliSec: 1546300800000}, {timespanDuration: time.Hour, UnixMiliSec: 1546419600000}},
			want1:  true,
		},
		{
			name:   "after",
			fields: fields{Results: results},
			args:   args{t: mustTime("2019-01-02 3:00:00 AM PST")},
			want:   []AggregatesResponse{{timespanDuration: time.Hour, UnixMiliSec: 1546423200000}},
			want1:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			arc := AggregatesResponseContainer{
				Results: tt.fields.Results,
			}
			got, got1 := arc.ClosestAggregate(tt.args.t)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AggregatesResponseContainer.ClosestAggregate() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("AggregatesResponseContainer.ClosestAggregate() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
