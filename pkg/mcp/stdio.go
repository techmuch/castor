package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
)

// StdioTransport implements Transport over stdin/stdout of a subprocess.
type StdioTransport struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout io.ReadCloser
	stderr io.ReadCloser
	
	scanner *bufio.Scanner
	mu      sync.Mutex
}

// NewStdioTransport starts a subprocess and returns a transport connected to it.
func NewStdioTransport(command string, args []string) (*StdioTransport, error) {
	cmd := exec.Command(command, args...)
	
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to get stdin pipe: %w", err)
	}
	
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to get stdout pipe: %w", err)
	}
	
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to get stderr pipe: %w", err)
	}
	
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start command: %w", err)
	}

	// Forward stderr to parent stderr for debugging
	go io.Copy(os.Stderr, stderr)

	return &StdioTransport{
		cmd:     cmd,
		stdin:   stdin,
		stdout:  stdout,
		stderr:  stderr,
		scanner: bufio.NewScanner(stdout),
	},
	nil
}

func (t *StdioTransport) Send(ctx context.Context, msg JSONRPCMessage) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal error: %w", err)
	}
	
	// MCP uses JSON-RPC over stdio, typically newline delimited
	if _, err := t.stdin.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("write error: %w", err)
	}
	return nil
}

func (t *StdioTransport) Receive(ctx context.Context) (JSONRPCMessage, error) {
	// Note: This implementation is synchronous and blocking on scanner.Scan().
	// A more robust implementation would handle context cancellation.
	
	if !t.scanner.Scan() {
		if err := t.scanner.Err(); err != nil {
			return JSONRPCMessage{}, fmt.Errorf("scan error: %w", err)
		}
		return JSONRPCMessage{}, io.EOF
	}
	
	var msg JSONRPCMessage
	if err := json.Unmarshal(t.scanner.Bytes(), &msg); err != nil {
		return JSONRPCMessage{}, fmt.Errorf("unmarshal error: %w", err)
	}
	
	return msg, nil
}

func (t *StdioTransport) Close() error {
	t.stdin.Close()
	t.stdout.Close()
	t.stderr.Close()
	return t.cmd.Process.Kill()
}
