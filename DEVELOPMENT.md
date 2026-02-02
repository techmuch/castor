# Development Guide

Welcome to the Castor development team! This guide will help you get set up and understand the project structure.

## Prerequisites

*   **Go:** Version 1.21 or higher.
*   **Git:** For version control.
*   **Make:** For running build automation commands.
*   **Local LLM Host:** (Optional but recommended) [Ollama](https://ollama.com/) or similar for testing local inference without API costs.

## Project Structure

```text
castor/
├── cmd/
│   └── castor/         # Main CLI entry point
├── pkg/
│   ├── agent/          # Core agent orchestration & session management
│   ├── llm/            # LLM provider interfaces & OpenAI client
│   ├── tools/          # Tool implementations
│   │   ├── fs/         # Filesystem tools (ls, read_file)
│   │   ├── edit/       # Edit tool (replace) with robust matching
│   │   └── registry.go # Tool registry
│   └── tui/            # Bubble Tea terminal UI
├── requirements.md     # Functional requirements
└── implementation_plan.md # Roadmap and task tracking
```

## Getting Started

1.  **Clone the repository:**
    ```bash
    git clone https://github.com/techmuch/castor.git
    cd castor
    ```

2.  **Install dependencies:**
    ```bash
    go mod download
    ```

3.  **Build:**
    ```bash
    make build
    ```

4.  **Run Tests:**
    ```bash
    make test
    ```

## Common Commands

*   `make run`: Run the CLI in interactive mode.
*   `make tui`: Run the CLI in TUI mode.
*   `make clean`: Remove build artifacts.

## Coding Standards

*   **Interfaces:** We define interfaces where they are used (e.g., `pkg/agent/tool.go`), but `LLMProvider` is defined in `pkg/llm` as a common contract.
*   **Error Handling:** Wrap errors with context: `fmt.Errorf("failed to load session: %w", err)`.
*   **Tooling:** New tools must implement the `agent.Tool` interface and provide a JSON schema. Ensure strict input validation and sandboxing for filesystem tools.

## Contribution Workflow

1.  Pick a task from `implementation_plan.md`.
2.  Create a feature branch.
3.  Implement the feature and add tests.
4.  Run `go fmt ./...` and `go vet ./...`.
5.  Submit a Pull Request.