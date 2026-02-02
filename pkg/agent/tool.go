package agent

import (
	"context"
)

// Tool defines the interface for executable tools.
type Tool interface {
	// Name returns the unique name of the tool (e.g., "read_file").
	Name() string

	// Description returns a human-readable description of what the tool does.
	Description() string

	// Schema returns the JSON schema for the tool's arguments.
	// This is typically a struct or map that marshals to the expected JSON schema format.
	Schema() interface{}

	// Execute runs the tool with the provided arguments.
	Execute(ctx context.Context, args map[string]interface{}) (interface{}, error)
}
