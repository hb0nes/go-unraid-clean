package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

type TautulliClient struct {
	http *HTTPClient
}

func NewTautulliClient(baseURL, apiKey string) (*TautulliClient, error) {
	hc, err := NewHTTPClient(baseURL, apiKey)
	if err != nil {
		return nil, err
	}
	return &TautulliClient{http: hc}, nil
}

func (c *TautulliClient) History(ctx context.Context) ([]map[string]any, error) {
	all := []map[string]any{}
	start := 0
	length := 200
	for {
		page, total, err := c.historyPage(ctx, start, length)
		if err != nil {
			return nil, err
		}
		all = append(all, page...)
		if len(page) == 0 || start+len(page) >= total {
			break
		}
		start += len(page)
	}
	return all, nil
}

func (c *TautulliClient) historyPage(ctx context.Context, start, length int) ([]map[string]any, int, error) {
	base := c.http.Resolve("api/v2")
	u, err := url.Parse(base)
	if err != nil {
		return nil, 0, err
	}
	q := u.Query()
	q.Set("cmd", "get_history")
	q.Set("apikey", c.http.APIKey)
	q.Set("length", strconv.Itoa(length))
	q.Set("start", strconv.Itoa(start))
	q.Set("order", "desc")
	q.Set("sort", "date")
	u.RawQuery = q.Encode()

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, 0, err
	}

	resp, err := doRequest(ctx, c.http.Client, req)
	if err != nil {
		return nil, 0, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := readBody(resp)
		return nil, 0, fmt.Errorf("tautulli history: status %d: %s", resp.StatusCode, string(body))
	}

	var payload map[string]any
	if err := decodeJSONBody(resp, &payload); err != nil {
		return nil, 0, err
	}

	respMap, ok := payload["response"].(map[string]any)
	if !ok {
		return nil, 0, fmt.Errorf("tautulli history: missing response")
	}
	if result, ok := respMap["result"].(string); ok && result != "success" {
		return nil, 0, fmt.Errorf("tautulli history: result %s", result)
	}

	dataMap, ok := respMap["data"].(map[string]any)
	if !ok {
		return nil, 0, fmt.Errorf("tautulli history: missing data")
	}

	total := getIntFromMap(dataMap, "recordsFiltered")
	if total == 0 {
		total = getIntFromMap(dataMap, "recordsTotal")
	}
	itemsRaw, ok := dataMap["data"].([]any)
	if !ok {
		return nil, 0, fmt.Errorf("tautulli history: missing data list")
	}

	items := make([]map[string]any, 0, len(itemsRaw))
	for _, entry := range itemsRaw {
		if item, ok := entry.(map[string]any); ok {
			items = append(items, item)
		}
	}
	return items, total, nil
}

func getIntFromMap(m map[string]any, key string) int {
	val, ok := m[key]
	if !ok {
		return 0
	}
	return coerceInt(val)
}

func coerceInt(val any) int {
	switch v := val.(type) {
	case int:
		return v
	case int64:
		return int(v)
	case float64:
		return int(v)
	case json.Number:
		if i, err := strconv.Atoi(v.String()); err == nil {
			return i
		}
	}
	return 0
}
