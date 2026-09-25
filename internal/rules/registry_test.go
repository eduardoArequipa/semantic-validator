package rules

import "testing"

func TestDefaultRegistry(t *testing.T) {
	registry := DefaultRegistry()
	for _, id := range []string{"person_name", "product_description", "address"} {
		rule, err := registry.Get(id)
		if err != nil || rule.ID != id || rule.Version != 1 || rule.Question == "" || rule.Description == "" {
			t.Fatalf("invalid built-in rule %q: %+v %v", id, rule, err)
		}
	}
	if _, err := registry.Get("missing"); err == nil {
		t.Fatal("unknown rule accepted")
	}
}
