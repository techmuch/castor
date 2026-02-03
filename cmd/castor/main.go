package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/techmuch/castor/pkg/agent"
	"github.com/techmuch/castor/pkg/llm/openai"
	"github.com/techmuch/castor/pkg/mcp"
	"github.com/techmuch/castor/pkg/tools/edit"
	"github.com/techmuch/castor/pkg/tools/fs"
	"github.com/techmuch/castor/pkg/tui"
)

func main() {
	apiKey := os.Getenv("OPENAI_API_KEY")
	model := flag.String("model", "gpt-3.5-turbo", "LLM model to use")
	baseURL := flag.String("url", "", "Base URL for OpenAI-compatible API (e.g. http://localhost:11434/v1)")
	systemPrompt := flag.String("system", "You are a helpful assistant with access to files.", "System prompt")
	interactive := flag.Bool("i", false, "Interactive mode (REPL)")
	gui := flag.Bool("tui", false, "Start Terminal UI")
	workspace := flag.String("w", ".", "Workspace root directory")
	sessionPath := flag.String("session", "", "Path to session file for persistence")
	mcpCmd := flag.String("mcp", "", "Command to run an MCP server")
	investigate := flag.Bool("investigate", false, "Run in investigator mode (requires prompt)")
	flag.Parse()

	if apiKey == "" {
		fmt.Println("Error: OPENAI_API_KEY environment variable is required.")
		os.Exit(1)
	}

	client := openai.NewClient(*baseURL, apiKey, *model)
	ag := agent.New(client, *systemPrompt)
	
	// Register Tools
	ag.RegisterTool(&fs.ListDirTool{WorkspaceRoot: *workspace})
	ag.RegisterTool(&fs.ReadFileTool{WorkspaceRoot: *workspace})
	ag.RegisterTool(&edit.EditTool{
		WorkspaceRoot: *workspace,
		Provider:      client,
	})
	
	ctx := context.Background()

	// Connect to MCP Server
	if *mcpCmd != "" {
		parts := strings.Fields(*mcpCmd)
		if len(parts) > 0 {
			transport, err := mcp.NewStdioTransport(parts[0], parts[1:])
			if err != nil {
				fmt.Printf("Error starting MCP server: %v\n", err)
				os.Exit(1)
			}
			defer transport.Close()

			mcpClient := mcp.NewClient(transport)
			if err := mcpClient.Initialize(ctx); err != nil {
				fmt.Printf("Error initializing MCP client: %v\n", err)
				os.Exit(1)
			}
			defer mcpClient.Close()

		
tools, err := mcpClient.ListTools(ctx)
			if err != nil {
				fmt.Printf("Error listing MCP tools: %v\n", err)
				os.Exit(1)
			}

			fmt.Printf("Connected to MCP server. Discovered %d tools:\n", len(tools))
			for _, t := range tools {
				ag.RegisterTool(t)
			}
		}
	}

	// Session Loading
	if *sessionPath != "" {
		if _, err := os.Stat(*sessionPath); err == nil {
			if err := ag.LoadSession(*sessionPath); err != nil {
				fmt.Printf("Warning: Failed to load session: %v\n", err)
			}
		}
	}

	// Mode Selection
	if *investigate {
		args := flag.Args()
		if len(args) == 0 {
			fmt.Println("Usage: castor -investigate <goal>")
			os.Exit(1)
		}
		goal := strings.Join(args, " ")
		inv := &agent.Investigator{Agent: ag}
		fmt.Printf("🔍 Investigating: %s\n", goal)
		
		report, err := inv.Investigate(ctx, goal)
		if err != nil {
			fmt.Printf("Investigation failed: %v\n", err)
			os.Exit(1)
		}
		
		jsonReport, _ := json.MarshalIndent(report, "", "  ")
		fmt.Println(string(jsonReport))
		return
	}

	if *gui {
		if err := tui.Run(ag); err != nil {
			fmt.Printf("Error running TUI: %v\n", err)
			os.Exit(1)
		}
	} else if *interactive {
		runInteractive(ctx, ag, *sessionPath)
	} else {
		args := flag.Args()
		if len(args) == 0 {
			fmt.Println("Usage: castor [flags] <prompt>")
			os.Exit(1)
		}
		prompt := strings.Join(args, " ")
		runOnce(ctx, ag, prompt, *sessionPath)
	}
}

func runOnce(ctx context.Context, ag *agent.Agent, prompt string, sessionPath string) {
	stream, err := ag.Chat(ctx, prompt)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	for event := range stream {
		if event.Error != nil {
			fmt.Printf("\nError during generation: %v\n", event.Error)
			return
		}
		if event.Delta != "" {
			fmt.Print(event.Delta)
		}
		if len(event.ToolCalls) > 0 {
			for _, tc := range event.ToolCalls {
				fmt.Printf("\n[Tool Call: %s(%v)]\n", tc.Name, tc.Args)
			}
		}
	}
	fmt.Println()

	if sessionPath != "" {
		if err := ag.SaveSession(sessionPath); err != nil {
			fmt.Printf("Error saving session: %v\n", err)
		}
	}
}

func runInteractive(ctx context.Context, ag *agent.Agent, sessionPath string) {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Castor Interactive Mode (Ctrl+C to exit)")
	fmt.Println("----------------------------------------")

	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}
		input := scanner.Text()
		if input == "" {
			continue
		}

		stream, err := ag.Chat(ctx, input)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}

		for event := range stream {
			if event.Error != nil {
				fmt.Printf("\nError: %v\n", event.Error)
				break
			}
			if event.Delta != "" {
				fmt.Print(event.Delta)
			}
			if len(event.ToolCalls) > 0 {
				for _, tc := range event.ToolCalls {
					fmt.Printf("\n[Tool Call: %s(%v)]\n", tc.Name, tc.Args)
				}
			}
		}
		fmt.Println()

		if sessionPath != "" {
			if err := ag.SaveSession(sessionPath); err != nil {
				fmt.Printf("Error saving session: %v\n", err)
			}
		}
	}
}