package rules

import (
	_ "embed"
	"encoding/json"
	"fmt"
)

type Rule struct {
	ID          string `json:"id"`
	Version     int    `json:"version"`
	Question    string `json:"question"`
	Description string `json:"description"`
}

//go:embed catalog.json
var defaultCatalog []byte

type Registry struct {
	rules map[string]Rule
}

func NewRegistry(definitions ...Rule) *Registry {
	registered := make(map[string]Rule, len(definitions))
	for _, rule := range definitions {
		registered[rule.ID] = rule
	}
	return &Registry{rules: registered}
}

func DefaultRegistry() *Registry {
	var definitions []Rule
	if err := json.Unmarshal(defaultCatalog, &definitions); err != nil {
		panic(fmt.Sprintf("invalid built-in rule catalog: %v", err))
	}
	return NewRegistry(definitions...)
}

func (r *Registry) Get(id string) (Rule, error) {
	rule, ok := r.rules[id]
	if !ok {
		return Rule{}, fmt.Errorf("rule %q is not registered", id)
	}
	return rule, nil
}
