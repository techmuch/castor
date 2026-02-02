package llm

import (
	"encoding/json"
	"fmt"
)

// Role represents the role of a message sender.
type Role string

const (
	RoleUser   Role = "user"
	RoleModel  Role = "model"
	RoleSystem Role = "system"
	RoleTool   Role = "tool"
)

// Part is a marker interface for message content parts.
type Part interface {
	isPart()
}

// TextPart represents a text content part.
type TextPart struct {
	Text string `json:"text"`
}

func (TextPart) isPart() {}

// ToolCallPart represents a request from the model to call a tool.
type ToolCallPart struct {
	ID   string                 `json:"id"`
	Name string                 `json:"name"`
	Args map[string]interface{} `json:"args"`
}

func (ToolCallPart) isPart() {}

// ToolResponsePart represents the result of a tool execution.
type ToolResponsePart struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Content string `json:"content"` // JSON or text result
}

func (ToolResponsePart) isPart() {}

// Message represents a single message in the chat history.
type Message struct {
	Role    Role   `json:"role"`
	Content []Part `json:"content"`
}

// Custom Marshaling for Parts to handle interface type
type partWrapper struct {
	Type     string           `json:"type"`
	Text     *TextPart        `json:"text_part,omitempty"`
	ToolCall *ToolCallPart    `json:"tool_call_part,omitempty"`
	ToolResp *ToolResponsePart `json:"tool_resp_part,omitempty"`
}

func (m *Message) MarshalJSON() ([]byte, error) {
	type Alias Message
	var parts []partWrapper
	for _, p := range m.Content {
		switch v := p.(type) {
		case TextPart:
			parts = append(parts, partWrapper{Type: "text", Text: &v})
		case ToolCallPart:
			parts = append(parts, partWrapper{Type: "tool_call", ToolCall: &v})
		case ToolResponsePart:
			parts = append(parts, partWrapper{Type: "tool_resp", ToolResp: &v})
		default:
			return nil, fmt.Errorf("unknown part type: %T", p)
		}
	}
	return json.Marshal(&struct {
		Role    Role          `json:"role"`
		Content []partWrapper `json:"content"`
	}{
		Role:    m.Role,
		Content: parts,
	})
}

func (m *Message) UnmarshalJSON(data []byte) error {
	type partWrap struct {
		Type     string           `json:"type"`
		Text     *TextPart        `json:"text_part,omitempty"`
		ToolCall *ToolCallPart    `json:"tool_call_part,omitempty"`
		ToolResp *ToolResponsePart `json:"tool_resp_part,omitempty"`
	}
	var msg struct {
		Role    Role       `json:"role"`
		Content []partWrap `json:"content"`
	}
	if err := json.Unmarshal(data, &msg); err != nil {
		return err
	}
	m.Role = msg.Role
	m.Content = nil
	for _, p := range msg.Content {
		switch p.Type {
		case "text":
			if p.Text != nil {
				m.Content = append(m.Content, *p.Text)
			}
		case "tool_call":
			if p.ToolCall != nil {
				m.Content = append(m.Content, *p.ToolCall)
			}
		case "tool_resp":
			if p.ToolResp != nil {
				m.Content = append(m.Content, *p.ToolResp)
			}
		}
	}
	return nil
}

// ToolDefinition defines a tool that can be called by the model.
type ToolDefinition struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Schema      interface{} `json:"schema"`
}
