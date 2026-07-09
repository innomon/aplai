# Specification: Implement Go-Swift Cgo Bridge and Core LLM Runner MVP

## 1. Overview
This track implements the MVP (Minimum Viable Product) of APLAI, an on-device local LLM runner and server optimized for Apple Silicon. Instead of using Python or external runners, the core engine leverages macOS Sequoia/Tahoe (15+) native `FoundationModels` and `LanguageModelSession` via an in-process Cgo-to-Swift bridge.

---

## 2. Functional Requirements

### 2.1 Swift Bridge (`apple_ai_bridge.swift`)
- Interface with the `FoundationModels` framework.
- Export C-compatible functions using the `@_cdecl` attribute.
- Implement thread-safe synchronization: wrap the async call to `LanguageModelSession.respond(to:)` inside a Swift `Task` and block the Cgo thread using `DispatchSemaphore`.
- Memory Safety: Allocate return strings using `strdup`. Export a companion cleanup function `FreeAppleAIString` to release memory.
- Error Isolation: Capture exceptions in `do-catch` blocks and return `"ERROR: <error_message>"`.

### 2.2 Go Wrapper & CLI (`main.go`, `internal/cli/`, `internal/engine/`)
- Link `Foundation`, `FoundationModels` and the bridge via Cgo.
- Provide idiomatic Go wrappers around C functions.
- Enforce strict `defer` memory cleanup.
- Config: Search for `config.yaml` in user-specified path, current directory, and binary directory.
- CLI: Parse options using standard library `flag` package. Support interactive `--chat` (REPL) and standard input piping.

### 2.3 HTTP Server (`internal/server/`)
- Expose an OpenAI-compatible `/v1/chat/completions` endpoint.
- Support SSE streaming.
- Implement concurrency control and request queueing.

### 2.4 Structured Logger & Rotator (`internal/logger/`)
- Implement a thread-safe custom `LogRotator` writing JSONL.
- Multiplex output to console and file via `MultiHandler`.
- Wrap `slog` to preserve the original program counter (caller file/line).

### 2.5 Build System (`Makefile`)
- Compile `apple_ai_bridge.swift` to `libapple_ai_bridge.dylib`.
- Build the Go application, setting correct runtime library paths.

---

## 3. Acceptance Criteria
- [ ] Swift bridge compiles successfully into a shared library.
- [ ] Go application links the shared library and executes local inference.
- [ ] Stdin piping prints plain text to stdout.
- [ ] `--chat` mode starts interactive REPL with history.
- [ ] HTTP server accepts OpenAI requests and streams tokens using SSE.
- [ ] Log messages are duplicated to console (text) and file (JSONL).
- [ ] Log rotation prunes files when exceeding max size/count.
- [ ] Logger logs contain correct caller source file and line numbers.
