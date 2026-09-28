package fxmacrodata

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// Page selects a slice of a list endpoint. List endpoints return 20 rows by
// default and at most 100 per request, newest first. Request the next page
// with the response's pagination.next_offset while pagination.has_more is true.
type Page struct {
	Limit  int
	Offset int
}

func (p Page) values() url.Values {
	values := url.Values{}
	if p.Limit > 0 {
		values.Set("limit", strconv.Itoa(p.Limit))
	}
	if p.Offset > 0 {
		values.Set("offset", strconv.Itoa(p.Offset))
	}
	return values
}

type Client struct {
	BaseURL    string
	APIKey     string
	HTTPClient *http.Client
}

func NewClient(apiKey string) *Client {
	return &Client{
		BaseURL:    "https://api.fxmacrodata.com/v1",
		APIKey:     apiKey,
		HTTPClient: http.DefaultClient,
	}
}

func (c *Client) DataCatalogue(ctx context.Context, currency string) ([]byte, error) {
	return c.get(ctx, "/data_catalogue/"+norm(currency), nil)
}

func (c *Client) Announcements(ctx context.Context, currency, indicator string) ([]byte, error) {
	return c.get(ctx, "/announcements/"+norm(currency)+"/"+indicator, nil)
}

func (c *Client) AnnouncementsPage(ctx context.Context, currency, indicator string, page Page) ([]byte, error) {
	return c.get(ctx, "/announcements/"+norm(currency)+"/"+indicator, page.values())
}

func (c *Client) Calendar(ctx context.Context, currency string) ([]byte, error) {
	return c.get(ctx, "/calendar/"+norm(currency), nil)
}

func (c *Client) Predictions(ctx context.Context, currency, indicator string) ([]byte, error) {
	return c.get(ctx, "/predictions/"+norm(currency)+"/"+indicator, nil)
}

func (c *Client) PredictionsPage(ctx context.Context, currency, indicator string, page Page) ([]byte, error) {
	return c.get(ctx, "/predictions/"+norm(currency)+"/"+indicator, page.values())
}

func (c *Client) Forex(ctx context.Context, base, quote string) ([]byte, error) {
	return c.get(ctx, "/forex/"+norm(base)+"/"+norm(quote), nil)
}

func (c *Client) ForexPage(ctx context.Context, base, quote string, page Page) ([]byte, error) {
	return c.get(ctx, "/forex/"+norm(base)+"/"+norm(quote), page.values())
}

func (c *Client) COT(ctx context.Context, currency string) ([]byte, error) {
	return c.get(ctx, "/cot/"+norm(currency), nil)
}

func (c *Client) COTPage(ctx context.Context, currency string, page Page) ([]byte, error) {
	return c.get(ctx, "/cot/"+norm(currency), page.values())
}

func (c *Client) CommoditiesLatest(ctx context.Context) ([]byte, error) {
	return c.get(ctx, "/commodities/latest", nil)
}

func (c *Client) Commodity(ctx context.Context, indicator string) ([]byte, error) {
	return c.get(ctx, "/commodities/"+indicator, nil)
}

func (c *Client) CommodityPage(ctx context.Context, indicator string, page Page) ([]byte, error) {
	return c.get(ctx, "/commodities/"+indicator, page.values())
}

func (c *Client) Curves(ctx context.Context, currency string) ([]byte, error) {
	return c.get(ctx, "/curves/"+norm(currency), nil)
}

func (c *Client) CurveProxies(ctx context.Context, currency string) ([]byte, error) {
	return c.get(ctx, "/curve_proxies/"+norm(currency), nil)
}

func (c *Client) ForwardCurves(ctx context.Context, currency string) ([]byte, error) {
	return c.get(ctx, "/forward_curves/"+norm(currency), nil)
}

func (c *Client) MarketSessions(ctx context.Context) ([]byte, error) {
	return c.get(ctx, "/market_sessions", nil)
}

func (c *Client) RiskSentiment(ctx context.Context) ([]byte, error) {
	return c.get(ctx, "/risk_sentiment", nil)
}

func (c *Client) News(ctx context.Context, currency string) ([]byte, error) {
	return c.get(ctx, "/news/"+norm(currency), nil)
}

func (c *Client) PressReleases(ctx context.Context, currency string) ([]byte, error) {
	return c.get(ctx, "/press-releases/"+norm(currency), nil)
}

func (c *Client) PressReleasesPage(ctx context.Context, currency string, page Page) ([]byte, error) {
	return c.get(ctx, "/press-releases/"+norm(currency), page.values())
}

func (c *Client) buildURL(path string, values url.Values) string {
	base := strings.TrimRight(c.BaseURL, "/")
	if len(values) == 0 {
		return base + path
	}
	return base + path + "?" + values.Encode()
}

func (c *Client) get(ctx context.Context, path string, values url.Values) ([]byte, error) {
	httpClient := c.HTTPClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.buildURL(path, values), nil)
	if err != nil {
		return nil, err
	}
	if c.APIKey != "" {
		req.Header.Set("X-API-Key", c.APIKey)
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("fxmacrodata: status %d", resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}

func norm(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}
