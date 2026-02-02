package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/techmuch/castor/pkg/agent"
	"github.com/techmuch/castor/pkg/llm/openai"
	"github.com/techmuch/castor/pkg/tools/edit"
	"github.com/techmuch/castor/pkg/tools/fs"
)

func main() {
	apiKey := os.Getenv("OPENAI_API_KEY")
	model := flag.String("model", "gpt-3.5-turbo", "LLM model to use")
	systemPrompt := flag.String("system", "You are a helpful assistant with access to files.", "System prompt")
	interactive := flag.Bool("i", false, "Interactive mode")
	workspace := flag.String("w", ".", "Workspace root directory")
	flag.Parse()

	if apiKey == "" {
		fmt.Println("Error: OPENAI_API_KEY environment variable is required.")
		os.Exit(1)
	}

	client := openai.NewClient("", apiKey, *model)
	ag := agent.New(client, *systemPrompt)
	
	// Register Tools
	lsTool := &fs.ListDirTool{WorkspaceRoot: *workspace}
	readTool := &fs.ReadFileTool{WorkspaceRoot: *workspace}
	replaceTool := &edit.EditTool{WorkspaceRoot: *workspace}

	ag.RegisterTool(lsTool)
	ag.RegisterTool(readTool)
	ag.RegisterTool(replaceTool)

	ctx := context.Background()

	if *interactive {
		runInteractive(ctx, ag)
	} else {
		// One-shot mode
		args := flag.Args()
		if len(args) == 0 {
			fmt.Println("Usage: castor [flags] <prompt>")
			os.Exit(1)
		}
		prompt := strings.Join(args, " ")
		runOnce(ctx, ag, prompt)
	}
}

func runOnce(ctx context.Context, ag *agent.Agent, prompt string) {
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
		// Print text delta
		if event.Delta != "" {
			fmt.Print(event.Delta)
		}
		// Print tool calls (optional feedback)
		if len(event.ToolCalls) > 0 {
			for _, tc := range event.ToolCalls {
				fmt.Printf("\n[Tool Call: %s(%v)]\n", tc.Name, tc.Args)
			}
		}
	}
	fmt.Println() 
}

func runInteractive(ctx context.Context, ag *agent.Agent) {
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
	}
}