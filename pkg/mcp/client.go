package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"sync/atomic"

	"github.com/techmuch/castor/pkg/agent"
)

type MCPClient struct {
	transport Transport
	nextID    int64
}

func NewClient(t Transport) *MCPClient {
	return &MCPClient{
		transport: t,
		nextID:    1,
	}
}

func (c *MCPClient) Initialize(ctx context.Context) error {
	// 1. Send initialize request
	req := JSONRPCMessage{
		JSONRPC: "2.0",
		Method:  "initialize",
		Params: json.RawMessage(`{
			"protocolVersion": "2024-11-05",
			"capabilities": {
				"roots": { "listChanged": false },
				"sampling": {}
			},
			"clientInfo": {
				"name": "castor",
				"version": "0.1.0"
			}
		}`),
		ID: c.newID(),
	}

	if err := c.transport.Send(ctx, req); err != nil {
		return err
	}

	// 2. Wait for response
	// TODO: In a real implementation, we need a dispatch loop to match IDs.
	// For now, assuming synchronous response order for simple handshake.
	resp, err := c.transport.Receive(ctx)
	if err != nil {
		return err
	}

	if resp.Error != nil {
		return fmt.Errorf("mcp init error: %s", resp.Error.Message)
	}

	// 3. Send initialized notification
	notif := JSONRPCMessage{
		JSONRPC: "2.0",
		Method:  "notifications/initialized",
	}
	return c.transport.Send(ctx, notif)
}

func (c *MCPClient) ListTools(ctx context.Context) ([]agent.Tool, error) {
	req := JSONRPCMessage{
		JSONRPC: "2.0",
		Method:  "tools/list",
		ID:      c.newID(),
	}
	
	if err := c.transport.Send(ctx, req); err != nil {
		return nil, err
	}

	resp, err := c.transport.Receive(ctx)
	if err != nil {
		return nil, err
	}
	
	if resp.Error != nil {
		return nil, fmt.Errorf("list tools error: %s", resp.Error.Message)
	}

	var result struct {
		Tools []struct {
			Name        string          `json:"name"`
			Description string          `json:"description"`
			InputSchema json.RawMessage `json:"inputSchema"`
		} `json:"tools"`
	}
	
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return nil, fmt.Errorf("failed to parse tools list: %w", err)
	}

	var tools []agent.Tool
	for _, t := range result.Tools {
		tools = append(tools, &mcpTool{
			client: c,
			name:   t.Name,
			desc:   t.Description,
			schema: t.InputSchema,
		})
	}

	return tools, nil
}

func (c *MCPClient) CallTool(ctx context.Context, name string, args map[string]interface{}) (interface{}, error) {
	paramsJSON, _ := json.Marshal(map[string]interface{}{
		"name":      name,
		"arguments": args,
	})

	req := JSONRPCMessage{
		JSONRPC: "2.0",
		Method:  "tools/call",
		Params:  paramsJSON,
		ID:      c.newID(),
	}

	if err := c.transport.Send(ctx, req); err != nil {
		return nil, err
	}

	resp, err := c.transport.Receive(ctx)
	if err != nil {
		return nil, err
	}
	
	if resp.Error != nil {
		return nil, fmt.Errorf("tool call error: %s (data: %v)", resp.Error.Message, resp.Error.Data)
	}

	// MCP returns { content: [ { type: "text", text: "..." } ], isError: bool }
	var result struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
		IsError bool `json:"isError"`
	}
	
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return nil, fmt.Errorf("failed to parse tool result: %w. Raw: %s", err, string(resp.Result))
	}

	var output string
	for _, c := range result.Content {
		if c.Type == "text" {
			output += c.Text
		}
	}
	
	if result.IsError {
		return nil, fmt.Errorf("tool reported error: %s", output)
	}

	return output, nil
}

func (c *MCPClient) Close() error {
	return c.transport.Close()
}

func (c *MCPClient) newID() *int64 {
	id := atomic.AddInt64(&c.nextID, 1)
	return &id
}

// mcpTool adapts a remote MCP tool to the agent.Tool interface
type mcpTool struct {
	client *MCPClient
	name   string
	desc   string
	schema json.RawMessage
}

func (t *mcpTool) Name() string { return t.name }
func (t *mcpTool) Description() string { return t.desc }
func (t *mcpTool) Schema() interface{} {
	var s interface{}
	_ = json.Unmarshal(t.schema, &s)
	return s
}
func (t *mcpTool) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	return t.client.CallTool(ctx, t.name, args)
}