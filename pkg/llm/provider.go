package llm

import "context"

// GenerateOptions contains configuration for the generation request.
type GenerateOptions struct {
	Temperature float32
	TopP        float32
	StopTokens  []string
	Tools       []ToolDefinition
	// JSONSchema can be added here when we implement structured output support
}

// StreamEvent represents a single event in the response stream.
type StreamEvent struct {
	// Delta is the new text fragment generated.
	Delta string
	// ToolCalls contains any tool calls generated in this chunk (usually at the end).
	ToolCalls []ToolCallPart
	// Error indicates if an error occurred during streaming.
	Error error
}

// Provider defines the interface that all LLM backends must implement.
type Provider interface {
	// GenerateContent sends a chat history to the model and returns a channel of events.
	// The channel is closed when generation is complete.
	GenerateContent(ctx context.Context, history []Message, opts GenerateOptions) (<-chan StreamEvent, error)

	// EmbedContent returns vector embeddings for the given texts.
	EmbedContent(ctx context.Context, texts []string) ([][]float32, error)
}