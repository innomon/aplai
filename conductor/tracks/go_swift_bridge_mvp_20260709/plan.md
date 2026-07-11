# Implementation Plan - Implement Go-Swift Cgo Bridge and Core LLM Runner MVP

## Phase 1: Structured Logger & Rotator [checkpoint: 5a7b2be]

- [x] Task: Create thread-safe LogRotator and slog MultiHandler [1b68b32]
    - [x] Write unit tests for LogRotator and MultiHandler
    - [x] Implement LogRotator with file size-based rotation and backup limit pruning
    - [x] Implement MultiHandler multiplexing console TextHandler and JSONHandler
- [x] Task: Implement caller-preserving slog wrapper [fa769af]
    - [x] Write unit tests verifying caller location mapping
    - [x] Implement custom logger with program counter traversal bypassing wrapper methods
- [x] Task: Conductor - User Manual Verification 'Phase 1: Structured Logger & Rotator' (Protocol in workflow.md) [5a7b2be]

## Phase 2: Swift Bridge & Cgo Integration [checkpoint: 9fa913b]

- [x] Task: Implement Swift Bridge (`apple_ai_bridge.swift`) [672cb30]
    - [x] Write basic verification test or execution script for Swift compilation
    - [x] Implement C-compatible exported functions utilizing FoundationModels
    - [x] Use Task and DispatchSemaphore to block Cgo thread synchronously
    - [x] Implement strdup memory allocation and FreeAppleAIString deallocator
- [x] Task: Implement Go Cgo Wrapper [fbecf3f]
    - [x] Write Cgo binding tests mocking raw C calls
    - [x] Implement Go wrapper package referencing the C library with proper defer cleanup
- [x] Task: Build Automation [f8a33c5]
    - [x] Create Makefile to build libapple_ai_bridge.dylib and compile Go application
- [x] Task: Conductor - User Manual Verification 'Phase 2: Swift Bridge & Cgo Integration' (Protocol in workflow.md) [9fa913b]

## Phase 3: CLI, REPL & Configuration [checkpoint: 7781fd8]

- [x] Task: Configuration management (`config.yaml`) [b25acb8]
    - [x] Write unit tests for config file detection and loading
    - [x] Implement config loading checking custom path, current directory, and executable location
- [x] Task: Command line parser & pipeline execution [1a47cc2]
    - [x] Write unit tests for flag parsing and stdin piping
    - [x] Implement standard flag parser and piping loop
- [x] Task: Interactive REPL chat (--chat) [ae93dfe]
    - [x] Write unit tests/mocks for the REPL console interface
    - [x] Implement REPL loop with prompt history
- [x] Task: Conductor - User Manual Verification 'Phase 3: CLI, REPL & Configuration' (Protocol in workflow.md) [7781fd8]

## Phase 4: OpenAI HTTP Server & Concurrency [checkpoint: 455e2c7]

- [x] Task: OpenAI-compatible completion server [9d31531]
    - [x] Write integration tests for completions endpoint
    - [x] Implement net/http router and /v1/chat/completions with SSE streaming
- [x] Task: Concurrency queueing [9d31531]
    - [x] Write concurrency tests simulating multiple request arrivals
    - [x] Implement request queueing to prevent engine hardware overload
- [x] Task: Conductor - User Manual Verification 'Phase 4: OpenAI HTTP Server & Concurrency' (Protocol in workflow.md) [455e2c7]
