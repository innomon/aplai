package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Session represents the engine inference session interface.
type Session interface {
	Respond(prompt string) (string, error)
	RespondStream(prompt string, onToken func(string)) error
	Close() error
}

// Engine represents the inference engine creator interface.
type Engine interface {
	NewSession(instructions string) (Session, error)
}

// ChatMessage represents a message in standard chat requests.
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatCompletionRequest is the incoming payload for completions.
type ChatCompletionRequest struct {
	Model    string        `json:"model"`
	Messages []ChatMessage `json:"messages"`
	Stream   bool          `json:"stream"`
}

// ChatCompletionChoice is a single choice in synchronous responses.
type ChatCompletionChoice struct {
	Index        int         `json:"index"`
	Message      ChatMessage `json:"message"`
	FinishReason string      `json:"finish_reason"`
}

// ChatCompletionResponse is the synchronous completion body.
type ChatCompletionResponse struct {
	ID      string                 `json:"id"`
	Object  string                 `json:"object"`
	Created int64                  `json:"created"`
	Model   string                 `json:"model"`
	Choices []ChatCompletionChoice `json:"choices"`
}

// ChatCompletionChunkChoice is a single choice in streaming chunks.
type ChatCompletionChunkChoice struct {
	Index        int         `json:"index"`
	Delta        ChatMessage `json:"delta"`
	FinishReason string      `json:"finish_reason"`
}

// ChatCompletionChunk is a streaming chunk payload.
type ChatCompletionChunk struct {
	ID      string                      `json:"id"`
	Object  string                      `json:"object"`
	Created int64                       `json:"created"`
	Model   string                      `json:"model"`
	Choices []ChatCompletionChunkChoice `json:"choices"`
}

// Server handles APLAI HTTP requests.
type Server struct {
	host                string
	port                int
	defaultInstructions string
	engine              Engine
	router              *http.ServeMux
	mu                  sync.Mutex // Mutex serialization prevents engine hardware overload
}

// NewServer returns a new initialized Server.
func NewServer(host string, port int, defaultInstructions string, engine Engine) *Server {
	s := &Server{
		host:                host,
		port:                port,
		defaultInstructions: defaultInstructions,
		engine:              engine,
		router:              http.NewServeMux(),
	}
	s.router.HandleFunc("/v1/chat/completions", s.handleChatCompletions)
	return s
}

// Router returns the internal multiplexer.
func (s *Server) Router() http.Handler {
	return s.router
}

// Start runs the HTTP server and blocks until the context is cancelled.
func (s *Server) Start(ctx context.Context) error {
	addr := fmt.Sprintf("%s:%d", s.host, s.port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", addr, err)
	}

	// Update port if dynamic port 0 was passed
	s.port = listener.Addr().(*net.TCPAddr).Port
	slog.Info("Server listening", "address", listener.Addr().String())

	srv := &http.Server{
		Handler: s.router,
	}

	errChan := make(chan error, 1)
	go func() {
		errChan <- srv.Serve(listener)
	}()

	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		slog.Info("Shutting down HTTP server...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return srv.Shutdown(shutdownCtx)
	}
}

func (s *Server) handleChatCompletions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ChatCompletionRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Malformed JSON request body", http.StatusBadRequest)
		return
	}

	var systemInstructions string
	var userPrompts []string

	for _, msg := range req.Messages {
		if msg.Role == "system" {
			systemInstructions = msg.Content
		} else if msg.Role == "user" {
			userPrompts = append(userPrompts, msg.Content)
		}
	}

	if systemInstructions == "" {
		systemInstructions = s.defaultInstructions
	}

	finalPrompt := strings.Join(userPrompts, "\n")
	if finalPrompt == "" {
		http.Error(w, "Missing user prompt in messages", http.StatusBadRequest)
		return
	}

	// Concurrency serialization block
	s.mu.Lock()
	defer s.mu.Unlock()

	session, err := s.engine.NewSession(systemInstructions)
	if err != nil {
		slog.Error("Failed to initialize session", "error", err)
		http.Error(w, "Failed to initialize inference session: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer session.Close()

	if req.Stream {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
			return
		}

		err = session.RespondStream(finalPrompt, func(token string) {
			chunk := ChatCompletionChunk{
				ID:      "chatcmpl-aplai",
				Object:  "chat.completion.chunk",
				Created: time.Now().Unix(),
				Model:   "apple-silicon",
				Choices: []ChatCompletionChunkChoice{
					{
						Index: 0,
						Delta: ChatMessage{
							Content: token,
						},
						FinishReason: "",
					},
				},
			}
			chunkBytes, _ := json.Marshal(chunk)
			fmt.Fprintf(w, "data: %s\n\n", chunkBytes)
			flusher.Flush()
		})
		if err != nil {
			slog.Error("Inference streaming failed", "error", err)
			return
		}

		finalChunk := ChatCompletionChunk{
			ID:      "chatcmpl-aplai",
			Object:  "chat.completion.chunk",
			Created: time.Now().Unix(),
			Model:   "apple-silicon",
			Choices: []ChatCompletionChunkChoice{
				{
					Index:        0,
					FinishReason: "stop",
				},
			},
		}
		chunkBytes, _ := json.Marshal(finalChunk)
		fmt.Fprintf(w, "data: %s\n\n", chunkBytes)
		fmt.Fprint(w, "data: [DONE]\n\n")
		flusher.Flush()

	} else {
		respText, err := session.Respond(finalPrompt)
		if err != nil {
			slog.Error("Inference failed", "error", err)
			http.Error(w, "Inference failed: "+err.Error(), http.StatusInternalServerError)
			return
		}

		resp := ChatCompletionResponse{
			ID:      "chatcmpl-aplai",
			Object:  "chat.completion",
			Created: time.Now().Unix(),
			Model:   "apple-silicon",
			Choices: []ChatCompletionChoice{
				{
					Index: 0,
					Message: ChatMessage{
						Role:    "assistant",
						Content: respText,
					},
					FinishReason: "stop",
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}
