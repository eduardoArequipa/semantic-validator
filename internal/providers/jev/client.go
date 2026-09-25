package jev

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/eduardoArequipa/semantic-validator/internal/metrics"
	"github.com/eduardoArequipa/semantic-validator/internal/providers"
)

var ErrInvalidResponse = errors.New("invalid Jev response")

type Client struct {
	httpClient *http.Client
	baseURL    string
	apiKey     string
	model      string
	metrics    *metrics.Metrics
}

func NewClient(httpClient *http.Client, baseURL, apiKey, model string) *Client {
	return NewClientWithMetrics(httpClient, baseURL, apiKey, model, nil)
}

func NewClientWithMetrics(httpClient *http.Client, baseURL, apiKey, model string, observability *metrics.Metrics) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &Client{
		httpClient: httpClient,
		baseURL:    strings.TrimRight(baseURL, "/"),
		apiKey:     apiKey,
		model:      model,
		metrics:    observability,
	}
}

func (c *Client) Evaluate(ctx context.Context, state, question string) (providers.Result, error) {
	payload := map[string]any{
		"model": c.model,
		"state": state,
		"questions": map[string]any{
			"result": map[string]any{
				"type":         "choice",
				"instructions": question,
				"criteria": map[string]string{
					"true":  "El texto cumple la condición.",
					"false": "El texto no cumple la condición.",
				},
			},
		},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return providers.Result{}, fmt.Errorf("marshal Jev request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/systemone", bytes.NewReader(body))
	if err != nil {
		return providers.Result{}, fmt.Errorf("create Jev request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	response, err := c.httpClient.Do(req)
	if c.metrics != nil {
		c.metrics.ObserveJevRequest(err)
	}
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return providers.Result{}, fmt.Errorf("%w: %v", providers.ErrTimeout, err)
		}
		return providers.Result{}, fmt.Errorf("call Jev: %w", err)
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(io.LimitReader(response.Body, 2<<20))
	if err != nil {
		return providers.Result{}, fmt.Errorf("read Jev response: %w", err)
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return providers.Result{}, fmt.Errorf("Jev returned HTTP %d", response.StatusCode)
	}

	result, err := parseResponse(responseBody)
	if err != nil {
		return providers.Result{}, err
	}
	return result, nil
}

func parseResponse(body []byte) (providers.Result, error) {
	var document any
	if err := json.Unmarshal(body, &document); err != nil {
		return providers.Result{}, fmt.Errorf("%w: malformed JSON", ErrInvalidResponse)
	}

	node := responseNode(document)
	value, ok := booleanValue(node)
	if !ok {
		return providers.Result{}, fmt.Errorf("%w: boolean answer not found", ErrInvalidResponse)
	}
	confidence, ok := confidenceValue(node, value)
	if !ok || confidence < 0 || confidence > 1 {
		return providers.Result{}, fmt.Errorf("%w: confidence not found or out of range", ErrInvalidResponse)
	}
	return providers.Result{Value: value, Confidence: confidence}, nil
}

func responseNode(document any) any {
	if object, ok := document.(map[string]any); ok {
		if answers, ok := object["answers"].(map[string]any); ok {
			if result, exists := answers["result"]; exists {
				return result
			}
			for _, answer := range answers {
				return answer
			}
		}
		for _, key := range []string{"result", "answer", "decision"} {
			if value, exists := object[key]; exists {
				return value
			}
		}
	}
	return document
}

func booleanValue(node any) (bool, bool) {
	switch value := node.(type) {
	case bool:
		return value, true
	case string:
		switch strings.ToLower(value) {
		case "true", "yes", "valid":
			return true, true
		case "false", "no", "invalid":
			return false, true
		}
	case map[string]any:
		for _, key := range []string{"value", "answer", "choice", "decision", "result"} {
			if candidate, exists := value[key]; exists {
				if parsed, ok := booleanValue(candidate); ok {
					return parsed, true
				}
			}
		}
	}
	return false, false
}

func confidenceValue(node any, answer bool) (float64, bool) {
	if object, ok := node.(map[string]any); ok {
		for _, key := range []string{"confidence", "probability", "score"} {
			if number, ok := object[key].(float64); ok {
				return number, true
			}
		}
		if probabilities, ok := object["probabilities"].(map[string]any); ok {
			key := "false"
			if answer {
				key = "true"
			}
			if number, ok := probabilities[key].(float64); ok {
				return number, true
			}
		}
	}
	return 0, false
}
