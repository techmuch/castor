package llm

import (
	"encoding/json"
	"testing"
)

func TestMessageMarshaling(t *testing.T) {
	// Create a message with mixed parts
	msg := Message{
		Role: RoleUser,
		Content: []Part{
			TextPart{Text: "Explain this code:"},
			TextPart{Text: "func main() {}"},
		},
	}

	data, err := json.Marshal(&msg)
	if err != nil {
		t.Fatalf("Failed to marshal message: %v", err)
	}

	var decoded Message
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal message: %v", err)
	}

	if decoded.Role != RoleUser {
		t.Errorf("Expected role %s, got %s", RoleUser, decoded.Role)
	}
	if len(decoded.Content) != 2 {
		t.Errorf("Expected 2 parts, got %d", len(decoded.Content))
	}
	
	if len(decoded.Content) > 0 {
		p1, ok := decoded.Content[0].(TextPart)
		if !ok || p1.Text != "Explain this code:" {
			t.Errorf("Part 1 mismatch: %v", decoded.Content[0])
		}
	}
}

func TestToolCallMarshaling(t *testing.T) {
	msg := Message{
		Role: RoleModel,
		Content: []Part{
			TextPart{Text: "I will use a tool."},
			ToolCallPart{
				ID:   "call_123",
				Name: "read_file",
				Args: map[string]interface{}{"path": "test.txt"},
			},
		},
	}

	data, err := json.Marshal(&msg)
	if err != nil {
		t.Fatalf("Failed to marshal tool call message: %v", err)
	}

	var decoded Message
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal tool call message: %v", err)
	}

	if len(decoded.Content) != 2 {
		t.Fatalf("Expected 2 parts, got %d", len(decoded.Content))
	}

	p2, ok := decoded.Content[1].(ToolCallPart)
	if !ok {
		t.Fatalf("Expected Part 2 to be ToolCallPart, got %T", decoded.Content[1])
	}
	if p2.ID != "call_123" {
		t.Errorf("Expected ID call_123, got %s", p2.ID)
	}
	if p2.Name != "read_file" {
		t.Errorf("Expected Name read_file, got %s", p2.Name)
	}
	if p2.Args["path"] != "test.txt" {
		t.Errorf("Expected arg path=test.txt, got %v", p2.Args["path"])
	}
}