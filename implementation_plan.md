# Implementation Plan

## Phase 1: Foundation (Completed)
- [x] **Define Core Interfaces**
    - [x] Create `pkg/llm/provider.go`: Define `LLMProvider` interface.
    - [x] Create `pkg/llm/types.go`: Define `Message`, `Part`, `Tool` types.
- [x] **Implement OpenAI/Ollama Provider**
    - [x] Create `pkg/llm/openai/client.go`: Basic chat completion implementation.
    - [x] Add support for `GenerateContent` streaming.
    - [x] Add support for `tool_calls` parsing and serialization.
- [x] **Basic Agent Loop**
    - [x] Create `pkg/agent/orchestrator.go`: Implement the Think-Act loop.
    - [x] Implement simple history management.
    - [x] Implement tool execution loop.

## Phase 2: Tooling & Configuration (In Progress)
- [x] **Tool Registry**
    - [x] Create `pkg/tools/registry.go`: Registry with profile support.
- [x] **Core Filesystem Tools**
    - [x] Implement `read_file`, `ls` (`pkg/tools/fs`).
    - [x] Implement workspace sandboxing.
- [x] **Advanced Edit Tool** (`pkg/tools/edit`)
    - [x] Implement Exact Match strategy.
    - [x] Implement hashing for Read-Modify-Write safety.
    - [x] Implement Flexible/Regex matching.
    - [x] Implement "Fixer" loop for failed edits.

## Phase 3: Modes & Use Cases (Completed)
- [x] **Headless CLI**
    - [x] Implement `cmd/castor/main.go` for flags and entry point.
    - [x] Add context persistence (save/load session).
- [x] **TUI (Terminal UI)**
    - [x] Integrate Bubble Tea.
    - [x] Implement chat view and input handling.
- [ ] **API Interface**
    - [ ] Define exported `Agent` struct for library usage (Refine `pkg/agent`).

## Phase 4: MCP & Advanced Agents (In Progress)
- [x] **MCP Client**
    - [x] Implement Stdio transport.
    - [x] Implement Tool discovery and registration.
- [ ] **Investigator Agent**
    - [ ] Create specialized loop for "Scratchpad" workflow.

## Continuous Tasks
- [ ] Add Unit Tests for each component.
- [x] Update `DEVELOPMENT.md` and `README.md`.