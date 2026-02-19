package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
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
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
