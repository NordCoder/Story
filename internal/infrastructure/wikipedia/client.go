package wikipedia

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cenkalti/backoff/v4"
	"io"
	"net/http"
	"net/url"
	"sort"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

// Default configuration constants
const (
	defaultAPIURL           = "https://en.wikipedia.org/w/api.php"
	defaultUserAgent        = "bigtech-app/1.0 (https://example.com; dev@example.com) go-http-client"
	defaultCategoryPageSize = 100
	defaultMaxLag           = 5
	defaultTimeout          = 15 * time.Second
	defaultMaxRetries       = 3
)

// ErrNoPages indicates that the API returned no pages
var ErrNoPages = errors.New("wikiapi: no pages returned")

// ArticleSummary represents a Wikipedia page summary
type ArticleSummary struct {
	Title    string
	Extract  string
	ImageURL string
	PageURL  string
}

// HTTPClient defines the minimal interface for making HTTP requests
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// Client provides methods to interact with the MediaWiki Action API
type Client struct {
	httpClient  HTTPClient
	apiURL      string
	userAgent   string
	pageSize    int
	maxLag      int
	retryPolicy backoff.BackOff
	logger      *zap.Logger
	// Metrics
	requestCount    prometheus.Counter
	requestDuration prometheus.Histogram
}

// Option configures the Client
type Option func(*Client)

// WithHTTPClient sets a custom HTTP client
func WithHTTPClient(h HTTPClient) Option {
	return func(c *Client) { c.httpClient = h }
}

// WithAPIURL overrides the API endpoint URL
func WithAPIURL(u string) Option {
	return func(c *Client) { c.apiURL = u }
}

// WithUserAgent sets a descriptive User-Agent
func WithUserAgent(ua string) Option {
	return func(c *Client) { c.userAgent = ua }
}

// WithPageSize sets the page size for category queries (max 500 if bot)
func WithPageSize(size int) Option {
	return func(c *Client) { c.pageSize = size }
}

// WithMaxLag sets the maxlag parameter
func WithMaxLag(sec int) Option {
	return func(c *Client) { c.maxLag = sec }
}

// WithLogger injects a zap.Logger for structured logging
func WithLogger(logger *zap.Logger) Option {
	return func(c *Client) { c.logger = logger }
}

// WithMetrics registers Prometheus metrics for API calls
func WithMetrics(reg prometheus.Registerer) Option {
	return func(c *Client) {
		c.requestCount = prometheus.NewCounter(prometheus.CounterOpts{
			Name: "wikiapi_requests_total",
			Help: "Total number of Wikipedia API requests",
		})
		c.requestDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "wikiapi_request_duration_seconds",
			Help:    "Histogram of API request durations",
			Buckets: prometheus.DefBuckets,
		})
		reg.MustRegister(c.requestCount, c.requestDuration)
	}
}

// NewClient creates a production-ready Client with sane defaults
func NewClient(opts ...Option) *Client {
	// Default exponential backoff
	retryPolicy := backoff.WithMaxRetries(backoff.NewExponentialBackOff(), defaultMaxRetries)
	client := &Client{
		httpClient:      &http.Client{Timeout: defaultTimeout},
		apiURL:          defaultAPIURL,
		userAgent:       defaultUserAgent,
		pageSize:        defaultCategoryPageSize,
		maxLag:          defaultMaxLag,
		retryPolicy:     retryPolicy,
		logger:          zap.NewNop(),
		requestCount:    prometheus.NewCounter(prometheus.CounterOpts{Name: "wikiapi_requests_total", Help: ""}),
		requestDuration: prometheus.NewHistogram(prometheus.HistogramOpts{Name: "wikiapi_request_duration_seconds", Help: ""}),
	}
	for _, opt := range opts {
		opt(client)
	}
	return client
}

// doRequest executes a GET to the API with retry, metrics, and structured logging
func (c *Client) doRequest(ctx context.Context, params url.Values, out interface{}) error {
	// Add common params
	params.Set("format", "json")
	params.Set("maxlag", fmt.Sprint(c.maxLag))
	endpoint := fmt.Sprintf("%s?%s", c.apiURL, params.Encode())

	var lastErr error
	start := time.Now()
	err := backoff.Retry(func() error {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
		if err != nil {
			return backoff.Permanent(err)
		}
		req.Header.Set("User-Agent", c.userAgent)
		req.Header.Set("Accept-Encoding", "gzip")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			c.logger.Warn("request failed", zap.Error(err))
			lastErr = err
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			err := fmt.Errorf("status=%d, body=%s", resp.StatusCode, string(body))
			c.logger.Warn("unexpected status", zap.Int("code", resp.StatusCode), zap.ByteString("body", body))
			lastErr = err
			return err
		}
		decoder := json.NewDecoder(resp.Body)
		if err := decoder.Decode(out); err != nil {
			c.logger.Error("json decode failed", zap.Error(err))
			return backoff.Permanent(err)
		}
		return nil
	}, c.retryPolicy)

	duration := time.Since(start).Seconds()
	c.requestCount.Inc()
	c.requestDuration.Observe(duration)

	if err != nil {
		c.logger.Error("all retries failed", zap.Duration("duration", time.Since(start)), zap.Error(lastErr))
		return fmt.Errorf("wikiapi: %w", lastErr)
	}
	return nil
}

// GetCategorySummaries retrieves up to limit summaries via generator=categorymembers
func (c *Client) GetCategorySummaries(ctx context.Context, category string, limit int) ([]*ArticleSummary, error) {
	params := url.Values{
		"action":       {"query"},
		"generator":    {"categorymembers"},
		"gcmtitle":     {"Category:" + category},
		"gcmnamespace": {"0"},
		"gcmlimit":     {fmt.Sprint(limit)},
		"prop":         {"extracts|pageimages"},
		"exintro":      {"true"},
		"explaintext":  {"true"},
		"piprop":       {"thumbnail"},
		"pithumbsize":  {fmt.Sprint(100)},
	}

	var resp struct {
		Query struct {
			Pages map[string]struct {
				Title     string `json:"title"`
				Extract   string `json:"extract"`
				Thumbnail struct {
					Source string `json:"source"`
				} `json:"thumbnail"`
				CanonicalURL string `json:"canonicalurl"`
			} `json:"pages"`
		} `json:"query"`
		Error *struct{ Code, Info string } `json:"error,omitempty"`
	}

	if err := c.doRequest(ctx, params, &resp); err != nil {
		return nil, err
	}
	if resp.Error != nil {
		return nil, fmt.Errorf("wikiapi error %s: %s", resp.Error.Code, resp.Error.Info)
	}

	if len(resp.Query.Pages) == 0 {
		return nil, ErrNoPages
	}

	// Convert map to slice and sort by title for determinism
	keys := make([]string, 0, len(resp.Query.Pages))
	for _, p := range resp.Query.Pages {
		keys = append(keys, p.Title)
	}
	sort.Strings(keys)

	summaries := make([]*ArticleSummary, 0, len(keys))
	for _, title := range keys {
		p := resp.Query.Pages[title]
		pageURL := p.CanonicalURL
		if pageURL == "" {
			pageURL = fmt.Sprintf("https://en.wikipedia.org/wiki/%s", url.PathEscape(p.Title))
		}
		summaries = append(summaries, &ArticleSummary{
			Title:    p.Title,
			Extract:  p.Extract,
			ImageURL: p.Thumbnail.Source,
			PageURL:  pageURL,
		})
	}

	c.logger.Info("fetched category summaries", zap.String("category", category), zap.Int("count", len(summaries)))
	return summaries, nil
}
