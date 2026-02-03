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
    *   **Investigator:** Specialized sub-agent loop for deep codebase research.

## Installation

### Prerequisites
*   Go 1.21 or higher
*   An OpenAI-compatible API backend (Cloud or Local)

### Build
```bash
git clone https://github.com/techmuch/castor.git
cd castor
go build -o castor ./cmd/castor
```

## Connecting to an LLM

Castor is designed to be model-agnostic. You can connect it to any provider that supports the OpenAI Chat Completions API.

### 1. Using OpenAI
Set your API key as an environment variable:
```bash
export OPENAI_API_KEY=sk-...
./castor -model gpt-4o -tui
```

### 2. Using Ollama (Local)
Ollama provides an OpenAI-compatible endpoint.
1.  Start Ollama: `ollama serve`
2.  Pull a model: `ollama pull llama3`
3.  Run Castor:
```bash
# Point to your local Ollama instance (typically port 11434)
export OPENAI_API_KEY=ollama 
./castor -url http://localhost:11434/v1 -model llama3 -tui
```

### 3. Using Llama.cpp / vLLM
Start your server with the OpenAI-compatible flag and point Castor to it:
```bash
export OPENAI_API_KEY=local
./castor -url http://localhost:8080/v1 -model your-model -tui
```

## Usage Examples

### 1. Interactive Terminal UI (Recommended)
```bash
./castor -tui
```

### 2. Headless / One-Shot Mode
```bash
./castor "Summarize the files in the current directory"
```

### 3. Investigator Mode
Run a specialized research loop with a structured report output.
```bash
./castor -investigate "Find the logic responsible for tool execution"
```

### 4. Session Persistence
```bash
# Start and save a session
./castor -session session.json "Hello, remember this secret code: 12345"

# Resume later
./castor -session session.json "What was the secret code?"
```

## Development

See [DEVELOPMENT.md](DEVELOPMENT.md) for contribution guidelines.

## License

MIT