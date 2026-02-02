package edit

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/techmuch/castor/pkg/agent"
)

// Ensure EditTool implements agent.Tool
var _ agent.Tool = (*EditTool)(nil)

// EditTool performs text replacements in files.
type EditTool struct {
	WorkspaceRoot string
}

func (t *EditTool) Name() string { return "replace" }

func (t *EditTool) Description() string {
	return "Replaces text within a file. Provide unique old_string to target the change."
}

func (t *EditTool) Schema() interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"path": map[string]interface{}{
				"type": "string",
			},
			"old_string": map[string]interface{}{
				"type": "string",
			},
			"new_string": map[string]interface{}{
				"type": "string",
			},
		},
		"required": []string{"path", "old_string", "new_string"},
	}
}

func (t *EditTool) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	pathStr, ok := args["path"].(string)
	if !ok {
		return nil, fmt.Errorf("missing path")
	}
	oldStr, ok := args["old_string"].(string)
	if !ok {
		return nil, fmt.Errorf("missing old_string")
	}
	newStr, ok := args["new_string"].(string)
	if !ok {
		return nil, fmt.Errorf("missing new_string")
	}

	// Validate path (basic sandboxing)
	// TODO: Use shared sandboxing util
	absRoot, _ := filepath.Abs(t.WorkspaceRoot)
	targetPath := filepath.Join(absRoot, pathStr) // simplified

	contentBytes, err := os.ReadFile(targetPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	content := string(contentBytes)

	// Strategy 1: Exact Match
	if strings.Count(content, oldStr) == 0 {
		return nil, fmt.Errorf("old_string not found in file")
	}
	if strings.Count(content, oldStr) > 1 {
		return nil, fmt.Errorf("old_string matches multiple locations; provide more context")
	}

	newContent := strings.Replace(content, oldStr, newStr, 1)

	if err := os.WriteFile(targetPath, []byte(newContent), 0644); err != nil {
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	return "Successfully replaced text.", nil
}