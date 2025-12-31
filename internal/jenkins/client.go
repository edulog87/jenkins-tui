// Package jenkins provides the HTTP client for interacting with Jenkins API.
package jenkins

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/elogrono/jenkins-tui/internal/config"
	"github.com/elogrono/jenkins-tui/internal/logger"
	"golang.org/x/time/rate"
)

// Client is the Jenkins API client
type Client struct {
	baseURL    string
	username   string
	apiToken   string
	httpClient *http.Client
	limiter    *rate.Limiter

	// Crumb for CSRF protection
	crumb       string
	crumbField  string
	crumbMu     sync.RWMutex
	crumbTested bool
}

// NewClient creates a new Jenkins client
func NewClient(cfg *config.Config) (*Client, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is nil")
	}

	baseURL := strings.TrimSuffix(cfg.Profile.BaseURL, "/")

	logger.Debug("Creating Jenkins client",
		"baseURL", baseURL,
		"username", cfg.Profile.Username,
		"timeout", cfg.Profile.TimeoutSeconds,
		"rateLimit", cfg.Profile.RateLimitRPS,
		"insecureTLS", cfg.Profile.InsecureSkipTLSVerify,
	)

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: cfg.Profile.InsecureSkipTLSVerify,
		},
		MaxIdleConns:        10,
		IdleConnTimeout:     30 * time.Second,
		DisableCompression:  false,
		MaxIdleConnsPerHost: 5,
	}

	httpClient := &http.Client{
		Transport: transport,
		Timeout:   time.Duration(cfg.Profile.TimeoutSeconds) * time.Second,
	}

	limiter := rate.NewLimiter(rate.Limit(cfg.Profile.RateLimitRPS), cfg.Profile.RateLimitRPS)

	logger.Info("Jenkins client created successfully")

	return &Client{
		baseURL:    baseURL,
		username:   cfg.Profile.Username,
		apiToken:   cfg.Profile.APIToken,
		httpClient: httpClient,
		limiter:    limiter,
	}, nil
}

// TestConnection tests the connection to Jenkins
func (c *Client) TestConnection() error {
	logger.Info("Testing connection to Jenkins", "baseURL", c.baseURL)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	start := time.Now()
	info, err := c.GetRootInfo(ctx)
	elapsed := time.Since(start)

	if err != nil {
		logger.Error("Connection test failed",
			"error", err,
			"elapsed", elapsed,
		)
		return err
	}

	logger.Info("Connection test successful",
		"elapsed", elapsed,
		"mode", info.Mode,
		"executors", info.NumExecutors,
		"useCrumbs", info.UseCrumbs,
	)

	return nil
}

// doRequest performs an HTTP request with authentication and rate limiting
func (c *Client) doRequest(ctx context.Context, method, path string, body io.Reader) (*http.Response, error) {
	// Ensure context is not nil
	if ctx == nil {
		ctx = context.Background()
	}

	fullURL := c.baseURL + path

	logger.Debug("Making HTTP request",
		"method", method,
		"url", fullURL,
	)

	// Wait for rate limiter
	start := time.Now()
	if err := c.limiter.Wait(ctx); err != nil {
		logger.Error("Rate limiter error", "error", err)
		return nil, fmt.Errorf("rate limiter error: %w", err)
	}
	waitTime := time.Since(start)
	if waitTime > 100*time.Millisecond {
		logger.Debug("Rate limiter wait", "duration", waitTime)
	}

	req, err := http.NewRequestWithContext(ctx, method, fullURL, body)
	if err != nil {
		logger.Error("Error creating request", "error", err)
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	// Set basic auth
	req.SetBasicAuth(c.username, c.apiToken)

	// Add crumb if available (for POST requests)
	if method != http.MethodGet {
		c.crumbMu.RLock()
		if c.crumb != "" && c.crumbField != "" {
			req.Header.Set(c.crumbField, c.crumb)
			logger.Debug("Added crumb header", "field", c.crumbField)
		}
		c.crumbMu.RUnlock()
	}

	req.Header.Set("Accept", "application/json")

	reqStart := time.Now()
	resp, err := c.httpClient.Do(req)
	reqElapsed := time.Since(reqStart)

	if err != nil {
		logger.Error("HTTP request failed",
			"error", err,
			"elapsed", reqElapsed,
			"url", fullURL,
		)
		return nil, fmt.Errorf("error making request: %w", err)
	}

	logger.Debug("HTTP response received",
		"status", resp.StatusCode,
		"elapsed", reqElapsed,
		"url", fullURL,
	)

	// Check for auth errors
	if resp.StatusCode == http.StatusUnauthorized {
		resp.Body.Close()
		logger.Error("Authentication failed", "status", resp.StatusCode)
		return nil, fmt.Errorf("authentication failed: invalid credentials")
	}

	if resp.StatusCode == http.StatusForbidden {
		resp.Body.Close()
		logger.Warn("Access forbidden", "status", resp.StatusCode, "crumbTested", c.crumbTested)
		// Try to fetch crumb if not done yet
		if !c.crumbTested && method != http.MethodGet {
			logger.Info("Attempting to fetch crumb for CSRF")
			if err := c.fetchCrumb(ctx); err == nil {
				logger.Info("Crumb fetched, retrying request")
				// Retry the request with crumb
				return c.doRequest(ctx, method, path, body)
			}
		}
		return nil, fmt.Errorf("access forbidden: check permissions")
	}

	return resp, nil
}

// fetchCrumb fetches the CSRF crumb from Jenkins
func (c *Client) fetchCrumb(ctx context.Context) error {
	c.crumbMu.Lock()
	defer c.crumbMu.Unlock()

	c.crumbTested = true

	logger.Debug("Fetching CSRF crumb")

	resp, err := c.doRequest(ctx, http.MethodGet, "/crumbIssuer/api/json", nil)
	if err != nil {
		logger.Warn("Failed to fetch crumb", "error", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.Warn("Crumb issuer returned non-200", "status", resp.StatusCode)
		return fmt.Errorf("crumb issuer returned status %d", resp.StatusCode)
	}

	var crumbResp struct {
		Crumb             string `json:"crumb"`
		CrumbRequestField string `json:"crumbRequestField"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&crumbResp); err != nil {
		logger.Error("Error decoding crumb response", "error", err)
		return fmt.Errorf("error decoding crumb response: %w", err)
	}

	c.crumb = crumbResp.Crumb
	c.crumbField = crumbResp.CrumbRequestField

	logger.Info("Crumb obtained successfully", "field", c.crumbField)

	return nil
}

// getJSON performs a GET request and decodes the JSON response
func (c *Client) getJSON(ctx context.Context, path string, v interface{}) error {
	resp, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		logger.Error("Unexpected status code",
			"status", resp.StatusCode,
			"body", string(body),
		)
		return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
		logger.Error("Error decoding JSON response", "error", err)
		return fmt.Errorf("error decoding response: %w", err)
	}

	return nil
}

// getText performs a GET request and returns the text response
func (c *Client) getText(ctx context.Context, path string, maxBytes int) (string, error) {
	resp, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	// Limit the amount of data we read
	limited := io.LimitReader(resp.Body, int64(maxBytes))
	body, err := io.ReadAll(limited)
	if err != nil {
		return "", fmt.Errorf("error reading response: %w", err)
	}

	logger.Debug("Text response received", "bytes", len(body))

	return string(body), nil
}

// buildTreeParam builds the tree parameter for API requests
func buildTreeParam(fields ...string) string {
	return "tree=" + url.QueryEscape(strings.Join(fields, ","))
}

// encodeJobPath encodes a job path for URL use (handles folders)
func encodeJobPath(jobName string) string {
	// Split by / and encode each part
	parts := strings.Split(jobName, "/")
	encoded := make([]string, len(parts))
	for i, part := range parts {
		encoded[i] = url.PathEscape(part)
	}
	return strings.Join(encoded, "/job/")
}
