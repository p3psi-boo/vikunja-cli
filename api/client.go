package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/p3psi-boo/vikunja-cli/config"
)

type APIError struct {
	StatusCode int
	Message    string
	Body       string
}

func (e *APIError) Error() string {
	if strings.TrimSpace(e.Message) != "" {
		return fmt.Sprintf("api error (%d): %s", e.StatusCode, e.Message)
	}

	if strings.TrimSpace(e.Body) != "" {
		return fmt.Sprintf("api error (%d): %s", e.StatusCode, e.Body)
	}

	return fmt.Sprintf("api error (%d)", e.StatusCode)
}

type Client struct {
	baseURL     string
	httpClient  *http.Client
	staticToken string
	token       string
}

type requestOptions struct {
	withAuth   bool
	retryOn401 bool
}

func NewClient(cfg *config.Config) (*Client, error) {
	client := &Client{
		baseURL:     strings.TrimRight(cfg.Server.APIURL, "/"),
		httpClient:  &http.Client{Timeout: 30 * time.Second},
		staticToken: strings.TrimSpace(cfg.Server.APIToken),
	}

	if client.staticToken == "" {
		token, err := LoadToken()
		if err == nil {
			client.token = token
		}
	}

	return client, nil
}

func TotalPagesFromHeader(header http.Header) int {
	raw := strings.TrimSpace(header.Get("x-pagination-total-pages"))
	if raw == "" {
		return 1
	}

	pages, err := strconv.Atoi(raw)
	if err != nil || pages < 1 {
		return 1
	}

	return pages
}

func (c *Client) GetJSON(ctx context.Context, path string, out any) error {
	return c.doJSON(ctx, http.MethodGet, path, nil, out, requestOptions{withAuth: true, retryOn401: true})
}

func (c *Client) PostJSON(ctx context.Context, path string, in any, out any) error {
	return c.doJSON(ctx, http.MethodPost, path, in, out, requestOptions{withAuth: true, retryOn401: true})
}

func (c *Client) PutJSON(ctx context.Context, path string, in any, out any) error {
	return c.doJSON(ctx, http.MethodPut, path, in, out, requestOptions{withAuth: true, retryOn401: true})
}

func (c *Client) DeleteJSON(ctx context.Context, path string, in any, out any) error {
	return c.doJSON(ctx, http.MethodDelete, path, in, out, requestOptions{withAuth: true, retryOn401: true})
}

func (c *Client) doJSON(ctx context.Context, method, path string, in any, out any, opts requestOptions) error {
	body, err := marshalBody(in)
	if err != nil {
		return err
	}

	retry := opts.retryOn401
	for {
		resp, err := c.doRequest(ctx, method, path, body, opts.withAuth)
		if err != nil {
			return err
		}

		respBody, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return fmt.Errorf("read response body: %w", err)
		}

		if resp.StatusCode == http.StatusUnauthorized && retry {
			retry = false
			if err := c.RefreshToken(ctx); err != nil {
				return fmt.Errorf("authentication failed; run `vja login`: %w", err)
			}
			continue
		}

		if resp.StatusCode < 200 || resp.StatusCode > 299 {
			return parseAPIError(resp.StatusCode, respBody)
		}

		if out == nil || len(bytes.TrimSpace(respBody)) == 0 {
			return nil
		}

		if err := json.Unmarshal(respBody, out); err != nil {
			return fmt.Errorf("decode response json: %w", err)
		}

		return nil
	}
}

func (c *Client) doRequest(ctx context.Context, method, path string, body []byte, withAuth bool) (*http.Response, error) {
	fullURL, err := resolveURL(c.baseURL, path)
	if err != nil {
		return nil, err
	}

	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, fullURL, reader)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	if withAuth {
		token := c.authToken()
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request %s %s failed: %w", method, fullURL, err)
	}

	return resp, nil
}

func (c *Client) authToken() string {
	if c.staticToken != "" {
		return c.staticToken
	}

	return c.token
}

func (c *Client) setToken(token string) {
	if c.staticToken != "" {
		return
	}

	c.token = strings.TrimSpace(token)
}

func marshalBody(in any) ([]byte, error) {
	if in == nil {
		return nil, nil
	}

	body, err := json.Marshal(in)
	if err != nil {
		return nil, fmt.Errorf("encode request json: %w", err)
	}

	return body, nil
}

func parseAPIError(statusCode int, body []byte) error {
	raw := strings.TrimSpace(string(body))
	if raw == "" {
		return &APIError{StatusCode: statusCode}
	}

	type messageEnvelope struct {
		Message string `json:"message"`
	}

	var envelope messageEnvelope
	if err := json.Unmarshal(body, &envelope); err == nil && strings.TrimSpace(envelope.Message) != "" {
		return &APIError{StatusCode: statusCode, Message: envelope.Message, Body: raw}
	}

	return &APIError{StatusCode: statusCode, Body: raw}
}

func resolveURL(baseURL, path string) (string, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return "", fmt.Errorf("request path must not be empty")
	}

	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		if _, err := url.ParseRequestURI(path); err != nil {
			return "", fmt.Errorf("invalid absolute url %q: %w", path, err)
		}
		return path, nil
	}

	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	fullURL := baseURL + path
	if _, err := url.ParseRequestURI(fullURL); err != nil {
		return "", fmt.Errorf("invalid request url %q: %w", fullURL, err)
	}

	return fullURL, nil
}
