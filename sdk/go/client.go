// Package validator provides a Go client for the Semantic Validator API.
package validator

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	defaultBaseURL = "http://localhost:8080"
	maxResponseSize = 4 << 20
)

type Status string

const (
	StatusValid     Status = "valid"
	StatusInvalid   Status = "invalid"
	StatusUncertain Status = "uncertain"
	StatusError     Status = "error"
)

type Result struct {
	Valid      *bool   `json:"valid"`
	Confidence float64 `json:"confidence"`
	Status     Status  `json:"status"`
}

type BatchItem struct {
	ID    string `json:"id"`
	Rule  string `json:"rule"`
	Value string `json:"value"`
}

type BatchItemError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type BatchResult struct {
	ID         string          `json:"id"`
	Valid      *bool           `json:"valid"`
	Confidence float64         `json:"confidence"`
	Status     Status          `json:"status"`
	Error      *BatchItemError `json:"error,omitempty"`
}

type Options struct {
	BaseURL    string
	Timeout    time.Duration
	HTTPClient *http.Client
}

type Client struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

// APIError represents an error response returned by Semantic Validator.
type APIError struct {
	Code       string
	Message    string
	StatusCode int
}

func (e *APIError) Error() string {
	if e.StatusCode == 0 {
		return fmt.Sprintf("semantic validator: %s: %s", e.Code, e.Message)
	}
	return fmt.Sprintf("semantic validator: %s (HTTP %d): %s", e.Code, e.StatusCode, e.Message)
}

// NewClient creates a client using localhost:8080 and a 30-second timeout.
func NewClient(apiKey string) *Client {
	return &Client{
		apiKey:     strings.TrimSpace(apiKey),
		baseURL:    defaultBaseURL,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// NewClientWithOptions creates a client with a custom endpoint and/or HTTP client.
func NewClientWithOptions(apiKey string, options Options) (*Client, error) {
	if strings.TrimSpace(apiKey) == "" {
		return nil, errors.New("semantic validator API key is required")
	}
	baseURL := strings.TrimRight(strings.TrimSpace(options.BaseURL), "/")
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	httpClient := options.HTTPClient
	if httpClient == nil {
		timeout := options.Timeout
		if timeout <= 0 {
			timeout = 30 * time.Second
		}
		httpClient = &http.Client{Timeout: timeout}
	}
	return &Client{apiKey: apiKey, baseURL: baseURL, httpClient: httpClient}, nil
}

func (c *Client) Validate(ctx context.Context, rule, value string) (Result, error) {
	var result Result
	err := c.post(ctx, "/v1/validate", struct {
		Rule  string `json:"rule"`
		Value string `json:"value"`
	}{Rule: rule, Value: value}, &result)
	if err != nil {
		return Result{}, err
	}
	if err := validateResult(result); err != nil {
		return Result{}, err
	}
	return result, nil
}

func (c *Client) Name(ctx context.Context, value string) (Result, error) {
	return c.Validate(ctx, "person_name", value)
}

func (c *Client) Check(ctx context.Context, value, question string) (Result, error) {
	var result Result
	err := c.post(ctx, "/v1/check", struct {
		Value    string `json:"value"`
		Question string `json:"question"`
	}{Value: value, Question: question}, &result)
	if err != nil {
		return Result{}, err
	}
	if err := validateResult(result); err != nil {
		return Result{}, err
	}
	return result, nil
}

func (c *Client) ValidateBatch(ctx context.Context, items []BatchItem) ([]BatchResult, error) {
	if len(items) == 0 || len(items) > 100 {
		return nil, errors.New("batch items must contain between 1 and 100 elements")
	}
	ids := make(map[string]struct{}, len(items))
	for _, item := range items {
		if strings.TrimSpace(item.ID) == "" {
			return nil, errors.New("each batch item requires an id")
		}
		if _, exists := ids[item.ID]; exists {
			return nil, fmt.Errorf("batch item ids must be unique: %q", item.ID)
		}
		ids[item.ID] = struct{}{}
	}
	var response struct {
		Items []BatchResult `json:"items"`
	}
	if err := c.post(ctx, "/v1/validate/batch", struct {
		Items []BatchItem `json:"items"`
	}{Items: items}, &response); err != nil {
		return nil, err
	}
	if len(response.Items) != len(items) {
		return nil, errors.New("semantic validator returned an invalid batch response")
	}
	for i, result := range response.Items {
		if result.ID != items[i].ID || !validStatus(result.Status, true) || result.Confidence < 0 || result.Confidence > 1 {
			return nil, errors.New("semantic validator returned an invalid batch result")
		}
		if result.Status == StatusError && result.Error == nil {
			return nil, errors.New("semantic validator returned a batch error without details")
		}
	}
	return response.Items, nil
}

func (c *Client) post(ctx context.Context, path string, payload, destination any) error {
	if strings.TrimSpace(c.apiKey) == "" {
		return errors.New("semantic validator API key is required")
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("encode request: %w", err)
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Authorization", "Bearer "+c.apiKey)
	request.Header.Set("Content-Type", "application/json")

	response, err := c.httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("semantic validator request failed: %w", err)
	}
	defer response.Body.Close()
	responseBody, err := io.ReadAll(io.LimitReader(response.Body, maxResponseSize+1))
	if err != nil {
		return fmt.Errorf("read semantic validator response: %w", err)
	}
	if len(responseBody) > maxResponseSize {
		return errors.New("semantic validator response exceeded the size limit")
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return parseAPIError(response.StatusCode, responseBody)
	}
	if err := json.Unmarshal(responseBody, destination); err != nil {
		return fmt.Errorf("decode semantic validator response: %w", err)
	}
	return nil
}

func parseAPIError(statusCode int, body []byte) error {
	var envelope struct {
		Error BatchItemError `json:"error"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil || envelope.Error.Code == "" {
		return &APIError{Code: "http_error", Message: http.StatusText(statusCode), StatusCode: statusCode}
	}
	return &APIError{Code: envelope.Error.Code, Message: envelope.Error.Message, StatusCode: statusCode}
}

func validateResult(result Result) error {
	if !validStatus(result.Status, false) || result.Confidence < 0 || result.Confidence > 1 {
		return errors.New("semantic validator returned an invalid validation result")
	}
	return nil
}

func validStatus(status Status, allowError bool) bool {
	return status == StatusValid || status == StatusInvalid || status == StatusUncertain || (allowError && status == StatusError)
}
