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
	"github.com/techmuch/castor/pkg/llm"
)

// Ensure EditTool implements agent.Tool
var _ agent.Tool = (*EditTool)(nil)

// EditTool performs text replacements in files.
type EditTool struct {
	WorkspaceRoot string
	Provider      llm.Provider // Optional: for self-correction
}

func (t *EditTool) Name() string { return "replace" }

func (t *EditTool) Description() string {
	return "Replaces text within a file. Provide unique old_string to target the change. Supports exact, flexible, and self-correcting matching."
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

	absRoot, _ := filepath.Abs(t.WorkspaceRoot)
	targetPath := filepath.Join(absRoot, pathStr) 
	if !strings.HasPrefix(targetPath, absRoot) {
		return nil, fmt.Errorf("access denied: path outside workspace")
	}

	contentBytes, err := os.ReadFile(targetPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	content := string(contentBytes)

	// 0. Verify Hash
	if expectedHash != "" {
		hasher := sha256.New()
		hasher.Write(contentBytes)
		currentHash := hex.EncodeToString(hasher.Sum(nil))
		if currentHash != expectedHash {
			return nil, fmt.Errorf("file content has changed (hash mismatch). Expected %s, got %s. Please re-read the file.", expectedHash, currentHash)
		}
	}

	// Strategy 1: Exact Match
	if t.tryExact(targetPath, content, oldStr, newStr) {
		return "Successfully replaced text (exact match).", nil
	}

	// Strategy 2: Flexible Match (Ignore Whitespace)
	if t.tryFlexible(targetPath, content, oldStr, newStr) {
		return "Successfully replaced text (flexible match).", nil
	}

	// Strategy 3: Self-Correction (Fixer LLM)
	if t.Provider != nil {
		fixedOldStr, err := t.runFixer(ctx, content, oldStr)
		if err == nil && fixedOldStr != "" && fixedOldStr != oldStr {
			if t.tryExact(targetPath, content, fixedOldStr, newStr) {
				return fmt.Sprintf("Successfully replaced text (auto-corrected old_string)."), nil
			}
		}
	}

	return nil, fmt.Errorf("old_string not found (tried exact, flexible, and fixer)")
}

func (t *EditTool) tryExact(path, content, oldStr, newStr string) bool {
	if strings.Count(content, oldStr) == 1 {
		newContent := strings.Replace(content, oldStr, newStr, 1)
		return t.write(path, newContent) == nil
	}
	return false
}

func (t *EditTool) tryFlexible(path, content, oldStr, newStr string) bool {
	fields := strings.Fields(oldStr)
	if len(fields) == 0 {
		return false
	}
	
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
		return false
	}

	matches := re.FindAllStringIndex(content, -1)
	if len(matches) == 1 {
		matchIdx := matches[0]
		start, end := matchIdx[0], matchIdx[1]
		newContent := content[:start] + newStr + content[end:]
		return t.write(path, newContent) == nil
	}
	return false
}

func (t *EditTool) write(path string, content string) error {
	return os.WriteFile(path, []byte(content), 0644)
}

func (t *EditTool) runFixer(ctx context.Context, fileContent, brokenOldStr string) (string, error) {
	// Construct a prompt to find the correct string
	// We truncate fileContent if it's too huge to avoid token limits,
	// but for now assume it fits.
	
	systemPrompt := "You are a specialized text correction agent. Your job is to find the closest match for a string in a file."
	userPrompt := fmt.Sprintf(`I want to replace a string in a file, but I can't find an exact match. 
Here is the string I'm looking for (it might have wrong indentation or whitespace):
<<<<<<<<
%s
>>>>>>>>

Here is the actual file content:
<<<<<<<<
%s
>>>>>>>>

Find the unique string in the file content that most likely matches my intent. 
Return ONLY the exact string from the file content, with no other text.
If there is no clear match or multiple matches, return nothing.`, brokenOldStr, fileContent)

	history := []llm.Message{
		{Role: llm.RoleSystem, Content: []llm.Part{llm.TextPart{Text: systemPrompt}}},
		{Role: llm.RoleUser, Content: []llm.Part{llm.TextPart{Text: userPrompt}}},
	}

	opts := llm.GenerateOptions{Temperature: 0.0} // Deterministic
	stream, err := t.Provider.GenerateContent(ctx, history, opts)
	if err != nil {
		return "", err
	}

	var result strings.Builder
	for event := range stream {
		if event.Error != nil {
			return "", event.Error
		}
		result.WriteString(event.Delta)
	}

	return strings.TrimSpace(result.String()), nil
}
