package fs

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestSandboxing(t *testing.T) {
	// Setup temporary workspace
	tmpDir, err := os.MkdirTemp("", "castor_test_ws")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a file inside workspace
	safeFile := filepath.Join(tmpDir, "safe.txt")
	if err := os.WriteFile(safeFile, []byte("safe content"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create a subdirectory inside workspace
	subDir := filepath.Join(tmpDir, "subdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create a file outside workspace
	outsideDir, err := os.MkdirTemp("", "castor_outside")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(outsideDir)
	
	outsideFile := filepath.Join(outsideDir, "secret.txt")
	if err := os.WriteFile(outsideFile, []byte("secret content"), 0644); err != nil {
		t.Fatal(err)
	}

	// Test ReadFileTool
	t.Run("ReadFileTool", func(t *testing.T) {
		tool := &ReadFileTool{WorkspaceRoot: tmpDir}
		ctx := context.Background()

		// Case 1: Read safe file (should succeed)
		res, err := tool.Execute(ctx, map[string]interface{}{"path": "safe.txt"})
		if err != nil {
			t.Errorf("expected success reading safe file, got error: %v", err)
		}
		if content, ok := res.(string); !ok || content != "safe content" {
			t.Errorf("expected 'safe content', got %v", res)
		}

		// Case 2: Read file via relative path (should succeed)
		res, err = tool.Execute(ctx, map[string]interface{}{"path": "./safe.txt"})
		if err != nil {
			t.Errorf("expected success reading ./safe.txt, got error: %v", err)
		}

		// Case 3: Read outside file using ../ (should fail)
		// We construct a path that tries to traverse out
		relPath, _ := filepath.Rel(tmpDir, outsideFile)
		_, err = tool.Execute(ctx, map[string]interface{}{"path": relPath})
		if err == nil {
			t.Error("expected error reading outside file, got success")
		} else {
			// Check if error message mentions access denied or outside workspace
			if len(err.Error()) == 0 {
				t.Error("expected meaningful error message")
			}
		}

		// Case 4: Read absolute path outside workspace (should fail)
		_, err = tool.Execute(ctx, map[string]interface{}{"path": outsideFile})
		if err == nil {
			t.Error("expected error reading absolute outside path, got success")
		}
	})

	// Test ListDirTool
	t.Run("ListDirTool", func(t *testing.T) {
		tool := &ListDirTool{WorkspaceRoot: tmpDir}
		ctx := context.Background()

		// Case 1: List root (should succeed)
		res, err := tool.Execute(ctx, map[string]interface{}{"path": "."})
		if err != nil {
			t.Errorf("expected success listing root, got error: %v", err)
		}
		list, ok := res.([]string)
		if !ok {
			t.Errorf("expected []string, got %T", res)
		}
		// Should contain safe.txt and subdir/
		foundSafe := false
		foundSub := false
		for _, item := range list {
			if item == "safe.txt" { foundSafe = true }
			if item == "subdir/" { foundSub = true }
		}
		if !foundSafe || !foundSub {
			t.Errorf("listing missing expected items: %v", list)
		}

		// Case 2: List outside dir (should fail)
		_, err = tool.Execute(ctx, map[string]interface{}{"path": outsideDir})
		if err == nil {
			t.Error("expected error listing outside dir, got success")
		}
	})
}
