package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Configuration for the integration tests
// Defaults match your running Ollama instance
const (
	DefaultBaseURL = "http://localhost:11434/v1"
	DefaultModel   = "qwen3:8b"
)

var (
	baseURL = os.Getenv("CASTOR_TEST_URL")
	model   = os.Getenv("CASTOR_TEST_MODEL")
)

func init() {
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}
	if model == "" {
		model = DefaultModel
	}
}

// buildCastor compiles the binary for testing
func buildCastor(t *testing.T) string {
	t.Helper()
	tempDir := t.TempDir()
	binaryPath := filepath.Join(tempDir, "castor")
	
	// We assume the test is run from the package directory, so we go up two levels
	cmd := exec.Command("go", "build", "-o", binaryPath, "../../cmd/castor")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to build castor: %v\nOutput: %s", err, output)
	}
	return binaryPath
}

func TestIntegration_Tools(t *testing.T) {
	// 1. Build Binary
	binary := buildCastor(t)

	// 2. Setup Safe Workspace
	workspace := t.TempDir()
	
	// Create some files for the agent to interact with
	readmePath := filepath.Join(workspace, "README.txt")
	err := os.WriteFile(readmePath, []byte("Welcome to the test project.\nThis is a safe space."), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Helper to run castor
	runCastor := func(prompt string) string {
		cmd := exec.Command(binary,
			"-url", baseURL,
			"-model", model,
			"-w", workspace, // SANDBOXED to temp dir
			prompt,
		)
		// Set a dummy key if not present, required by client validation
		cmd.Env = append(os.Environ(), "OPENAI_API_KEY=test-key")
		
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Logf("Command failed: %v", err)
			t.Logf("Output: %s", output)
			// Don't fail immediately, let the checks fail if output is missing
		}
		return string(output)
	}

	t.Run("ListFiles", func(t *testing.T) {
		output := runCastor("list the files in the current directory")
		t.Logf("Output: %s", output)

		// Check for tool usage log
		if !strings.Contains(output, "Tool Call: list_directory") {
			t.Error("Agent did not call list_directory tool")
		}
		// Check for file existence in output (Agent usually mentions it)
		if !strings.Contains(output, "README.txt") {
			t.Error("Agent did not mention README.txt in the response")
		}
	})

	t.Run("ReadFile", func(t *testing.T) {
		output := runCastor("read the content of README.txt")
		t.Logf("Output: %s", output)

		if !strings.Contains(output, "Tool Call: read_file") {
			t.Error("Agent did not call read_file tool")
		}
		if !strings.Contains(output, "safe space") {
			t.Error("Agent did not report the file content 'safe space'")
		}
	})

	t.Run("EditFile", func(t *testing.T) {
		// Instructions: Change "safe space" to "dangerous place"
		output := runCastor("change the text 'safe space' to 'dangerous place' in README.txt")
		t.Logf("Output: %s", output)

		if !strings.Contains(output, "Tool Call: replace") {
			t.Error("Agent did not call replace tool")
		}

		// Verify the file was actually changed on disk
		content, err := os.ReadFile(readmePath)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(content), "dangerous place") {
			t.Errorf("File content was not updated. Got: %s", string(content))
		}
	})
}