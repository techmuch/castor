package llm

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
	Text string
}

func (TextPart) isPart() {}

// ToolCallPart represents a request from the model to call a tool.
type ToolCallPart struct {
	ID   string
	Name string
	Args map[string]interface{}
}

func (ToolCallPart) isPart() {}

// ToolResponsePart represents the result of a tool execution.
type ToolResponsePart struct {
	ID      string
	Name    string
	Content string // JSON or text result
}

func (ToolResponsePart) isPart() {}

// Message represents a single message in the chat history.
type Message struct {
	Role    Role
	Content []Part
}

// ToolDefinition defines a tool that can be called by the model.
type ToolDefinition struct {
	Name        string
	Description string
	Schema      interface{}
}