package tools

import (
	"sort"

	"github.com/techmuch/castor/pkg/agent"
)

// Registry manages a collection of tools.
type Registry struct {
	tools map[string]agent.Tool
}

// NewRegistry creates a new tool registry.
func NewRegistry() *Registry {
	return &Registry{
		tools: make(map[string]agent.Tool),
	}
}

// Register adds a tool to the registry.
func (r *Registry) Register(t agent.Tool) {
	r.tools[t.Name()] = t
}

// Get retrieves a tool by name.
func (r *Registry) Get(name string) (agent.Tool, bool) {
	t, ok := r.tools[name]
	return t, ok
}

// List returns all registered tools, sorted by name.
func (r *Registry) List() []agent.Tool {
	list := make([]agent.Tool, 0, len(r.tools))
	for _, t := range r.tools {
		list = append(list, t)
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].Name() < list[j].Name()
	})
	return list
}
