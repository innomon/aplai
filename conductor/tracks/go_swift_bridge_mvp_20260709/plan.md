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

## Phase 2: Swift Bridge & Cgo Integration

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
- [ ] Task: Conductor - User Manual Verification 'Phase 2: Swift Bridge & Cgo Integration' (Protocol in workflow.md)

## Phase 3: CLI, REPL & Configuration

- [ ] Task: Configuration management (`config.yaml`)
    - [ ] Write unit tests for config file detection and loading
    - [ ] Implement config loading checking custom path, current directory, and executable location
- [ ] Task: Command line parser & pipeline execution
    - [ ] Write unit tests for flag parsing and stdin piping
    - [ ] Implement standard flag parser and piping loop
- [ ] Task: Interactive REPL chat (--chat)
    - [ ] Write unit tests/mocks for the REPL console interface
    - [ ] Implement REPL loop with prompt history
- [ ] Task: Conductor - User Manual Verification 'Phase 3: CLI, REPL & Configuration' (Protocol in workflow.md)

## Phase 4: OpenAI HTTP Server & Concurrency

- [ ] Task: OpenAI-compatible completion server
    - [ ] Write integration tests for completions endpoint
    - [ ] Implement net/http router and /v1/chat/completions with SSE streaming
- [ ] Task: Concurrency queueing
    - [ ] Write concurrency tests simulating multiple request arrivals
    - [ ] Implement request queueing to prevent engine hardware overload
- [ ] Task: Conductor - User Manual Verification 'Phase 4: OpenAI HTTP Server & Concurrency' (Protocol in workflow.md)
