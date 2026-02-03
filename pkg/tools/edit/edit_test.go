package edit

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
)

func TestEditTool(t *testing.T) {
	// Setup workspace
	tmpDir, err := os.MkdirTemp("", "castor_edit_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	targetFile := filepath.Join(tmpDir, "code.js")
	initialContent := "function hello() {\n    console.log('hello world');\n}"
	
	setupFile := func() {
		if err := os.WriteFile(targetFile, []byte(initialContent), 0644); err != nil {
			t.Fatal(err)
		}
	}

	tool := &EditTool{WorkspaceRoot: tmpDir}
	ctx := context.Background()

	t.Run("ExactMatch", func(t *testing.T) {
		setupFile()
		
		oldStr := "console.log('hello world');"
		newStr := "console.log('hello universe');"
		
		_, err := tool.Execute(ctx, map[string]interface{}{
			"path": "code.js",
			"old_string": oldStr,
			"new_string": newStr,
		})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		content, _ := os.ReadFile(targetFile)
		if string(content) != "function hello() {\n    console.log('hello universe');\n}" {
			t.Errorf("content mismatch: %s", string(content))
		}
	})

	t.Run("FlexibleMatch", func(t *testing.T) {
		setupFile()
		
		// File: `function hello() { ... }`
		// Setup file with weird spacing
		weirdContent := "function   hello()   {\n    console.log('hello world');\n}"
		os.WriteFile(targetFile, []byte(weirdContent), 0644)
		
		cleanOld := "function hello() {"
		newVal := "function greetings() {"
		
		_, err := tool.Execute(ctx, map[string]interface{}{
			"path": "code.js",
			"old_string": cleanOld,
			"new_string": newVal,
		})
		if err != nil {
			t.Errorf("unexpected error on flexible match: %v", err)
		}
		
		content, _ := os.ReadFile(targetFile)
		expected := "function greetings() {\n    console.log('hello world');\n}"
		if string(content) != expected {
			t.Errorf("content mismatch.\nGot: %s\nWant: %s", string(content), expected)
		}
	})

	t.Run("HashVerificationSuccess", func(t *testing.T) {
		setupFile()
		
		hasher := sha256.New()
		hasher.Write([]byte(initialContent))
		hash := hex.EncodeToString(hasher.Sum(nil))

		_, err := tool.Execute(ctx, map[string]interface{}{
			"path": "code.js",
			"old_string": "hello world",
			"new_string": "hash check",
			"expected_hash": hash,
		})
		if err != nil {
			t.Errorf("expected success with correct hash, got: %v", err)
		}
	})

	t.Run("HashVerificationFailure", func(t *testing.T) {
		setupFile()
		
		_, err := tool.Execute(ctx, map[string]interface{}{
			"path": "code.js",
			"old_string": "hello world",
			"new_string": "hash check",
			"expected_hash": "badhash123",
		})
		if err == nil {
			t.Error("expected error with incorrect hash, got success")
		}
	})
}
