package rules

import "fmt"

type Rule struct {
	ID       string
	Version  int
	Question string
}

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
	return NewRegistry(Rule{
		ID:       "person_name",
		Version:  1,
		Question: "¿Este texto parece representar el nombre de una persona?",
	})
}

func (r *Registry) Get(id string) (Rule, error) {
	rule, ok := r.rules[id]
	if !ok {
		return Rule{}, fmt.Errorf("rule %q is not registered", id)
	}
	return rule, nil
}
