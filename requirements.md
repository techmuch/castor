# Castor: Local-First Agentic Framework Requirements

## 1. Core Principles
*   **Local-First:** Primary support for local inference (Ollama, vLLM) and local tool execution.
*   **Model Agnostic:** Decoupled LLM provider interface supporting OpenAI-compatible APIs and custom implementations.
*   **Agentic:** Built-in support for multi-turn reasoning, tool execution loops, and specialized sub-agents.
*   **Embeddable:** Designed as a Go module to be integrated into larger applications (CLIs, Servers, GUIs).
*   **Extensible:** Native support for the Model Context Protocol (MCP).

## 2. Operational Modes
### 2.1 Headless Scripting
*   Non-interactive execution for automation.
*   Support for session context persistence (load/save state).
*   Scoped tooling for file I/O and shell execution.

### 2.2 TUI Interactive
*   Rich terminal user interface (markdown rendering, syntax highlighting).
*   Streaming responses.
*   Interactive tool confirmation mechanisms.

### 2.3 Application API (Embedding)
*   Go interface for host application integration.
*   Support for "Host Tools" (querying app state).
*   Targeted context injection.

## 3. Architectural Components
### 3.1 LLM Provider Interface
*   Standard interface for `GenerateContent` (Chat) and `EmbedContent` (RAG).
*   Default implementation for OpenAI-compatible endpoints.
*   Support for structured output (JSON Schema).

### 3.2 Tooling System
*   **Registry:** Mechanism to register and retrieve tools based on operational profiles.
*   **Schema:** Auto-generation of JSON schemas for LLM consumption.
*   **Robust Edit Tool:**
    *   Read-Modify-Write safety (hashing).
    *   Multi-strategy matching (Exact, Flexible/Whitespace, Regex).
    *   Self-correction loop using a "fixer" LLM.

### 3.3 Agent Orchestrator
*   Manages the "Think-Act" loop.
*   Handles history and context window management (pruning/summarization).
*   Middleware support for human-in-the-loop confirmation.

### 3.4 Model Context Protocol (MCP) Client
*   Transports: Stdio and SSE.
*   Discovery: List tools/resources from MCP servers.
*   Execution: Proxy calls to MCP servers.

## 4. Specialized Agents
*   **Codebase Investigator:**
    *   Restricted read-only toolset.
    *   Structured "Scratchpad" workflow.
    *   JSON report output.

## 5. Security & Sandbox
*   Filesystem operations restricted to a defined workspace root.
*   Explicit confirmation flow for side-effect tools.
