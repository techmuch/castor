package fs

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/techmuch/castor/pkg/agent"
)

// Ensure tools implement agent.Tool
var _ agent.Tool = (*ListDirTool)(nil)
var _ agent.Tool = (*ReadFileTool)(nil)

// ensureInWorkspace checks if the target path is within the allowed workspace.
func ensureInWorkspace(root, target string) (string, error) {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return "", fmt.Errorf("invalid root path: %w", err)
	}

	// Handle relative paths
	absTarget := target
	if !filepath.IsAbs(target) {
		absTarget = filepath.Join(absRoot, target)
	} else {
		absTarget = filepath.Clean(target)
	}

	if !strings.HasPrefix(absTarget, absRoot) {
		return "", fmt.Errorf("access denied: path %s is outside workspace %s", target, root)
	}

	return absTarget, nil
}

// --- List Directory Tool ---

type ListDirTool struct {
	WorkspaceRoot string
}

func (t *ListDirTool) Name() string { return "list_directory" }

func (t *ListDirTool) Description() string {
	return "Lists files and subdirectories in a specific directory."
}

func (t *ListDirTool) Schema() interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"path": map[string]interface{}{
				"type":        "string",
				"description": "The directory path relative to the workspace root.",
			},
		},
		"required": []string{"path"},
	}
}

func (t *ListDirTool) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	pathStr, ok := args["path"].(string)
	if !ok {
		pathStr = "."
	}

	targetPath, err := ensureInWorkspace(t.WorkspaceRoot, pathStr)
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(targetPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read dir: %w", err)
	}

	var results []string
	for _, e := range entries {
		suffix := ""
		if e.IsDir() {
			suffix = "/"
		}
		results = append(results, e.Name()+suffix)
	}
	return results, nil
}

// --- Read File Tool ---

type ReadFileTool struct {
	WorkspaceRoot string
}

func (t *ReadFileTool) Name() string { return "read_file" }

func (t *ReadFileTool) Description() string {
	return "Reads the content of a file."
}

func (t *ReadFileTool) Schema() interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"path": map[string]interface{}{
				"type":        "string",
				"description": "The file path relative to the workspace root.",
			},
		},
		"required": []string{"path"},
	}
}

func (t *ReadFileTool) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	pathStr, ok := args["path"].(string)
	if !ok {
		return nil, fmt.Errorf("missing argument: path")
	}

	targetPath, err := ensureInWorkspace(t.WorkspaceRoot, pathStr)
	if err != nil {
		return nil, err
	}

	content, err := os.ReadFile(targetPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return string(content), nil
}