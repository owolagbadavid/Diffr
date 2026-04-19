package strategy

import (
	"diffr/internal/model"
	"fmt"
)

// Strategy defines how PR files are organized for review.
type Strategy interface {
	Name() string
	Description() string
	Organize(files []model.FileDiff) []model.FileGroup
}

var registry = map[string]Strategy{}

// Register adds a strategy to the global registry.
func Register(s Strategy) {
	registry[s.Name()] = s
}

// Get returns a strategy by name.
func Get(name string) (Strategy, error) {
	s, ok := registry[name]
	if !ok {
		return nil, fmt.Errorf("unknown strategy %q, available: %v", name, Names())
	}
	return s, nil
}

// Names returns all registered strategy names.
func Names() []string {
	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}
	return names
}
