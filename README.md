# APLAI - Apple Silicon Local AI Runner MVP

APLAI is a high-performance, developer-friendly local AI inference engine designed for macOS Sequoia and Tahoe. It compiles native Swift bindings to Apple's on-device `LanguageModelSession` (FoundationModels) and exposes them through a Go-based UNIX-style CLI, interactive REPL console, and an OpenAI-compatible HTTP completions server.

## Features

- **Native Apple Silicon Inference:** Connects directly to macOS `LanguageModelSession` via an in-process Swift-to-C dynamic bridge.
- **UNIX Pipeline Support:** Stream prompts into `aplai` via stdin or trailing arguments, returning plain text suited for scripting and terminal pipes.
- **Colorized Interactive REPL (`--chat`):** A premium console chat interface with ANSI colors and real-time streaming token deltas.
- **OpenAI-Compatible HTTP Server (`--server`):** Implements `/v1/chat/completions` supporting both JSON responses and Server-Sent Events (SSE) streaming.
- **Hardware-Safe Concurrency Queueing:** Serializes concurrent request arrivals using an execution mutex to prevent engine/hardware overload.
- **Structured Dual Logging:** Structured logging writing JSON logs to size-rotated, chronological backup files while printing clean messages to console.

---

## Getting Started

### Prerequisites

- **macOS Sequoia (15.0) or Tahoe (16.0)**
- **Xcode / Command Line Tools** with Swift compiler installed (`swiftc`)
- **Go 1.21+**

### Building the Project

The project is equipped with a `Makefile` that handles Cgo environment flags and dynamic loader configuration:

```bash
# Compile the Swift dynamic library and build the 'aplai' binary
make build

# Run the full Go test suite
make test

# Clean build artifacts
make clean
```

---

## Usage Modes

### 1. UNIX Pipeline & Single Prompts
Pass a prompt as a trailing positional argument:
```bash
CGO_LDFLAGS_ALLOW="-Wl,-rpath,@loader_path/.*" ./aplai "Explain Apple Silicon in one sentence."
```

Or pipe a prompt via standard input (`stdin`):
```bash
echo "List three colors." | CGO_LDFLAGS_ALLOW="-Wl,-rpath,@loader_path/.*" ./aplai
```

### 2. Interactive REPL Chat
Start an interactive chat session with ANSI colors and real-time streaming:
```bash
CGO_LDFLAGS_ALLOW="-Wl,-rpath,@loader_path/.*" ./aplai --chat
```
- Exit the chat using `/exit`, `/quit`, `exit`, or `quit`.

### 3. OpenAI-Compatible Web Server
Start the HTTP completions server:
```bash
CGO_LDFLAGS_ALLOW="-Wl,-rpath,@loader_path/.*" ./aplai --server
```
The server listens on `127.0.0.1:8080` by default.

#### Example API Requests:
**Synchronous JSON Request:**
```bash
curl -X POST http://127.0.0.1:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{"messages": [{"role": "user", "content": "What is 2+2?"}]}'
```

**Streaming SSE Request:**
```bash
curl -X POST http://127.0.0.1:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{"stream": true, "messages": [{"role": "user", "content": "What is 2+2?"}]}'
```

---

## Configuration (`config.yaml`)

APLAI searches for configuration parameters in this priority order:
1. Explicitly passed path via the `-config <path>` flag.
2. Current working directory (`./config.yaml`).
3. Binary's executable directory (`<exe_dir>/config.yaml`).

If no file is found, sensible defaults are applied.

### Configuration Template

```yaml
server:
  host: "127.0.0.1"
  port: 8080

logger:
  level: "INFO"
  console_enabled: true
  file_enabled: true
  dir: "logs"
  filename: "aplai.log"
  max_size_mb: 10
  max_backups: 5

engine:
  instructions: "You are a helpful on-device AI assistant."
```
