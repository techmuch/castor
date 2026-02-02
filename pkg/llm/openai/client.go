package openai

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/techmuch/castor/pkg/llm"
)

type Client struct {
	BaseURL string
	APIKey  string
	Model   string
	HTTP    *http.Client
}

func NewClient(baseURL, apiKey, model string) *Client {
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}
	if model == "" {
		model = "gpt-3.5-turbo"
	}
	return &Client{
		BaseURL: strings.TrimRight(baseURL, "/"),
		APIKey:  apiKey,
		Model:   model,
		HTTP:    &http.Client{},
	}
}

type openAITool struct {
	Type     string `json:"type"`
	Function struct {
		Name        string      `json:"name"`
		Description string      `json:"description,omitempty"`
		Parameters  interface{} `json:"parameters,omitempty"`
	} `json:"function"`
}

type openAIToolCall struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
}

type openAIMessage struct {
	Role       string           `json:"role"`
	Content    string           `json:"content,omitempty"`
	ToolCalls  []openAIToolCall `json:"tool_calls,omitempty"`
	ToolCallID string           `json:"tool_call_id,omitempty"`
}

type chatRequest struct {
	Model       string          `json:"model"`
	Messages    []openAIMessage `json:"messages"`
	Stream      bool            `json:"stream"`
	Temperature float32         `json:"temperature,omitempty"`
	TopP        float32         `json:"top_p,omitempty"`
	Tools       []openAITool    `json:"tools,omitempty"`
}

type toolCallChunk struct {
	Index    int    `json:"index"`
	ID       string `json:"id"`
	Type     string `json:"type"`
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
}

type streamChoice struct {
	Delta struct {
		Content   string          `json:"content"`
		ToolCalls []toolCallChunk `json:"tool_calls"`
	} `json:"delta"`
	FinishReason string `json:"finish_reason"`
}

type streamResponse struct {
	Choices []streamChoice `json:"choices"`
}

func (c *Client) GenerateContent(ctx context.Context, history []llm.Message, opts llm.GenerateOptions) (<-chan llm.StreamEvent, error) {
	msgs := make([]openAIMessage, 0, len(history))
	for _, m := range history {
		msg := openAIMessage{
			Role: string(m.Role),
		}

		var contentParts []string
		
		for _, p := range m.Content {
			switch v := p.(type) {
			case llm.TextPart:
				contentParts = append(contentParts, v.Text)
			case llm.ToolCallPart:
				// Convert to OpenAI tool call
				argsJSON, _ := json.Marshal(v.Args)
				msg.ToolCalls = append(msg.ToolCalls, openAIToolCall{
					ID:   v.ID,
					Type: "function",
					Function: struct {
						Name      string `json:"name"`
						Arguments string `json:"arguments"`
					}{
						Name:      v.Name,
						Arguments: string(argsJSON),
					},
				})
			case llm.ToolResponsePart:
				msg.ToolCallID = v.ID
				contentParts = append(contentParts, v.Content)
			}
		}
		
		msg.Content = strings.Join(contentParts, "\n")
		// OpenAI Requirement: Content must be null if tool_calls are present and content is empty.
		// But in Go json omitempty works if string is empty.
		// However, for Assistant messages, content can be null.
		
		msgs = append(msgs, msg)
	}

	var tools []openAITool
	if len(opts.Tools) > 0 {
		for _, t := range opts.Tools {
			tools = append(tools, openAITool{
				Type: "function",
				Function: struct {
					Name        string      `json:"name"`
					Description string      `json:"description,omitempty"`
					Parameters  interface{} `json:"parameters,omitempty"`
				}{
					Name:        t.Name,
					Description: t.Description,
					Parameters:  t.Schema,
				},
			})
		}
	}

	reqBody := chatRequest{
		Model:       c.Model,
		Messages:    msgs,
		Stream:      true,
		Temperature: opts.Temperature,
		TopP:        opts.TopP,
		Tools:       tools,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.BaseURL+"/chat/completions", bytes.NewReader(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.APIKey)
	}

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("api returned status: %s", resp.Status)
	}

	ch := make(chan llm.StreamEvent)
	go func() {
		defer resp.Body.Close()
		defer close(ch)

		// State for accumulating tool calls
		type pendingToolCall struct {
			Index int
			ID    string
			Name  string
			Args  string
		}
		pendingCalls := make(map[int]*pendingToolCall)

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Text()
			if line == "" || !strings.HasPrefix(line, "data: ") {
				continue
			}
			data := strings.TrimPrefix(line, "data: ")
			if data == "[DONE]" {
				return
			}

			var streamResp streamResponse
			if err := json.Unmarshal([]byte(data), &streamResp); err != nil {
				ch <- llm.StreamEvent{Error: fmt.Errorf("unmarshal error: %w", err)}
				return
			}

			if len(streamResp.Choices) == 0 {
				continue
			}

			choice := streamResp.Choices[0]
			
			// Handle Text Content
			if choice.Delta.Content != "" {
				ch <- llm.StreamEvent{Delta: choice.Delta.Content}
			}

			// Handle Tool Calls
			for _, tc := range choice.Delta.ToolCalls {
				idx := tc.Index
				if _, exists := pendingCalls[idx]; !exists {
					pendingCalls[idx] = &pendingToolCall{Index: idx}
				}
				p := pendingCalls[idx]
				
				if tc.ID != "" {
					p.ID = tc.ID
				}
				if tc.Function.Name != "" {
					p.Name = tc.Function.Name
				}
				if tc.Function.Arguments != "" {
					p.Args += tc.Function.Arguments
				}
			}

			if choice.FinishReason == "tool_calls" || choice.FinishReason == "stop" {
				var finalCalls []llm.ToolCallPart
				for _, p := range pendingCalls {
					var argsMap map[string]interface{}
					if p.Args != "" {
						_ = json.Unmarshal([]byte(p.Args), &argsMap)
					}
					finalCalls = append(finalCalls, llm.ToolCallPart{
						ID:   p.ID,
						Name: p.Name,
						Args: argsMap,
					})
				}
				if len(finalCalls) > 0 {
					ch <- llm.StreamEvent{ToolCalls: finalCalls}
				}
			
pendingCalls = make(map[int]*pendingToolCall)
			}
		}
		if err := scanner.Err(); err != nil {
			ch <- llm.StreamEvent{Error: err}
		}
	}()

	return ch, nil
}

func (c *Client) EmbedContent(ctx context.Context, texts []string) ([][]float32, error) {
	return nil, fmt.Errorf("not implemented")
}
