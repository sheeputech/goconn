package connpassgo

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"
)

var (
	mux    *http.ServeMux
	ctx    = context.TODO()
	client *Client
	server *httptest.Server
)

func setup() {
	mux = http.NewServeMux()
	server = httptest.NewServer(mux)

	client = NewClient(nil)
	url, _ := url.Parse(server.URL)
	client.BaseURL = url
}
func closeServer() {
	server.Close()
}

func TestNewClient(t *testing.T) {
	t.Helper()

	c := NewClient(nil)
	testClientDefaults(t, c)
}
func testClientDefaults(t *testing.T, c *Client) {
	t.Helper()

	if c.BaseURL == nil || c.BaseURL.String() != baseURL {
		t.Errorf("NewClient BaseURL: got %v, expected %v", c.BaseURL, baseURL)
	}
	if c.UserAgent != userAgent {
		t.Errorf("NewClient UserAgent: got %v, expected %v", c.UserAgent, userAgent)
	}
}

func TestBuildURL_noQuery(t *testing.T) {
	t.Helper()

	qp := QueryParams{}
	s, err := buildURL(qp)
	if err != nil {
		t.Fatalf("baseURL is invalid: %s", err)
	}
	u, err := url.Parse(s)
	if err != nil {
		t.Errorf("built URL is invalid: %s", err)
	}
	if u.RawQuery != "" {
		t.Errorf("Query should be empty: got %s, expected \"\"", u)
	}
}
func TestBuildURL_withQuery(t *testing.T) {
	t.Helper()

	qp := QueryParams{}
	qp.EventIds = []int{10, 20}
	qp.KeywordsAnd = []string{"a", "b"}
	qp.KeywordsOr = []string{"c", "d"}
	qp.Times = []Time{
		Time{Year: 2018, Month: 02, Day: 10},
		Time{Year: 2019, Month: 3, Day: 11},
		Time{Year: 2020, Month: 4},
	}
	qp.ParticipantNicknames = []string{"foo", "bar"}
	qp.OwnerNicknames = []string{"foo", "bar", "baz"}
	qp.SeriesIds = []int{1000, 2000, 5000}
	qp.Start = 5
	qp.Order = QueryOrderUpdate
	qp.Count = 100
	qp.Format = QueryFormatJSON

	s, _ := buildURL(qp)
	u, err := url.Parse(s)
	if err != nil {
		t.Fatalf("Built URL is invalid: %s", err)
	}
	testQueryParam(t, u, "event_id", "10,20")
	testQueryParam(t, u, "keyword", "a,b")
	testQueryParam(t, u, "keyword_or", "c,d")
	testQueryParam(t, u, "ym", "202004")
	testQueryParam(t, u, "ymd", "20180210,20190311")
	testQueryParam(t, u, "nickname", "foo,bar")
	testQueryParam(t, u, "owner_nickname", "foo,bar,baz")
	testQueryParam(t, u, "series_id", "1000,2000,5000")
	testQueryParam(t, u, "start", "5")
	testQueryParam(t, u, "order", "1")
	testQueryParam(t, u, "count", "100")
	testQueryParam(t, u, "format", "json")
}
func testQueryParam(t *testing.T, u *url.URL, key string, expected string) {
	v := u.Query()[key]
	if !reflect.DeepEqual(expected, v[0]) {
		t.Errorf("QueryParams.%s is invalid: got %s, expected %s", key, v[0], expected)
	}
}

func TestNewRequest(t *testing.T) {
	t.Helper()
}

func TestDo(t *testing.T) {
	t.Helper()

	setup()
	defer closeServer()

	type test struct {
		FOO string
	}

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if m := http.MethodGet; m != r.Method {
			t.Errorf("Request method is invalid: got %v, expected %v", r.Method, m)
		}
		fmt.Fprint(w, `{"FOO":"bar"}`)
	})

	req, _ := client.newRequest(ctx, http.MethodGet, "/", nil)
	body := new(test)
	_, err := client.do(context.Background(), req, body)
	if err != nil {
		t.Fatalf("Do(): %v", err)
	}

	expected := &test{"bar"}
	if !reflect.DeepEqual(body, expected) {
		t.Errorf("Response body is invalid: got %v, expected %v", body, expected)
	}
}

func TestSearchEvents(t *testing.T) {
	t.Helper()
}
