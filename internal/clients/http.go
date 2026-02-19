package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"go-unraid-clean/internal/logging"

	"github.com/rs/zerolog"
)

type HTTPClient struct {
	BaseURL *url.URL
	APIKey  string
	Client  *http.Client
}

func NewHTTPClient(baseURL, apiKey string) (*HTTPClient, error) {
	parsed, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("parse base_url: %w", err)
	}
	return &HTTPClient{
		BaseURL: parsed,
		APIKey:  apiKey,
		Client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

func (c *HTTPClient) Resolve(path string) string {
	u := *c.BaseURL
	trimmed := strings.TrimLeft(path, "/")
	rawQuery := ""
	if idx := strings.Index(trimmed, "?"); idx >= 0 {
		rawQuery = trimmed[idx+1:]
		trimmed = trimmed[:idx]
	}
	u.Path = strings.TrimRight(u.Path, "/") + "/" + trimmed
	if rawQuery != "" {
		if u.RawQuery != "" {
			u.RawQuery = u.RawQuery + "&" + rawQuery
		} else {
			u.RawQuery = rawQuery
		}
	}
	return u.String()
}

func decodeJSONBody(resp *http.Response, out any) error {
	defer resp.Body.Close()
	dec := json.NewDecoder(resp.Body)
	dec.UseNumber()
	if err := dec.Decode(out); err != nil {
		return fmt.Errorf("decode json: %w", err)
	}
	return nil
}

func readBody(resp *http.Response) ([]byte, error) {
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func doRequest(ctx context.Context, client *http.Client, req *http.Request) (*http.Response, error) {
	if ctx != nil {
		req = req.WithContext(ctx)
	}
	log := logging.L()
	if log.GetLevel() <= zerolog.DebugLevel {
		log.Debug().Str("method", req.Method).Str("url", req.URL.String()).Msg("HTTP request")
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if log.GetLevel() <= zerolog.DebugLevel {
		log.Debug().Int("status", resp.StatusCode).Str("url", req.URL.String()).Msg("HTTP response")
	}
	if log.GetLevel() <= zerolog.TraceLevel {
		body, err := readBody(resp)
		if err != nil {
			return nil, err
		}
		log.Trace().
			Int("status", resp.StatusCode).
			Str("url", req.URL.String()).
			Str("body", truncateBody(string(body))).
			Msg("HTTP response body")
		resp.Body = io.NopCloser(bytes.NewReader(body))
	}
	return resp, nil
}

func truncateBody(body string) string {
	const limit = 2000
	if len(body) <= limit {
		return body
	}
	return body[:limit] + "...(truncated)"
}
