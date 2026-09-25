package validation

import (
	"context"
	"errors"
	"testing"

	"github.com/eduardoArequipa/semantic-validator/internal/providers"
	"github.com/eduardoArequipa/semantic-validator/internal/rules"
)

type fakeProvider struct {
	result providers.Result
	err    error
}

type countingProvider struct {
	calls  int
	result providers.Result
}

func (p *countingProvider) Evaluate(context.Context, string, string) (providers.Result, error) {
	p.calls++
	return p.result, nil
}

func (f fakeProvider) Evaluate(context.Context, string, string) (providers.Result, error) {
	return f.result, f.err
}

func TestServiceValidate(t *testing.T) {
	tests := []struct {
		name   string
		result providers.Result
		status string
		valid  *bool
	}{
		{name: "valid", result: providers.Result{Value: true, Confidence: .98}, status: "valid", valid: boolPtr(true)},
		{name: "invalid", result: providers.Result{Value: false, Confidence: .97}, status: "invalid", valid: boolPtr(false)},
		{name: "uncertain", result: providers.Result{Value: true, Confidence: .63}, status: "uncertain", valid: nil},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service := NewService(rules.DefaultRegistry(), fakeProvider{result: test.result})
			result, err := service.Validate(context.Background(), "person_name", "Jorge Eduardo")
			if err != nil || result.Status != test.status || !sameBool(result.Valid, test.valid) {
				t.Fatalf("result=%+v err=%v", result, err)
			}
		})
	}
}

func TestServiceRejectsInvalidValue(t *testing.T) {
	service := NewService(rules.DefaultRegistry(), fakeProvider{})
	if _, err := service.Validate(context.Background(), "person_name", " "); !errors.Is(err, ErrInvalidValue) {
		t.Fatalf("expected ErrInvalidValue, got %v", err)
	}
}

func TestServiceValidatesNewRules(t *testing.T) {
	service := NewService(rules.DefaultRegistry(), fakeProvider{result: providers.Result{Value: true, Confidence: .95}})
	for _, ruleID := range []string{"product_description", "address"} {
		result, err := service.Validate(context.Background(), ruleID, "Taladro de 18 V / Av. Bolívar 123")
		if err != nil || result.Valid == nil || !*result.Valid {
			t.Fatalf("rule=%s result=%+v err=%v", ruleID, result, err)
		}
	}
}

func TestServiceCheck(t *testing.T) {
	service := NewService(rules.DefaultRegistry(), fakeProvider{result: providers.Result{Value: true, Confidence: .97}})
	result, err := service.Check(context.Background(), "Necesito devolver el producto", "¿El cliente solicita una devolución?")
	if err != nil || result.Valid == nil || !*result.Valid || result.Status != "valid" {
		t.Fatalf("result=%+v err=%v", result, err)
	}
}

func TestServiceCheckRequiresQuestion(t *testing.T) {
	service := NewService(rules.DefaultRegistry(), fakeProvider{})
	if _, err := service.Check(context.Background(), "texto", " "); !errors.Is(err, ErrInvalidQuestion) {
		t.Fatalf("expected ErrInvalidQuestion, got %v", err)
	}
}

func TestServiceValidateBatchPreservesOrder(t *testing.T) {
	service := NewService(rules.DefaultRegistry(), fakeProvider{result: providers.Result{Value: true, Confidence: .98}})
	results := service.ValidateBatch(context.Background(), []BatchItem{
		{ID: "first", Rule: "person_name", Value: "Jorge Eduardo"},
		{ID: "second", Rule: "person_name", Value: "Ana"},
	})
	if len(results) != 2 || results[0].ID != "first" || results[1].ID != "second" {
		t.Fatalf("unexpected order: %+v", results)
	}
}

func TestServiceUsesCache(t *testing.T) {
	provider := &countingProvider{result: providers.Result{Value: true, Confidence: .98}}
	cache := &testCache{entries: make(map[string]Result)}
	service := NewServiceWithCache(rules.DefaultRegistry(), provider, cache)

	first, firstErr := service.Validate(context.Background(), "person_name", "Jorge Eduardo")
	second, secondErr := service.Validate(context.Background(), "person_name", "Jorge Eduardo")
	if firstErr != nil || secondErr != nil || provider.calls != 1 || first.Status != second.Status {
		t.Fatalf("first=%+v second=%+v calls=%d", first, second, provider.calls)
	}
}

type testCache struct {
	entries map[string]Result
}

func (c *testCache) Get(key string) (Result, bool) {
	result, found := c.entries[key]
	return result, found
}

func (c *testCache) Set(key string, result Result) {
	c.entries[key] = result
}

func boolPtr(value bool) *bool { return &value }

func sameBool(left, right *bool) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return *left == *right
}
