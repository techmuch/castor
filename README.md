# Castor

Castor is a local-first, model-agnostic AI agent framework written in Go. It is designed to be the intelligence engine for your applications, capable of running headless scripts, powering interactive terminal UIs, or embedding directly into larger Go software.

## Vision

Unlike many agent frameworks that rely heavily on cloud APIs, Castor treats local LLMs (via Ollama, vLLM, etc.) as first-class citizens. It provides a robust, type-safe foundation for building agentic workflows that can run entirely on a user's machine or scale to cloud providers when needed.

## Features

*   **🔌 Model Agnostic:** Plug-and-play support for any OpenAI-compatible API (Ollama, Llama.cpp) and extensible interfaces for others.
*   **🛠️ Robust Tooling:** Includes a sophisticated file editing tool with self-correction capabilities and support for the Model Context Protocol (MCP).
*   **🧠 Context Management:** Smart context window handling, history persistence, and session management.
*   **📦 Embeddable:** Use it as a library in your own Go tools, or run it as a standalone CLI.
*   **🛡️ Secure:** Workspace sandboxing and human-in-the-loop confirmation mechanisms.

## Operational Modes

1.  **Headless:** For automation scripts and shell integration.
2.  **TUI:** A rich terminal user interface for interactive coding assistance.
3.  **Embedded:** An API mode for integrating agentic capabilities into host applications.

## Getting Started

*(Instructions coming soon as the project initializes)*

## License

[License Information Here]
