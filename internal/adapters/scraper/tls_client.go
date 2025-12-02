package scraper

import (
	"fmt"
	"io"
	"time"

	http "github.com/bogdanfinn/fhttp"
	tls_client "github.com/bogdanfinn/tls-client"
	"github.com/bogdanfinn/tls-client/profiles"
)

// TLSClient wraps the tls-client library for anti-detection HTTP requests.
type TLSClient struct {
	client tls_client.HttpClient
}

// NewTLSClient creates a new TLS client with Chrome fingerprint.
func NewTLSClient() (*TLSClient, error) {
	jar := tls_client.NewCookieJar()
	options := []tls_client.HttpClientOption{
		tls_client.WithTimeoutSeconds(30),
		tls_client.WithClientProfile(profiles.Chrome_120),
		tls_client.WithRandomTLSExtensionOrder(),
		tls_client.WithCookieJar(jar),
		// Enable automatic redirect following with cookie jar support
		// tls_client.WithNotFollowRedirects(), // Removed: was causing redirect loop
	}

	client, err := tls_client.NewHttpClient(tls_client.NewNoopLogger(), options...)
	if err != nil {
		return nil, fmt.Errorf("failed to create TLS client: %w", err)
	}

	tlsClient := &TLSClient{client: client}

	// Warm up: visit the main Digikala site to establish cookies before API calls
	if err := tlsClient.warmup(); err != nil {
		fmt.Printf("Warning: warmup request failed: %v\n", err)
		// Continue anyway - warmup is best-effort
	}

	return tlsClient, nil
}

// warmup makes a request to the main Digikala site to establish cookies.
func (c *TLSClient) warmup() error {
	req, err := http.NewRequest(http.MethodGet, "https://www.digikala.com/", nil)
	if err != nil {
		return err
	}

	req.Header = http.Header{
		"User-Agent":      {"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"},
		"Accept":          {"text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8"},
		"Accept-Language": {"en-US,en;q=0.9,fa;q=0.8"},
		"Accept-Encoding": {"gzip, deflate, br"},
		"Cache-Control":   {"max-age=0"},
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Just drain and discard the body
	_, _ = io.ReadAll(resp.Body)

	// Small delay after warmup
	time.Sleep(500 * time.Millisecond)

	return nil
}

// Get performs a GET request with anti-detection headers.
// Redirects are followed automatically by the underlying client with cookie persistence.
func (c *TLSClient) Get(url string) (string, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Add realistic headers
	req.Header = http.Header{
		"User-Agent":                {"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"},
		"Accept":                    {"application/json, text/plain, */*"},
		"Accept-Language":           {"en-US,en;q=0.9,fa;q=0.8"},
		"Accept-Encoding":           {"gzip, deflate, br"},
		"Cache-Control":             {"max-age=0"},
		"Sec-Ch-Ua":                 {`"Not_A Brand";v="8", "Chromium";v="120", "Google Chrome";v="120"`},
		"Sec-Ch-Ua-Mobile":          {"?0"},
		"Sec-Ch-Ua-Platform":        {`"Windows"`},
		"Sec-Fetch-Dest":            {"empty"},
		"Sec-Fetch-Mode":            {"cors"},
		"Sec-Fetch-Site":            {"none"},
		"Sec-Fetch-User":            {"?1"},
		"Upgrade-Insecure-Requests": {"1"},
		http.HeaderOrderKey: {
			"accept",
			"accept-language",
			"accept-encoding",
			"cache-control",
			"sec-ch-ua",
			"sec-ch-ua-mobile",
			"sec-ch-ua-platform",
			"sec-fetch-dest",
			"sec-fetch-mode",
			"sec-fetch-site",
			"sec-fetch-user",
			"upgrade-insecure-requests",
			"user-agent",
		},
	}

	// Add small delay to simulate human behavior
	time.Sleep(time.Duration(1+time.Now().UnixNano()%3) * time.Second)

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	return string(body), nil
}
