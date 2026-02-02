package agent

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/techmuch/castor/pkg/llm"
)

// Session represents a persistable agent state.
type Session struct {
	SystemPrompt string        `json:"system_prompt"`
	History      []llm.Message `json:"history"`
}

// SaveSession saves the agent's current state to a file.
func (a *Agent) SaveSession(path string) error {
	session := Session{
		SystemPrompt: a.SystemPrompt,
		History:      a.History,
	}

	data, err := json.MarshalIndent(session, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}

	return os.WriteFile(path, data, 0644)
}

// LoadSession loads an agent's state from a file.
func (a *Agent) LoadSession(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read session file: %w", err)
	}

	var session Session
	if err := json.Unmarshal(data, &session); err != nil {
		return fmt.Errorf("failed to unmarshal session: %w", err)
	}

	a.SystemPrompt = session.SystemPrompt
	a.History = session.History
	return nil
}
