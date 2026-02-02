# Development Guide

Welcome to the Castor development team! This guide will help you get set up and understand the project structure.

## Prerequisites

*   **Go:** Version 1.21 or higher.
*   **Git:** For version control.
*   **Make:** (Optional) For running build scripts.
*   **Local LLM Host:** (Optional but recommended) [Ollama](https://ollama.com/) or similar for testing local inference.

## Project Structure

```text
castor/
├── cmd/                # Entry points for the application (CLI)
├── pkg/
│   ├── agent/          # Core agent orchestration logic
│   ├── llm/            # LLM provider interfaces and implementations
│   ├── tools/          # Built-in tool implementations (fs, edit, etc.)
│   ├── mcp/            # Model Context Protocol client implementation
│   └── tui/            # Terminal User Interface components
├── requirements.md     # Functional requirements
└── implementation_plan.md # Roadmap and task tracking
```

## Setup Instructions

1.  **Clone the repository:**
    ```bash
    git clone https://github.com/techmuch/castor.git
    cd castor
    ```

2.  **Install dependencies:**
    ```bash
    go mod download
    ```

3.  **Run Tests:**
    ```bash
    go test ./...
    ```

## Coding Standards

*   **Style:** Follow standard Go idioms (Effective Go). Run `go fmt` before committing.
*   **Testing:** New features must include unit tests. Use table-driven tests where appropriate.
*   **Comments:** Document exported functions and types. Focus on *why*, not just *what*.

## Contribution Workflow

1.  Pick a task from `implementation_plan.md`.
2.  Create a feature branch.
3.  Implement the feature and add tests.
4.  Run linters (if configured) and tests.
5.  Submit a Pull Request.

## Key Concepts for Developers

*   **LLMProvider:** This interface abstracts the AI model. We treat local and remote models identically via this interface.
*   **Tooling:** Tools are more than just functions; they have schemas and safety checks. The `Edit` tool is complex—study the "Read-Modify-Write" safety requirements carefully.
*   **Context:** We manage conversation history manually to support "sliding windows" for local models with smaller context limits.
