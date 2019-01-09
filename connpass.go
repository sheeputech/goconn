package connpass

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const (
	libraryVersion = "1.0.0"
	baseURL        = "https://connpass.com/api/v1/event/"
	userAgent      = "connpass-go/" + libraryVersion
	mediaType      = "application/json"
)

// Client includes information necessary for each request
type Client struct {
	HTTPClient *http.Client
	BaseURL    *url.URL
	UserAgent  string
}

const (
	QueryOrderUpdate Order = 1 + iota // by updated time
	QueryOrderStart                   // by start time
	QueryOrderCreate                  // new arrival order
)

const (
	QueryFormatJSON Format = "json"
)

// QueryParams includes search queries to connpass
type QueryParams struct {
	EventIds             []int
	KeywordsAnd          []string
	KeywordsOr           []string
	Times                []Time
	ParticipantNicknames []string
	OwnerNicknames       []string
	SeriesIds            []int
	Start                int
	Order                Order
	Count                int
	Format               Format
}
type Time struct {
	Year  int
	Month int
	Day   int
}
type Order int
type Format string

// Results is responded contents from the connpass API
type Results struct {
	ResultsReturned  int     `json:"results_returned"`
	ResultsAvailable int     `json:"results_available"`
	ResultsStart     int     `json:"results_start"`
	Events           []Event `json:"events"`
}
type Event struct {
	EventID          int       `json:"event_id"`
	Title            string    `json:"title"`
	Catch            string    `json:"catch"`
	Description      string    `json:"description"`
	EventURL         string    `json:"event_url"`
	HashTag          string    `json:"hash_tag"`
	StartedAt        time.Time `json:"started_at"`
	EndedAt          time.Time `json:"ended_at"`
	Limit            int       `json:"limit"`
	EventType        string    `json:"event_type"`
	Series           Series    `json:"series"`
	Address          string    `json:"address"`
	Place            string    `json:"place"`
	Lat              string    `json:"lat"`
	Lon              string    `json:"lon"`
	OwnerID          int       `json:"owner_id"`
	OwnerNickname    string    `json:"owner_nickname"`
	OwnerDisplayName string    `json:"owner_display_name"`
	Accepted         int       `json:"accepted"`
	Waiting          int       `json:"waiting"`
	UpdatedAt        time.Time `json:"updated_at"`
}
type Series struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	URL   string `json:"url"`
}

// NewClient returns a new connpass API client
func NewClient(httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	baseURL, _ := url.Parse(baseURL)

	c := &Client{HTTPClient: httpClient, BaseURL: baseURL, UserAgent: userAgent}

	return c
}

// buildURL builds query string using user specified QueryParams
func buildURL(q QueryParams) (string, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}

	v := url.Values{}
	setIntValues(v, "event_id", q.EventIds)
	setStringValues(v, "keyword", q.KeywordsAnd)
	setStringValues(v, "keyword_or", q.KeywordsOr)
	setTimeValues(v, q.Times)
	setStringValues(v, "nickname", q.ParticipantNicknames)
	setStringValues(v, "owner_nickname", q.OwnerNicknames)
	setIntValues(v, "series_id", q.SeriesIds)
	if q.Start > 0 {
		v.Set("start", strconv.Itoa(q.Start))
	}
	if q.Order > 0 {
		v.Set("order", fmt.Sprint(q.Order))
	}
	if q.Count > 0 {
		v.Set("count", strconv.Itoa(q.Count))
	}
	if q.Format != "" {
		v.Set("format", fmt.Sprint(q.Format))
	}

	u.RawQuery = v.Encode()

	return u.String(), nil
}

// NewRequest creates an connpass API request
func (c *Client) newRequest(ctx context.Context, method, urlStr string, body interface{}) (*http.Request, error) {
	rel, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	u := c.BaseURL.ResolveReference(rel)

	buf := new(bytes.Buffer)
	if body != nil {
		err = json.NewEncoder(buf).Encode(body)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, u.String(), buf)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", mediaType)
	req.Header.Set("Accept", mediaType)
	req.Header.Set("User-Agent", c.UserAgent+" "+runtime.Version())

	return req, nil
}

// Do sends an connpass API request and return the API response.
func (c *Client) do(ctx context.Context, req *http.Request, v interface{}) (*http.Response, error) {
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// JSON decode
	if err = json.NewDecoder(resp.Body).Decode(v); err != nil {
		return nil, err
	}

	return resp, nil
}

func setIntValues(uv url.Values, k string, v []int) {
	if v != nil {
		ss := []string{}
		for _, iv := range v {
			ss = append(ss, strconv.Itoa(iv))
		}
		uv.Set(k, strings.Join(ss, ","))
	}
}
func setStringValues(uv url.Values, k string, v []string) {
	if v != nil {
		uv.Set(k, strings.Join(v, ","))
	}
}
func setTimeValues(uv url.Values, v []Time) {
	if v != nil && len(v) > 0 {
		ymd := []string{}
		ym := []string{}
		for _, tv := range v {
			if tv.Year > 0 && tv.Month > 0 && tv.Month < 12 {
				if tv.Day > 0 {
					ymd = append(ymd, fmt.Sprintf("%04d%02d%02d", tv.Year, tv.Month, tv.Day))
				} else {
					ym = append(ym, fmt.Sprintf("%04d%02d", tv.Year, tv.Month))
				}
			}
		}
		if len(ymd) > 0 {
			uv.Set("ymd", strings.Join(ymd, ","))
		}
		if len(ym) > 0 {
			uv.Set("ym", strings.Join(ym, ","))
		}
	}
}

// SearchEvents search and return events from connpass API by user specified query params
func (c *Client) SearchEvents(ctx context.Context, q QueryParams) (*Results, error) {
	urlStr, err := buildURL(q)
	if err != nil {
		return nil, err
	}

	req, err := c.newRequest(ctx, http.MethodGet, urlStr, nil)
	if err != nil {
		return nil, err
	}

	var results Results
	_, err = c.do(ctx, req, &results)
	if err != nil {
		return nil, err
	}

	return &results, err
}
