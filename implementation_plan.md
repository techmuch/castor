# Implementation Plan

## Phase 1: Foundation (Current Focus)
- [ ] **Define Core Interfaces**
    - [ ] Create `pkg/llm/provider.go`: Define `LLMProvider` interface.
    - [ ] Create `pkg/agent/types.go`: Define `Message`, `Part`, `Tool` types.
- [ ] **Implement OpenAI/Ollama Provider**
    - [ ] Create `pkg/llm/openai/client.go`: Basic chat completion implementation.
    - [ ] Add support for `GenerateContent` streaming.
- [ ] **Basic Agent Loop**
    - [ ] Create `pkg/agent/orchestrator.go`: Implement the Think-Act loop.
    - [ ] Implement simple history management (append only for now).

## Phase 2: Tooling & Configuration
- [ ] **Tool Registry**
    - [ ] Create `pkg/tools/registry.go`: Registry with profile support.
- [ ] **Core Filesystem Tools**
    - [ ] Implement `read_file`, `ls`, `glob` with workspace sandboxing.
- [ ] **Advanced Edit Tool**
    - [ ] Implement hashing for Read-Modify-Write safety.
    - [ ] Implement Exact Match strategy.
    - [ ] Implement Flexible/Regex matching.
    - [ ] Implement "Fixer" loop for failed edits.

## Phase 3: Modes & Use Cases
- [ ] **Headless CLI**
    - [ ] Implement `cmd/castor/main.go` for flags and entry point.
    - [ ] Add context persistence (save/load session).
- [ ] **TUI (Terminal UI)**
    - [ ] Integrate Bubble Tea.
    - [ ] Implement chat view and input handling.
- [ ] **API Interface**
    - [ ] Define exported `Agent` struct for library usage.

## Phase 4: MCP & Advanced Agents
- [ ] **MCP Client**
    - [ ] Implement Stdio transport.
    - [ ] Implement Tool discovery and registration.
- [ ] **Investigator Agent**
    - [ ] Create specialized loop for "Scratchpad" workflow.

## Continuous Tasks
- [ ] Add Unit Tests for each component.
- [ ] Update `DEVELOPMENT.md` with new findings.
