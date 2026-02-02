package agent

import (
	"context"
	"fmt"

	"github.com/techmuch/castor/pkg/llm"
)

// Investigator represents a specialized agent loop for research tasks.
type Investigator struct {
	Agent *Agent
}

// InvestigationReport represents the structured output of an investigation.
type InvestigationReport struct {
	Goal          string   `json:"goal"`
	Findings      []string `json:"findings"`
	FilesExplored []string `json:"files_explored"`
	Conclusion    string   `json:"conclusion"`
}

// Investigate executes the scratchpad loop to solve a complex query.
func (inv *Investigator) Investigate(ctx context.Context, goal string) (*InvestigationReport, error) {
	// 1. Setup System Prompt specialized for investigation
	sysPrompt := `You are a Codebase Investigator. Your goal is to answer the user's query by exploring the codebase.
You must maintain a structured thought process.
Do not guess. Verify facts by reading files.
You have access to 'ls', 'read_file', and 'grep' (if available). 
Use them to explore the file structure and content.

When you have gathered enough information, call the 'report_findings' tool to finalize the task.
`
	reportTool := &ReportTool{}
	inv.Agent.RegisterTool(reportTool)
	
	originalPrompt := inv.Agent.SystemPrompt
	inv.Agent.SystemPrompt = sysPrompt + "\nOriginal Instructions: " + originalPrompt
	originalHistory := inv.Agent.History
	inv.Agent.History = []llm.Message{
		{Role: llm.RoleSystem, Content: []llm.Part{llm.TextPart{Text: inv.Agent.SystemPrompt}}},
		{Role: llm.RoleUser, Content: []llm.Part{llm.TextPart{Text: "Investigate: " + goal}}},
	}

	defer func() {
		// Restore agent state
		inv.Agent.SystemPrompt = originalPrompt
		inv.Agent.History = originalHistory
		delete(inv.Agent.Tools, reportTool.Name())
	}()

	maxTurns := 15
	for i := 0; i < maxTurns; i++ {
		var stream <-chan llm.StreamEvent
		var err error
		
		if i == 0 {
			inv.Agent.History = []llm.Message{
				{Role: llm.RoleSystem, Content: []llm.Part{llm.TextPart{Text: inv.Agent.SystemPrompt}}},
			}
			stream, err = inv.Agent.Chat(ctx, "Investigate: "+goal)
		} else {
			stream, err = inv.Agent.Chat(ctx, "Continue. If you have enough info, call report_findings.")
		}

		if err != nil {
			return nil, err
		}

		for event := range stream {
			if event.Error != nil {
				return nil, event.Error
			}
		}

		if reportTool.Report != nil {
			return reportTool.Report, nil
		}
	}

	return nil, fmt.Errorf("investigation timed out after %d turns without a report", maxTurns)
}

// ReportTool is a special tool for the investigator to submit its final report.
type ReportTool struct {
	Report *InvestigationReport
}

func (t *ReportTool) Name() string { return "report_findings" }
func (t *ReportTool) Description() string { return "Submit the final investigation report." }
func (t *ReportTool) Schema() interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"goal":           map[string]interface{}{"type": "string"},
			"findings":       map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}},
			"files_explored": map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}},
			"conclusion":     map[string]interface{}{"type": "string"},
		},
		"required": []string{"goal", "findings", "conclusion"},
	}
}

func (t *ReportTool) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	report := &InvestigationReport{}
	
	if g, ok := args["goal"].(string); ok {
		report.Goal = g
	}
	if c, ok := args["conclusion"].(string); ok {
		report.Conclusion = c
	}
	
	if findings, ok := args["findings"].([]interface{}); ok {
		for _, f := range findings {
			if s, ok := f.(string); ok {
				report.Findings = append(report.Findings, s)
			}
		}
	}
	
	if files, ok := args["files_explored"].([]interface{}); ok {
		for _, f := range files {
			if s, ok := f.(string); ok {
				report.FilesExplored = append(report.FilesExplored, s)
			}
		}
	}

	t.Report = report
	return "Report submitted successfully.", nil
}