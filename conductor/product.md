# Initial Concept
APLAI: A Go-native local LLM runner and server optimized for Apple Silicon (based on aplai-prompt.md)

---

# Product Guide - APLAI

## 1. Product Vision & Overview
APLAI is a high-performance, 100% on-device local LLM (Large Language Model) runner and server optimized specifically for Apple Silicon. It is implemented as a Go application that leverages a direct, in-process Cgo-to-Swift bridge to natively execute on-device inference using Apple's `FoundationModels` framework and `LanguageModelSession` (available in macOS Sequoia/Tahoe 15+).

APLAI serves two main purposes:
1. A versatile, UNIX-style CLI for prompt execution, interactive REPL chats, and pipeline processing.
2. A lightweight, OpenAI-compatible HTTP server to easily integrate local LLMs into developer workflows.

---

## 2. Target Audience
- **CLI Power Users:** Developers and system administrators who want a fast, scriptable, local LLM runner that supports standard UNIX input piping.
- **Application Developers:** Software engineers looking for an on-device, OpenAI-compatible API server to integrate LLMs into local applications without external API dependencies.

---

## 3. Key Functional Features

### 3.1 Cgo-to-Swift Bridge
- **In-process Native Inference:** Direct Cgo calls to a compiled Swift dynamic library (`libapple_ai_bridge.dylib`).
- **macOS FoundationModels integration:** Interfacing with native Apple Silicon NPUs and GPUs via `LanguageModelSession`.
- **Thread Synchronization:** Wrap asynchronous Swift calls in a task blocked by `DispatchSemaphore` to keep the Cgo calling thread synchronous.
- **Memory Safety:** Heap allocation of outputs via `strdup` on the Swift side, with explicit deallocation in Go using a custom cleanup function.

### 3.2 CLI Interface & Configuration
- **UNIX CLI Modes:** Supports direct prompt invocation, interactive REPL chat, and stdin pipe processing.
- **Configuration Management (`config.yaml`):**
  - Configured via a `config.yaml` file with sensible defaults.
  - Search order:
    1. Explicitly passed configuration path flag.
    2. Current working directory (`./config.yaml`).
    3. Executable's directory (directory containing the `aplai` binary).
- **Argument Parsing:** Implemented strictly using the Go standard library `flag` package (no third-party CLI frameworks like Cobra or pflag).

### 3.3 OpenAI-Compatible Local HTTP Server
- **Endpoint:** Exposes a `/v1/chat/completions` endpoint compatible with the OpenAI specification.
- **SSE Streaming:** Supports Server-Sent Events (SSE) for real-time streaming of token responses.
- **Concurrency & Queueing:** Robust handling of multiple concurrent client requests with queueing to prevent hardware overload.

### 3.4 Logging, Monitoring & Diagnostics
- **Dual Console & JSONL Logging:** Simultaneous output of human-readable text to `os.Stderr`/console and machine-readable JSONL to a log file.
- **Log Rotation:** Custom, thread-safe size-based log file rotation and backup pruning (no external dependencies).
- **Caller Preservation:** Preserves actual caller source file and line number in structured logs across logger wrappers using runtime program counter tracing.
- **Performance Metrics:** Real-time generation statistics in logs, including token generation speed (tokens/sec) and time-to-first-token (TTFT).
