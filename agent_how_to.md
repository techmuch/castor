# Building a Local-First AI Agent Framework in Go

This document outlines the requirements and architectural blueprint for creating a Go-based library that replicates and extends the capabilities of the Gemini CLI. The primary goal is to provide a set of **agentic AI features** as a module that can be integrated into larger applications, while maintaining support for self-hosted, local Large Language Models (LLMs) and local tool execution.

## 1. Vision & Scope

The module is designed to be the intelligence engine for various application contexts. Unlike the Node.js reference implementation which targets the Gemini API, this library must be **model-agnostic**, treating local LLMs (via Ollama, vLLM, or local endpoints) as first-class citizens.

### Key Characteristics
- **Local-First:** All tools and model inference can run on the user's machine.
- **Model Agnostic:** Plug-and-play support for any OpenAI-compatible API or custom local model interfaces.
- **Agentic:** Supports multi-turn reasoning, tool use, and specialized sub-agents.
- **Extensible:** Full support for the Model Context Protocol (MCP) to extend capabilities.
- **Modular Integration:** Designed to be embedded into other Go applications with specific toolsets for different runtime modes.

## 2. Primary Use Cases

The module must support three distinct operational modes, each with granular control over enabled tools and capabilities.

### 2.1 Headless Scripting Mode
A non-interactive mode designed for automation and shell scripting, similar to `gemini -r`.
*   **Behavior:** Accepts a prompt and optional context flag, executes the task, and returns the result to `stdout`.
*   **Context Persistence:** Must be able to load previous conversation history and save the new state, enabling multi-step scripted workflows (e.g., `agent -context=session_id "analyze error.log"` followed by `agent -context=session_id "fix the error"`).
*   **Tooling:** Typically high-privilege tools (File I/O, Shell execution) but strictly scoped by the caller.

### 2.2 TUI Interactive Mode
A full terminal user interface experience, identical in behavior and polish to the Node.js `gemini-cli`.
*   **Behavior:** Interactive chat loop, rich text rendering (Markdown, syntax highlighting), streaming responses, and real-time tool status updates.
*   **Tooling:** Standard set of developer tools (Edit, LS, Grep) plus interactive confirmations.

### 2.3 Web/App API Mode
An embedding mode where the agent serves as a backend logic provider for a web-based or larger desktop application.
*   **Behavior:** Exposes an API (e.g., HTTP/REST or Go Interface) to the host application.
*   **App Integration:** Enables specific "Host Tools" that allow the LLM to query the state of the parent application (e.g., `get_app_config`, `query_database`).
*   **Targeted Context:** Allows the host app to inject specific context (e.g., current view data, user selection) into the agent's prompt without polluting the global history.

*Requirement:* Each use case must have an independent configuration profile. It should be possible to enable the "Edit" tool for the TUI but disable it for the Web API mode, or register a custom "DatabaseQuery" tool only for the Web API mode.

## 3. Core Architecture

The library should be structured around three primary interfaces: `LLMProvider`, `Tool`, and `Agent`.

### 3.1 LLM Provider Interface
To support self-hosted models, strictly decouple the agent logic from the model API.

```go
type Message struct {
    Role    string // "user", "model", "system"
    Content []Part // Text, Images, ToolCalls, ToolResponses
}

type GenerateOptions struct {
    Temperature float32
    TopP        float32
    StopTokens  []string
    JSONSchema  *Schema // For structured output/tool calling
}

type LLMProvider interface {
    // Standard chat completion
    GenerateContent(ctx context.Context, history []Message, opts GenerateOptions) (<-chan StreamEvent, error)
    
    // For embedding-based retrieval (RAG)
    EmbedContent(ctx context.Context, texts []string) ([][]float32, error)
}
```
*Requirement:* Provide a default implementation for OpenAI-compatible APIs (which covers Ollama, vLLM, Llama.cpp) and a standard interface to easily add others.

### 3.2 Tooling System
Tools are the hands of the agent. The system needs a robust registry and execution framework.

#### Tool Definition
Tools must define their schema (for the LLM) and their execution logic (for the runtime).
```go
type Tool interface {
    Name() string
    Description() string
    Schema() *JSONSchema // Schema for arguments
    Execute(ctx context.Context, args map[string]interface{}) (Result, error)
}
```

#### The "Edit" Tool (Critical Requirement)
The reference implementation's `EditTool` is highly sophisticated. The Go version must replicate its reliability features:
1.  **Read-Modify-Write Safety:** Hash file content before editing to ensure no external changes occurred.
2.  **Multi-Strategy Matching:**
    *   *Exact Match:* Replace literal string.
    *   *Flexible Match:* Ignore whitespace/indentation differences.
    *   *Regex Match:* Use pattern matching for structure.
3.  **Self-Correction:** If an edit fails (e.g., "string not found"), the tool should automatically query a "fixer" LLM with the error and file content to generate a corrected search string *before* returning an error to the user.

### 3.3 Agent Orchestrator
The main loop that ties everything together.

1.  **Initialization:** Load system prompt, tools (filtered by Use Case), and context.
2.  **Think-Act Loop:**
    *   Send history to LLM.
    *   Parse response (Text or Tool Call).
    *   *If Text:* Stream to user.
    *   *If Tool Call:* Validate arguments -> Execute Tool -> Add Result to History -> Repeat.
3.  **Context Management:** Maintain a sliding window of chat history. Prune old messages or summarize them if the context limit of the local model is reached.

## 4. Specialized Agents

The system should support "Specialized Agents" – transient agents spawned for specific tasks.

### 4.1 Codebase Investigator
Replicate the `codebase_investigator` pattern.
*   **Restricted Toolset:** Only read-only tools (`ls`, `grep`, `glob`, `read_file`).
*   **Structured Thinking:** Enforce a "Scratchpad" workflow where the model must maintain a checklist of what it has explored and what it needs to find.
*   **Output:** Returns a structured report (JSON) rather than conversational text.

## 5. Model Context Protocol (MCP) Support

To match the extensibility of the reference implementation, the library must implement an MCP Client.

*   **Transport:** Implement `Stdio` (for local processes) and `SSE` (for remote servers) transports.
*   **Discovery:**
    *   Connect to an MCP server.
    *   List available tools, resources, and prompts.
    *   Dynamically register these as `Tool` instances in the agent's registry.
*   **Execution:** Proxy tool calls from the LLM to the MCP server and return the results.

## 6. File System & Safety

*   **Sandboxing:** All file operations should be restricted to a specific root directory (the workspace).
*   **Confirmation:** Implement a middleware hooks system to allow for "Human-in-the-loop" confirmation before executing side-effect tools (like `edit` or `run_shell_command`).

## 7. Implementation Roadmap

1.  **Phase 1: Foundation**
    *   Define `LLMProvider` interface.
    *   Implement OpenAI/Ollama provider.
    *   Build basic `Chat` loop with history and context persistence.

2.  **Phase 2: Tooling & Configuration**
    *   Implement `ToolRegistry` with support for "Profiles" (enabled/disabled sets).
    *   Build core `fs` tools (`read_file`, `ls`, `glob`).
    *   Implement the robust `edit` tool with multi-strategy matching.

3.  **Phase 3: Modes & Use Cases**
    *   **Headless:** Implement the CLI flags and context storage logic for scripted execution.
    *   **TUI:** Port the interactive loop, integrating a Go TUI library (e.g., Bubble Tea).
    *   **API:** Define the Go interface/structs for embedding the agent into a larger Go binary.

4.  **Phase 4: MCP & Advanced Agents**
    *   Implement MCP Client (Client -> Transport -> Server).
    *   Implement the Investigator agent pattern.

This document serves as the master specification for the Go library.
