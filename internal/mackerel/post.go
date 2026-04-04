package mackerel

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// PostServiceMetrics はサービスメトリクスを投稿する。metrics が空なら何もしない。
func PostServiceMetrics(ctx context.Context, client *http.Client, baseURL, serviceName, apiKey string, metrics []ServiceMetric) error {
	if len(metrics) == 0 {
		return nil
	}
	if apiKey == "" {
		return fmt.Errorf("mackerel: api key is empty")
	}
	if serviceName == "" {
		return fmt.Errorf("mackerel: service name is empty")
	}

	endpoint, err := tsdbEndpoint(baseURL, serviceName)
	if err != nil {
		return err
	}

	body, err := json.Marshal(metrics)
	if err != nil {
		return fmt.Errorf("mackerel: encode json: %w", err)
	}

	const maxAttempts = 3
	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if attempt > 1 {
			backoff := time.Duration(attempt-1) * time.Second
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoff):
			}
		}
		lastErr = postOnce(ctx, client, endpoint, apiKey, body)
		if lastErr == nil {
			return nil
		}
		if !isRetryable(lastErr) {
			return lastErr
		}
	}
	return fmt.Errorf("mackerel: post failed after %d attempts: %w", maxAttempts, lastErr)
}

func tsdbEndpoint(baseURL, serviceName string) (string, error) {
	base := strings.TrimSuffix(strings.TrimSpace(baseURL), "/")
	if base == "" {
		base = DefaultBaseURL
	}
	if _, err := url.Parse(base); err != nil {
		return "", fmt.Errorf("mackerel: base url: %w", err)
	}
	return fmt.Sprintf("%s/api/v0/services/%s/tsdb", base, url.PathEscape(serviceName)), nil
}

func postOnce(ctx context.Context, client *http.Client, endpoint, apiKey string, body []byte) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Api-Key", apiKey)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 64<<10))

	if resp.StatusCode == http.StatusOK {
		return nil
	}
	return &httpStatusError{status: resp.StatusCode, statusLine: resp.Status, body: respBody}
}

type httpStatusError struct {
	status     int
	statusLine string
	body       []byte
}

func (e *httpStatusError) Error() string {
	msg := strings.TrimSpace(string(e.body))
	if msg == "" {
		return fmt.Sprintf("mackerel: %s", e.statusLine)
	}
	return fmt.Sprintf("mackerel: %s: %s", e.statusLine, msg)
}

func isRetryable(err error) bool {
	var hs *httpStatusError
	if errors.As(err, &hs) {
		return hs.status >= 500 || hs.status == http.StatusTooManyRequests
	}
	// ネットワークエラーなど HTTP 応答が無い場合はリトライする
	return true
}
