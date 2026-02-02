package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/techmuch/castor/pkg/llm"
)

// Agent orchestrates the interaction between the user, the LLM, and tools.
type Agent struct {
	Provider     llm.Provider
	Tools        map[string]Tool
	History      []llm.Message
	SystemPrompt string
	MaxTurns     int
}

// New creates a new Agent instance.
func New(provider llm.Provider, systemPrompt string) *Agent {
	agent := &Agent{
		Provider:     provider,
		Tools:        make(map[string]Tool),
		SystemPrompt: systemPrompt,
		History:      make([]llm.Message, 0),
		MaxTurns:     10, // Default safety limit
	}

	// Initialize history with system prompt if provided
	if systemPrompt != "" {
		agent.History = append(agent.History, llm.Message{
			Role:    llm.RoleSystem,
			Content: []llm.Part{llm.TextPart{Text: systemPrompt}},
		})
	}

	return agent
}

// RegisterTool adds a tool to the agent's registry.
func (a *Agent) RegisterTool(t Tool) {
	a.Tools[t.Name()] = t
}

// Chat sends a message to the agent and returns a stream of events.
// It handles the "Think-Act" loop: Model -> Tool Call -> Execution -> Model ...
func (a *Agent) Chat(ctx context.Context, input string) (<-chan llm.StreamEvent, error) {
	// Add user message to history
	userMsg := llm.Message{
		Role:    llm.RoleUser,
		Content: []llm.Part{llm.TextPart{Text: input}},
	}
	a.History = append(a.History, userMsg)

	outCh := make(chan llm.StreamEvent)

	go func() {
		defer close(outCh)
		
		for turn := 0; turn < a.MaxTurns; turn++ {
			// Prepare tools
			var toolDefs []llm.ToolDefinition
			for _, t := range a.Tools {
				toolDefs = append(toolDefs, llm.ToolDefinition{
					Name:        t.Name(),
					Description: t.Description(),
					Schema:      t.Schema(),
				})
			}

			opts := llm.GenerateOptions{
				Temperature: 0.7,
				Tools:       toolDefs,
			}

			stream, err := a.Provider.GenerateContent(ctx, a.History, opts)
			if err != nil {
				outCh <- llm.StreamEvent{Error: err}
				return
			}

			var fullText strings.Builder
			var toolCalls []llm.ToolCallPart

			// Consume stream
			for event := range stream {
				if event.Error != nil {
					outCh <- event
					return
				}
				
				if event.Delta != "" {
					fullText.WriteString(event.Delta)
					// Pass text to user
					outCh <- event
				}
				
				if len(event.ToolCalls) > 0 {
					toolCalls = append(toolCalls, event.ToolCalls...)
					// Pass tool calls to user (optional, for UI feedback)
					outCh <- event
				}
			}

			// Add model response to history
			modelMsg := llm.Message{
				Role:    llm.RoleModel,
				Content: []llm.Part{},
			}
			if fullText.Len() > 0 {
				modelMsg.Content = append(modelMsg.Content, llm.TextPart{Text: fullText.String()})
			}
			// We store tool calls in history too, if they exist
			// Note: Our current Message structure treats content as a flat list.
			// We need to properly represent ToolCalls as parts.
			for _, tc := range toolCalls {
				modelMsg.Content = append(modelMsg.Content, tc)
			}
			a.History = append(a.History, modelMsg)

			// If no tool calls, we are done
			if len(toolCalls) == 0 {
				return
			}

			// Execute Tools
			for _, tc := range toolCalls {
				tool, exists := a.Tools[tc.Name]
				var resultStr string
				
				if !exists {
					resultStr = fmt.Sprintf("Error: Tool '%s' not found.", tc.Name)
				} else {
					res, err := tool.Execute(ctx, tc.Args)
					if err != nil {
						resultStr = fmt.Sprintf("Error executing tool: %v", err)
					} else {
						// Marshal result to JSON string
						resBytes, _ := json.Marshal(res)
						resultStr = string(resBytes)
					}
				}

				// Add tool result to history
				// Note: Tool responses usually need to link back to the call ID.
				// OpenAI expects role "tool", tool_call_id, and content.
				// Our `ToolResponsePart` has ID.
				
				// We create a new message for EACH tool response?
				// Usually yes, role="tool".
				toolMsg := llm.Message{
					Role: llm.RoleTool,
					Content: []llm.Part{
						llm.ToolResponsePart{
							ID:      tc.ID,
							Name:    tc.Name,
							Content: resultStr,
						},
					},
				}
				a.History = append(a.History, toolMsg)
			}
			// Loop continues to next turn to feed tool results back to LLM
		}
	}()

	return outCh, nil
}