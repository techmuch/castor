package mcp

import (
	"context"
	"encoding/json"

	"github.com/techmuch/castor/pkg/agent"
)

// Client represents a connection to an MCP server.
type Client interface {
	// Initialize performs the handshake with the server.
	Initialize(ctx context.Context) error
	
	// ListTools retrieves the tools available on the server.
	ListTools(ctx context.Context) ([]agent.Tool, error)
	
	// CallTool calls a specific tool on the server.
	CallTool(ctx context.Context, name string, args map[string]interface{}) (interface{}, error)
	
	// Close terminates the connection.
	Close() error
}

// Transport defines the communication channel (Stdio, SSE).
type Transport interface {
	Send(ctx context.Context, msg JSONRPCMessage) error
	Receive(ctx context.Context) (JSONRPCMessage, error)
	Close() error
}

// JSONRPCMessage represents a JSON-RPC 2.0 message.
type JSONRPCMessage struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
	ID      *int64          `json:"id,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *JSONRPCError   `json:"error,omitempty"`
}

type JSONRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}
