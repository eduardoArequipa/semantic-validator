package validation

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"github.com/eduardoArequipa/semantic-validator/internal/access"
	"strings"
	"sync"

	"github.com/eduardoArequipa/semantic-validator/internal/providers"
	"github.com/eduardoArequipa/semantic-validator/internal/rules"
)

const UncertainThreshold = 0.80

var ErrInvalidValue = errors.New("value must contain between 1 and 10000 characters")
var ErrUnknownRule = errors.New("unknown rule")
var ErrInvalidQuestion = errors.New("question must contain between 1 and 10000 characters")

type Result struct {
	Valid      *bool   `json:"valid"`
	Confidence float64 `json:"confidence"`
	Status     string  `json:"status"`
}

type BatchItem struct {
	ID    string
	Rule  string
	Value string
}

type BatchError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type BatchResult struct {
	ID         string      `json:"id"`
	Valid      *bool       `json:"valid"`
	Confidence float64     `json:"confidence"`
	Status     string      `json:"status"`
	Error      *BatchError `json:"error,omitempty"`
}

type ResultCache interface {
	Get(key string) (Result, bool)
	Set(key string, result Result)
}

type Service struct {
	registry *rules.Registry
	provider providers.Provider
	cache    ResultCache
}

func NewService(registry *rules.Registry, provider providers.Provider) *Service {
	return NewServiceWithCache(registry, provider, nil)
}

func NewServiceWithCache(registry *rules.Registry, provider providers.Provider, resultCache ResultCache) *Service {
	return &Service{registry: registry, provider: provider, cache: resultCache}
}

func (s *Service) Validate(ctx context.Context, ruleID, value string) (Result, error) {
	value = strings.TrimSpace(value)
	if !validText(value) {
		return Result{}, ErrInvalidValue
	}
	rule, err := s.registry.Get(ruleID)
	if err != nil {
		return Result{}, fmt.Errorf("%w: %s", ErrUnknownRule, ruleID)
	}
	key := cacheKey("validate", rule.ID, rule.Version, value)
	if result, found := s.getCached(key); found {
		return result, nil
	}
	return s.evaluateAndCache(ctx, key, value, rule.Question)
}

func (s *Service) Check(ctx context.Context, value, question string) (Result, error) {
	value = strings.TrimSpace(value)
	question = strings.TrimSpace(question)
	if !validText(value) {
		return Result{}, ErrInvalidValue
	}
	if !validText(question) {
		return Result{}, ErrInvalidQuestion
	}
	key := cacheKey("check", question, 1, value)
	if result, found := s.getCached(key); found {
		return result, nil
	}
	return s.evaluateAndCache(ctx, key, value, question)
}

func (s *Service) ValidateBatch(ctx context.Context, items []BatchItem) []BatchResult {
	results := make([]BatchResult, len(items))
	var waitGroup sync.WaitGroup
	waitGroup.Add(len(items))
	for index, item := range items {
		go func(index int, item BatchItem) {
			defer waitGroup.Done()
			result, err := s.Validate(ctx, item.Rule, item.Value)
			batchResult := BatchResult{ID: item.ID}
			if err != nil {
				batchResult.Status = "error"
				batchResult.Error = batchError(err)
			} else {
				batchResult.Valid = result.Valid
				batchResult.Confidence = result.Confidence
				batchResult.Status = result.Status
			}
			results[index] = batchResult
		}(index, item)
	}
	waitGroup.Wait()
	return results
}

func (s *Service) evaluateAndCache(ctx context.Context, key, value, question string) (Result, error) {
	if err := access.CountProviderAttempt(ctx); err != nil {
		return Result{}, err
	}
	answer, err := s.provider.Evaluate(ctx, value, question)
	if err != nil {
		return Result{}, err
	}
	if answer.Confidence < 0 || answer.Confidence > 1 {
		return Result{}, fmt.Errorf("provider returned invalid confidence")
	}
	result := Result{Confidence: answer.Confidence}
	if answer.Confidence < UncertainThreshold {
		result.Status = "uncertain"
		s.setCached(key, result)
		return result, nil
	}
	result.Valid = &answer.Value
	if answer.Value {
		result.Status = "valid"
	} else {
		result.Status = "invalid"
	}
	s.setCached(key, result)
	return result, nil
}

func validText(value string) bool {
	value = strings.TrimSpace(value)
	return value != "" && len([]rune(value)) <= 10000
}

func (s *Service) getCached(key string) (Result, bool) {
	if s.cache == nil || key == "" {
		return Result{}, false
	}
	return s.cache.Get(key)
}

func (s *Service) setCached(key string, result Result) {
	if s.cache != nil && key != "" {
		s.cache.Set(key, result)
	}
}

func cacheKey(operation, subject string, version int, value string) string {
	data := fmt.Sprintf("%s\x00%s\x00%d\x00%s", operation, subject, version, value)
	return fmt.Sprintf("%x", sha256.Sum256([]byte(data)))
}

func batchError(err error) *BatchError {
	switch {
	case errors.Is(err, access.ErrAccounting):
		return &BatchError{Code: "usage_unavailable", Message: "usage storage unavailable"}
	case errors.Is(err, ErrInvalidValue):
		return &BatchError{Code: "invalid_request", Message: err.Error()}
	case errors.Is(err, ErrUnknownRule):
		return &BatchError{Code: "unknown_rule", Message: "requested rule is not registered"}
	case errors.Is(err, providers.ErrTimeout):
		return &BatchError{Code: "provider_timeout", Message: "semantic provider request timed out"}
	default:
		return &BatchError{Code: "provider_error", Message: "semantic provider request failed"}
	}
}
