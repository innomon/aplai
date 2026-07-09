# Product Guidelines - APLAI

This document outlines the user experience (UX), error handling, API design, and architectural guidelines that must be adhered to throughout the development of APLAI.

---

## 1. User Experience & CLI Design

### 1.1 Output Formatting
- **Interactive REPL Mode (`--chat`):** Output must be clean and user-friendly. Use subtle colorization (e.g., ANSI escape codes for prompt prefix, user vs. assistant names) to distinguish between turns. Provide a real-time progress indicator or spinner while the model is compiling the graph or preparing the response.
- **Pipelined / Scripting Mode:** When input is piped via stdin or the tool is run in non-interactive mode, stdout must be strictly plain text. Color codes and progress indicators must be suppressed to ensure output can be easily parsed or redirected.

### 1.2 Error Handling & Verbosity
- Direct execution errors must be printed to `os.Stderr` as short, clear, human-friendly summaries.
- Avoid printing raw Go stack traces to the terminal during normal execution.
- Record the full stack trace and verbose error context inside the JSONL log file for developer debugging.

---

## 2. API Design & Integration

### 2.1 OpenAI Compatibility
- All HTTP server endpoints must closely match the OpenAI API specification to ensure seamless drop-in compatibility for local development.
- The `/v1/chat/completions` endpoint must return responses conforming to the OpenAI Chat Completion JSON structure.

### 2.2 API Error Format
- HTTP API errors must return the standard OpenAI error JSON payload format:
  ```json
  {
    "error": {
      "message": "Specific error details...",
      "type": "invalid_request_error",
      "code": "context_length_exceeded"
    }
  }
  ```
- Use appropriate HTTP status codes (e.g., `400 Bad Request`, `404 Not Found`, `500 Internal Server Error`).

---

## 3. Architecture & Coding Standards

### 3.1 Dependency Minimization
- Enforce a strict standard-library preference. Do not import third-party packages unless absolutely necessary.
- Accepted third-party dependencies are limited to core ML/Hardware acceleration libraries:
  - `github.com/gomlx/gomlx` (ML Engine)
  - `github.com/gomlx/go-darwinml` (Backend)
  - `github.com/daulet/tokenizers` (Tokenizer Wrapper)
- Do NOT use CLI frameworks (like Cobra/pflag), YAML parsers (unless standard or minimal), or logging frameworks (use standard `log/slog`).

### 3.2 Performance & Resource Management
- Optimize code for high performance and low garbage collector overhead on local workstations.
- Reuse buffers for tokenization, generation, and log formatting. Favour standard Go optimization patterns (like `sync.Pool`) for frequently allocated structures.
- Implement thread-safe synchronization across the log rotator and concurrent API request queues.

### 3.3 Separation of Concerns
- Maintain distinct module boundaries with zero circular dependencies:
  - `engine`: Pure tensor ML operations, KV cache management, and hardware backend.
  - `cli`: Argument parsing, REPL session lifecycle, and UNIX piping.
  - `server`: HTTP server configuration, request queueing, and route handlers.
  - `logger`: Logging rotator, structured slog multiplexer, and caller preservation wrapper.
