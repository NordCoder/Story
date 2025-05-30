package wikipedia

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/NordCoder/Story/internal/entity"
	"github.com/cenkalti/backoff/v4"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

// todo implement wiki-mock before first start (it could be another service)
// todo redesign to make this client return only raw data and create another entity to parse it

type Thumbnail struct {
	Source string `json:"source"`
}

// Default configuration constants
const (
	defaultAPIURL           = "https://ru.wikipedia.org/w/api.php"
	defaultUserAgent        = "story/1.0 (dimakorzh2005@gmail.com) go-http-client"
	defaultCategoryPageSize = 100
	defaultMaxLag           = 5
	defaultTimeout          = 15 * time.Second
	defaultMaxRetries       = 3
)

// ToFact конвертирует ArticleSummary в entity.Fact.
func (a *ArticleSummary) ToFact(category entity.Category) *entity.Fact {
	return &entity.Fact{
		ID:        entity.NewFactID(), // создаём новый уникальный ID
		Title:     a.Title,
		Category:  category,
		Summary:   trimSummary(a.Extract), // обрезаем до 280 символов
		ImageURL:  a.ImageURL,
		SourceURL: a.PageURL,
		FetchedAt: time.Now(), // ставим текущее время
	}
}

// trimSummary обрезает текст до 280 символов.
func trimSummary(s string) string {
	if len(s) <= 280 {
		return s
	}
	return s[:280]
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
func NewClient(opts ...Option) WikiClient {
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

	//c.logger.Info("wikiapi endpoint", zap.String("url", endpoint))

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

		var bodyReader io.Reader = resp.Body
		if resp.Header.Get("Content-Encoding") == "gzip" {
			gzipReader, err := gzip.NewReader(resp.Body)
			if err != nil {
				c.logger.Error("failed to create gzip reader", zap.Error(err))
				return backoff.Permanent(err)
			}
			defer gzipReader.Close()
			bodyReader = gzipReader
		}

		decoder := json.NewDecoder(bodyReader)
		if err := decoder.Decode(out); err != nil {
			c.logger.Error("json decode failed", zap.Error(err))
			return backoff.Permanent(err)
		}
		return nil

		//rawBody, _ := io.ReadAll(bodyReader)
		//c.logger.Info("raw JSON", zap.ByteString("json", rawBody))
		//
		//// временно прерываем выполнение, чтобы просто увидеть JSON:
		//return backoff.Permanent(fmt.Errorf("debug json output"))
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

// isValidArticle filters out articles that are likely lists or lack images
func isValidArticle(title, extract string, thumb *Thumbnail) bool {
	t := strings.ToLower(title)
	e := strings.ToLower(extract)

	if strings.Contains(t, "список") || strings.Contains(t, "обзор") {
		return false
	}
	if strings.Contains(e, "список") || strings.Contains(e, "это список") {
		return false
	}
	if thumb == nil || thumb.Source == "" {
		return false
	}
	return true
}

// GetCategorySummaries retrieves up to limit summaries via generator=categorymembers
func (c *Client) GetCategorySummaries(ctx context.Context, category entity.Category, limit int) ([]*ArticleSummary, error) {
	params := url.Values{
		"action":       {"query"},
		"generator":    {"categorymembers"},
		"gcmtitle":     {"Category:" + string(category)},
		"gcmnamespace": {"0"},
		"gcmlimit":     {fmt.Sprint(limit)},
		"prop":         {"extracts|pageimages|info"},
		"inprop":       {"url"},
		"exintro":      {"true"},
		"explaintext":  {"true"},
		"piprop":       {"thumbnail"},
		"pithumbsize":  {fmt.Sprint(1500)},
	}

	var resp struct {
		Query struct {
			Pages map[string]struct {
				Title        string     `json:"title"`
				Extract      string     `json:"extract"`
				Thumbnail    *Thumbnail `json:"thumbnail,omitempty"`
				CanonicalURL string     `json:"canonicalurl"`
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

	summaries := make([]*ArticleSummary, 0, len(resp.Query.Pages))
	for _, p := range resp.Query.Pages {
		if !isValidArticle(p.Title, p.Extract, p.Thumbnail) {
			continue
		}

		pageURL := p.CanonicalURL
		if pageURL == "" {
			pageURL = fmt.Sprintf("https://ru.wikipedia.org/wiki/%s", url.PathEscape(p.Title))
		}
		imageURL := ""
		if p.Thumbnail != nil {
			imageURL = p.Thumbnail.Source
		}
		summaries = append(summaries, &ArticleSummary{
			Title:    p.Title,
			Category: category,
			Extract:  p.Extract,
			ImageURL: imageURL,
			PageURL:  pageURL,
		})
	}

	c.logger.Info("fetched category summaries", zap.String("category", string(category)), zap.Int("count", len(summaries)))
	return summaries, nil
}

func (c *Client) Ping(ctx context.Context) error {
	// Мы не реально проверяем Wikipedia API, поэтому считаем, что всё ок.
	// Можно позже реализовать реальный ping через запрос siteinfo.
	return nil
}

// GetSubcategories retrieves up to limit subcategories of the given category title.
func (c *Client) GetSubcategories(ctx context.Context, category entity.Category, limit int) ([]entity.Category, error) {
	params := url.Values{
		"action":  {"query"},
		"list":    {"categorymembers"},
		"cmtitle": {"Category:" + string(category)},
		"cmtype":  {"subcat"},
		"cmlimit": {fmt.Sprint(limit)},
	}

	var resp struct {
		Query struct {
			CategoryMembers []struct {
				Title string `json:"title"`
			} `json:"categorymembers"`
		} `json:"query"`
		Error *struct{ Code, Info string } `json:"error,omitempty"`
	}

	if err := c.doRequest(ctx, params, &resp); err != nil {
		return nil, err
	}
	if resp.Error != nil {
		return nil, fmt.Errorf("wikiapi error %s: %s", resp.Error.Code, resp.Error.Info)
	}

	names := make([]entity.Category, len(resp.Query.CategoryMembers))
	for i, m := range resp.Query.CategoryMembers {
		// strip the "Category:" prefix
		names[i] = entity.Category(strings.TrimPrefix(m.Title, "Category:"))
	}
	return names, nil
}
