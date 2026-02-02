# Castor

Castor is a local-first, model-agnostic AI agent framework written in Go. It is designed to be the intelligence engine for your applications, capable of running headless scripts, powering interactive terminal UIs, or embedding directly into larger Go software.

## Vision

Unlike many agent frameworks that rely heavily on cloud APIs, Castor treats local LLMs (via Ollama, vLLM, etc.) as first-class citizens. It provides a robust, type-safe foundation for building agentic workflows that can run entirely on a user's machine or scale to cloud providers when needed.

## Features

*   **🔌 Model Agnostic:** Plug-and-play support for OpenAI-compatible APIs (Ollama, Llama.cpp, OpenAI).
*   **🛠️ Robust Tooling:**
    *   **Filesystem:** Safely list and read files within a sandboxed workspace.
    *   **Smart Edit:** A robust `replace` tool with exact matching, whitespace-insensitive flexible matching, and hash-based verification for safety.
*   **🧠 Context Management:**
    *   **Session Persistence:** Save and load chat history to JSON files to resume conversations later.
    *   **History Management:** Type-safe message history handling.
*   **🖥️ Operational Modes:**
    *   **Headless CLI:** Scriptable interface for automation.
    *   **TUI:** Rich interactive Terminal User Interface powered by [Bubble Tea](https://github.com/charmbracelet/bubbletea).
    *   **REPL:** Simple interactive command-line loop.

## Installation

### Prerequisites
*   Go 1.21 or higher
*   An OpenAI-compatible API key (or a local runner like Ollama)

### Build
```bash
git clone https://github.com/techmuch/castor.git
cd castor
go build -o castor ./cmd/castor
```

## Usage

Set your API key (if using OpenAI):
```bash
export OPENAI_API_KEY=sk-...
```

### 1. Interactive Terminal UI (Recommended)
The TUI provides a modern chat experience with scrolling history and text editing.
```bash
./castor -tui
```

### 2. Headless / One-Shot Mode
Execute a single prompt and exit. Useful for scripting.
```bash
./castor "List the files in the current directory"
```

### 3. Session Persistence
Save the conversation state to a file to resume it later.
```bash
# Start a session
./castor -session my_chat.json -i

# The agent remembers previous context when you run it again with the same session file
```

### 4. Workspace Sandboxing
By default, Castor uses the current directory as the workspace. You can specify a different root to restrict file access:
```bash
./castor -w /path/to/safe/workspace -tui
```

## Development

See [DEVELOPMENT.md](DEVELOPMENT.md) for contribution guidelines.

## License

MIT