package validator

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
	"unicode/utf8"
)

const (
	defaultJevURL = "https://api.typesafe.ai"
)

// Keep these questions in sync with internal/rules/catalog.json.
var directRuleQuestions = map[string]string{
	"person_name":         "¿Este texto parece representar el nombre de una persona?",
	"product_description": "¿El texto identifica un producto y al menos una característica concreta, más allá de una opinión genérica?",
	"address":             "¿El texto parece una dirección física suficientemente específica para ubicar un lugar, y no solo el nombre de una ciudad o país?",
}

type DirectOptions struct {
	BaseURL    string
	Timeout    time.Duration
	HTTPClient *http.Client
}

// DirectClient contacts Jev with the caller's own key; no Semantic Validator server is used.
type DirectClient struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

// DirectError describes a local or Jev failure without disclosing the key or input text.
type DirectError struct {
	Code       string
	StatusCode int
}

func (e *DirectError) Error() string {
	if e.StatusCode != 0 {
		return fmt.Sprintf("Jev %s (HTTP %d)", e.Code, e.StatusCode)
	}
	return "Jev " + e.Code
}

func NewDirectClient(jevAPIKey string) (*DirectClient, error) {
	return NewDirectClientWithOptions(jevAPIKey, DirectOptions{})
}

func NewDirectClientWithOptions(jevAPIKey string, options DirectOptions) (*DirectClient, error) {
	if strings.TrimSpace(jevAPIKey) == "" {
		return nil, errors.New("Jev API key is required")
	}
	base := strings.TrimRight(strings.TrimSpace(options.BaseURL), "/")
	if base == "" {
		base = defaultJevURL
	}
	u, err := url.Parse(base)
	if err != nil || u.Hostname() == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" || (u.Path != "" && u.Path != "/") ||
		!((u.Scheme == "https" && u.Hostname() == "api.typesafe.ai") || (u.Scheme == "http" && (u.Hostname() == "localhost" || u.Hostname() == "127.0.0.1" || u.Hostname() == "::1"))) {
		return nil, errors.New("Jev endpoint must be api.typesafe.ai (except loopback tests)")
	}
	httpClient := options.HTTPClient
	if httpClient == nil {
		timeout := options.Timeout
		if timeout <= 0 {
			timeout = 30 * time.Second
		}
		httpClient = &http.Client{Timeout: timeout}
	}
	copyClient := *httpClient
	copyClient.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	return &DirectClient{apiKey: strings.TrimSpace(jevAPIKey), baseURL: base, httpClient: &copyClient}, nil
}

func (c *DirectClient) Validate(ctx context.Context, rule, value string) (Result, error) {
	if _, err := directText(value); err != nil {
		return Result{}, err
	}
	question, ok := directRuleQuestions[rule]
	if !ok {
		return Result{}, &DirectError{Code: "unknown_rule"}
	}
	return c.Check(ctx, value, question)
}

func (c *DirectClient) Name(ctx context.Context, value string) (Result, error) {
	return c.Validate(ctx, "person_name", value)
}

func (c *DirectClient) Check(ctx context.Context, value, question string) (Result, error) {
	state, err := directText(value)
	if err != nil {
		return Result{}, err
	}
	instructions, err := directText(question)
	if err != nil {
		return Result{}, err
	}
	payload := map[string]any{
		"model": "jev-latest", "state": state,
		"questions": map[string]any{"result": map[string]any{
			"type": "choice", "instructions": instructions,
			"criteria": map[string]string{"true": "El texto cumple la condición.", "false": "El texto no cumple la condición."},
		}},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return Result{}, &DirectError{Code: "invalid_request"}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/systemone", bytes.NewReader(body))
	if err != nil {
		return Result{}, &DirectError{Code: "invalid_request"}
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	response, err := c.httpClient.Do(req)
	if err != nil {
		var netErr net.Error
		if errors.Is(err, context.DeadlineExceeded) || (errors.As(err, &netErr) && netErr.Timeout()) {
			return Result{}, &DirectError{Code: "provider_timeout"}
		}
		return Result{}, &DirectError{Code: "provider_error"}
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return Result{}, &DirectError{Code: "provider_error", StatusCode: response.StatusCode}
	}
	raw, err := io.ReadAll(io.LimitReader(response.Body, 2<<20+1))
	if err != nil || len(raw) > 2<<20 {
		return Result{}, &DirectError{Code: "invalid_response"}
	}
	var document struct {
		Answers map[string]struct {
			Choice        string             `json:"choice"`
			Confidence    *float64           `json:"confidence"`
			Probabilities map[string]float64 `json:"probabilities"`
		} `json:"answers"`
	}
	if err := json.Unmarshal(raw, &document); err != nil {
		return Result{}, &DirectError{Code: "invalid_response"}
	}
	answer, ok := document.Answers["result"]
	if !ok || (answer.Choice != "true" && answer.Choice != "false") {
		return Result{}, &DirectError{Code: "invalid_response"}
	}
	confidence := answer.Confidence
	if confidence == nil {
		if value, ok := answer.Probabilities[answer.Choice]; ok {
			confidence = &value
		}
	}
	if confidence == nil || *confidence < 0 || *confidence > 1 {
		return Result{}, &DirectError{Code: "invalid_response"}
	}
	result := Result{Confidence: *confidence, Status: StatusUncertain}
	if *confidence < 0.8 {
		return result, nil
	}
	valid := answer.Choice == "true"
	result.Valid = &valid
	if valid {
		result.Status = StatusValid
	} else {
		result.Status = StatusInvalid
	}
	return result, nil
}

func (c *DirectClient) ValidateBatch(ctx context.Context, items []BatchItem) ([]BatchResult, error) {
	if len(items) < 1 || len(items) > 100 {
		return nil, errors.New("items must contain between 1 and 100 elements")
	}
	ids := make(map[string]bool, len(items))
	for _, item := range items {
		if strings.TrimSpace(item.ID) == "" || ids[item.ID] {
			return nil, errors.New("batch items require unique ids")
		}
		ids[item.ID] = true
	}
	results := make([]BatchResult, len(items))
	jobs := make(chan int)
	var workers sync.WaitGroup
	for i := 0; i < 4 && i < len(items); i++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for index := range jobs {
				item := items[index]
				answer, err := c.Validate(ctx, item.Rule, item.Value)
				if err == nil {
					results[index] = BatchResult{ID: item.ID, Valid: answer.Valid, Confidence: answer.Confidence, Status: answer.Status}
					continue
				}
				code := "provider_error"
				var directErr *DirectError
				if errors.As(err, &directErr) {
					code = directErr.Code
				}
				message := "semantic provider request failed"
				if code == "unknown_rule" {
					message = "requested rule is not registered"
				}
				if code == "invalid_request" {
					message = "invalid value"
				}
				results[index] = BatchResult{ID: item.ID, Status: StatusError, Error: &BatchItemError{Code: code, Message: message}}
			}
		}()
	}
	for index := range items {
		jobs <- index
	}
	close(jobs)
	workers.Wait()
	return results, nil
}

func directText(value string) (string, error) {
	text := strings.TrimSpace(value)
	if text == "" || utf8.RuneCountInString(text) > 10000 {
		return "", &DirectError{Code: "invalid_request"}
	}
	return text, nil
}
