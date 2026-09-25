package validation

import (
	"context"
	"testing"

	"github.com/eduardoArequipa/semantic-validator/internal/providers"
	"github.com/eduardoArequipa/semantic-validator/internal/rules"
)

func BenchmarkServiceValidateWithoutCache(b *testing.B) {
	service := NewService(rules.DefaultRegistry(), fakeProvider{result: providers.Result{Value: true, Confidence: .98}})
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := service.Validate(context.Background(), "person_name", "Jorge Eduardo"); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkServiceValidateCacheHit(b *testing.B) {
	cache := &testCache{entries: make(map[string]Result)}
	service := NewServiceWithCache(rules.DefaultRegistry(), fakeProvider{result: providers.Result{Value: true, Confidence: .98}}, cache)
	_, _ = service.Validate(context.Background(), "person_name", "Jorge Eduardo")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := service.Validate(context.Background(), "person_name", "Jorge Eduardo"); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkServiceValidateBatch(b *testing.B) {
	service := NewService(rules.DefaultRegistry(), fakeProvider{result: providers.Result{Value: true, Confidence: .98}})
	items := make([]BatchItem, 20)
	for i := range items {
		items[i] = BatchItem{ID: "item", Rule: "person_name", Value: "Jorge Eduardo"}
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		results := service.ValidateBatch(context.Background(), items)
		if len(results) != len(items) {
			b.Fatal("unexpected batch size")
		}
	}
}
