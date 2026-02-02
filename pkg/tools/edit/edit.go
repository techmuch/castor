package edit

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
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
	return "Replaces text within a file. Provide unique old_string to target the change. Supports exact and flexible (whitespace-insensitive) matching."
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
			"expected_hash": map[string]interface{}{
				"type":        "string",
				"description": "SHA-256 hash of the file content before editing. Optional but recommended for safety.",
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
	
	// Optional hash check
	expectedHash, _ := args["expected_hash"].(string)

	// Validate path (basic sandboxing)
	absRoot, _ := filepath.Abs(t.WorkspaceRoot)
	// TODO: Use shared sandboxing util from fs package if exported, or duplicate logic
	targetPath := filepath.Join(absRoot, pathStr) 
	if !strings.HasPrefix(targetPath, absRoot) {
		return nil, fmt.Errorf("access denied: path outside workspace")
	}

	contentBytes, err := os.ReadFile(targetPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	content := string(contentBytes)

	// 0. Verify Hash if provided
	if expectedHash != "" {
		hasher := sha256.New()
		hasher.Write(contentBytes)
		currentHash := hex.EncodeToString(hasher.Sum(nil))
		if currentHash != expectedHash {
			return nil, fmt.Errorf("file content has changed (hash mismatch). Expected %s, got %s. Please re-read the file.", expectedHash, currentHash)
		}
	}

	// Strategy 1: Exact Match
	if count := strings.Count(content, oldStr); count > 0 {
		if count > 1 {
			return nil, fmt.Errorf("exact match found multiple times (%d); provide more context to be unique", count)
		}
		newContent := strings.Replace(content, oldStr, newStr, 1)
		return t.write(targetPath, newContent)
	}

	// Strategy 2: Flexible Match (Ignore Whitespace)
	// We normalize both content and oldStr to find a match.
	// This is complex because we need to map the normalized match back to the original string indices to replace it.
	
	// Simple approach: exact match on lines, ignoring leading/trailing whitespace?
	// Robust approach: Tokenize or Regex.
	
	// Let's try a Regex that turns whitespace in oldStr into `\s+`
	// We first replace specific whitespace chars with spaces to simplify
	fields := strings.Fields(oldStr)
	if len(fields) == 0 {
		// oldStr was just whitespace?
		return nil, fmt.Errorf("old_string is empty or only whitespace")
	}
	
	// Reconstruct pattern: field1\s+field2\s+...
	var patternBuilder strings.Builder
	for i, field := range fields {
		if i > 0 {
			patternBuilder.WriteString(`\s+`)
		}
		patternBuilder.WriteString(regexp.QuoteMeta(field))
	}
	flexiblePattern := patternBuilder.String()
	
	re, err := regexp.Compile(flexiblePattern)
	if err != nil {
		return nil, fmt.Errorf("failed to build flexible match regex: %w", err)
	}

	matches := re.FindAllStringIndex(content, -1)
	if len(matches) == 0 {
		return nil, fmt.Errorf("old_string not found (tried exact and flexible match)")
	}
	if len(matches) > 1 {
		return nil, fmt.Errorf("flexible match found multiple times (%d); provide more context", len(matches))
	}

	// Perform replacement on the specific range found
	matchIdx := matches[0]
	start, end := matchIdx[0], matchIdx[1]
	
	newContent := content[:start] + newStr + content[end:]
	return t.write(targetPath, newContent)
}

func (t *EditTool) write(path string, content string) (interface{}, error) {
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return nil, fmt.Errorf("failed to write file: %w", err)
	}
	return "Successfully replaced text.", nil
}