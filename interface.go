package polygonio

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type Polygonio interface {
	HistoricQuotes(ctx context.Context, request HistoricQuotesRequest) (*HistoricQuotesResponseContainer, error)
}

type Cacher interface {
	//should only return error if response is unuseable
	Save(request *http.Request, response *http.Response) error
	Get(request *http.Request) (*http.Response, error)
}

func NewPolygonioClient(APIKey string, HTTPClient *http.Client) PolygonioClient {
	return PolygonioClient{APIKey: APIKey, HTTPClient: HTTPClient, BaseHost: "api.polygon.io", BaseScheme: "https"}
}

type PolygonioClient struct {
	HTTPClient *http.Client
	APIKey     string
	BaseHost   string
	BaseScheme string
	Cacher     Cacher
}

type StatusError int

func (st StatusError) Error() string {
	return "htttpStatus:" + strconv.Itoa(int(st))
}

func DateFormat(time time.Time) string {
	return time.Format("2006-01-02")
}

func (pc PolygonioClient) URL() url.URL {
	u := url.URL{
		Scheme: pc.BaseScheme,
		Host:   pc.BaseHost,
	}

	q := u.Query()
	q.Add("apiKey", pc.APIKey)
	u.RawQuery = q.Encode()

	return u
}
