package polygonio

import (
	"context"
	"net/http"
	"net/url"
	"reflect"
	"testing"
	"time"
)

func TestPolygonioClient_HistoricQuotesRequest(t *testing.T) {

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
		request HistoricQuotesRequest
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			fields: fields{APIKey: "apiKey", BaseHost: "base", BaseScheme: "http"},
			args:   args{ctx: context.Background(), request: HistoricQuotesRequest{Ticker: "AAPL", Date: time.Date(2018, 02, 02, 0, 0, 0, 0, anyc)}},
			want:   "http://base/v2/ticks/stocks/nbbo/AAPL/2018-02-02?apiKey=apiKey&limit=0&reverse=false&timestamp=0&timestampLimit=0",
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

			if got := pc.HistoricQuotesRequest(tt.args.ctx, tt.args.request); !reflect.DeepEqual(got.URL, wantUrl) {
				t.Errorf("PolygonioClient.HistoricQuotes() = %v, want %v", got, wantUrl)
			}
		})
	}
}
