# Tech Stack - APLAI

This document defines the official programming language, libraries, frameworks, and tools used in the development of APLAI.

---

## 1. Core Languages & Environment
- **Go (Golang):** Used for the main application CLI, the HTTP server, argument parsing, logging, and concurrency control.
- **Swift:** Used to write the direct native macOS bridge interfacing with Apple's `FoundationModels` framework.
- **Cgo:** Serves as the in-process bridge between Go and Swift using C-compatible exports (`@_cdecl`).

---

## 2. Machine Learning Engine
- **Core Engine:** macOS `FoundationModels` framework (Sequoia/Tahoe 15+).
- **Session API:** `LanguageModelSession` for native, hardware-accelerated, on-device LLM inference using Unified Memory and Apple Silicon NPUs/GPUs.

---

## 3. Web & Network Layer
- **HTTP Server Framework:** Standard Library `net/http`
  - Zero external dependencies.
  - Exposes the `/v1/chat/completions` endpoint.
  - Native support for Server-Sent Events (SSE) streaming.

---

## 4. System & Build Utilities
- **Build Automation:** `Makefile` orchestrating:
  - Compilation of Swift dynamic library via `swiftc -emit-library`.
  - Linking of macOS frameworks (`Foundation`, `FoundationModels`).
  - Compilation of the Go binary with dynamic library linking.
- **CLI Parser:** Go Standard Library `flag` package (no third-party frameworks like Cobra).
- **Logging & Rotation:** Go Standard Library `log/slog` with a custom thread-safe `LogRotator`.
