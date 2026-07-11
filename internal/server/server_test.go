package server

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type MockSession struct {
	LastSystemPrompt string
	LastPrompt       string
	MockResult       string
	MockErr          error
}

func (m *MockSession) Respond(prompt string) (string, error) {
	m.LastPrompt = prompt
	return m.MockResult, m.MockErr
}

func (m *MockSession) RespondStream(prompt string, onToken func(string)) error {
	m.LastPrompt = prompt
	if m.MockErr != nil {
		return m.MockErr
	}
	// Simulate streaming tokens
	tokens := strings.Fields(m.MockResult)
	for _, tok := range tokens {
		onToken(tok + " ")
	}
	return nil
}

func (m *MockSession) Close() error {
	return nil
}

type MockEngine struct {
	LastInstructions string
	MockSess         *MockSession
	MockErr          error
}

func (e *MockEngine) NewSession(instructions string) (Session, error) {
	e.LastInstructions = instructions
	if e.MockErr != nil {
		return nil, e.MockErr
	}
	return e.MockSess, nil
}

func TestHandleChatCompletions_JSON(t *testing.T) {
	mockSess := &MockSession{
		MockResult: "Hello there!",
	}
	mockEngine := &MockEngine{
		MockSess: mockSess,
	}

	srv := NewServer("127.0.0.1", 8080, "Default system instructions", mockEngine)

	reqBody := ChatCompletionRequest{
		Model: "apple-silicon",
		Messages: []ChatMessage{
			{Role: "system", Content: "Custom system instruction"},
			{Role: "user", Content: "Hi"},
		},
		Stream: false,
	}

	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewReader(bodyBytes))
	rec := httptest.NewRecorder()

	srv.Router().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var resp ChatCompletionResponse
	err := json.NewDecoder(rec.Body).Decode(&resp)
	if err != nil {
		t.Fatalf("failed to decode JSON response: %v", err)
	}

	if len(resp.Choices) == 0 {
		t.Fatal("expected at least one choice in response")
	}

	if resp.Choices[0].Message.Content != "Hello there!" {
		t.Errorf("expected content 'Hello there!', got '%s'", resp.Choices[0].Message.Content)
	}

	if mockEngine.LastInstructions != "Custom system instruction" {
		t.Errorf("expected system instructions to be custom, got '%s'", mockEngine.LastInstructions)
	}

	if mockSess.LastPrompt != "Hi" {
		t.Errorf("expected prompt 'Hi', got '%s'", mockSess.LastPrompt)
	}
}

func TestHandleChatCompletions_Stream(t *testing.T) {
	mockSess := &MockSession{
		MockResult: "Hello there from stream!",
	}
	mockEngine := &MockEngine{
		MockSess: mockSess,
	}

	srv := NewServer("127.0.0.1", 8080, "Default system instructions", mockEngine)

	reqBody := ChatCompletionRequest{
		Model: "apple-silicon",
		Messages: []ChatMessage{
			{Role: "user", Content: "Hi stream"},
		},
		Stream: true,
	}

	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewReader(bodyBytes))
	rec := httptest.NewRecorder()

	srv.Router().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	respStr := rec.Body.String()
	if !strings.Contains(respStr, "data:") {
		t.Error("expected event stream data prefix in response")
	}
	if !strings.Contains(respStr, "[DONE]") {
		t.Error("expected event stream to terminate with [DONE]")
	}
}

func TestHandleChatCompletions_EngineError(t *testing.T) {
	mockEngine := &MockEngine{
		MockErr: errors.New("failed to load model"),
	}

	srv := NewServer("127.0.0.1", 8080, "Default", mockEngine)

	reqBody := ChatCompletionRequest{
		Messages: []ChatMessage{{Role: "user", Content: "Hi"}},
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewReader(bodyBytes))
	rec := httptest.NewRecorder()

	srv.Router().ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", rec.Code)
	}
}

func TestServer_StartAndShutdown(t *testing.T) {
	mockEngine := &MockEngine{}
	srv := NewServer("127.0.0.1", 0, "Default", mockEngine) // Port 0 to allocate automatically

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errChan := make(chan error, 1)
	go func() {
		errChan <- srv.Start(ctx)
	}()

	// Trigger shutdown by cancelling context
	cancel()

	err := <-errChan
	if err != nil && err != http.ErrServerClosed {
		t.Fatalf("server start failed: %v", err)
	}
}

func TestHandleChatCompletions_InvalidMethod(t *testing.T) {
	srv := NewServer("127.0.0.1", 8080, "Default", &MockEngine{})
	req := httptest.NewRequest(http.MethodGet, "/v1/chat/completions", nil)
	rec := httptest.NewRecorder()

	srv.Router().ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405 Method Not Allowed, got %d", rec.Code)
	}
}

func TestHandleChatCompletions_MalformedJSON(t *testing.T) {
	srv := NewServer("127.0.0.1", 8080, "Default", &MockEngine{})
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader("{invalid-json"))
	rec := httptest.NewRecorder()

	srv.Router().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400 Bad Request, got %d", rec.Code)
	}
}

func TestHandleChatCompletions_MissingPrompt(t *testing.T) {
	srv := NewServer("127.0.0.1", 8080, "Default", &MockEngine{})
	reqBody := ChatCompletionRequest{
		Messages: []ChatMessage{{Role: "system", Content: "System note only"}},
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewReader(bodyBytes))
	rec := httptest.NewRecorder()

	srv.Router().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400 Bad Request, got %d", rec.Code)
	}
}

func TestHandleChatCompletions_RespondError(t *testing.T) {
	mockSess := &MockSession{
		MockErr: errors.New("respond failure"),
	}
	mockEngine := &MockEngine{
		MockSess: mockSess,
	}
	srv := NewServer("127.0.0.1", 8080, "Default", mockEngine)

	reqBody := ChatCompletionRequest{
		Messages: []ChatMessage{{Role: "user", Content: "Hi"}},
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewReader(bodyBytes))
	rec := httptest.NewRecorder()

	srv.Router().ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500 Internal Server Error, got %d", rec.Code)
	}
}

type ConcurrencyMockSession struct {
	mu           sync.Mutex
	activeCount  int32
	maxActive    int32
	RespondDelay time.Duration
}

func (m *ConcurrencyMockSession) Respond(prompt string) (string, error) {
	currentActive := atomic.AddInt32(&m.activeCount, 1)
	defer atomic.AddInt32(&m.activeCount, -1)

	m.mu.Lock()
	if currentActive > m.maxActive {
		m.maxActive = currentActive
	}
	m.mu.Unlock()

	time.Sleep(m.RespondDelay)
	return "done", nil
}

func (m *ConcurrencyMockSession) RespondStream(prompt string, onToken func(string)) error {
	return nil
}

func (m *ConcurrencyMockSession) Close() error {
	return nil
}

type ConcurrencyMockEngine struct {
	MockSess *ConcurrencyMockSession
}

func (e *ConcurrencyMockEngine) NewSession(instructions string) (Session, error) {
	return e.MockSess, nil
}

func TestHandleChatCompletions_ConcurrencyQueueing(t *testing.T) {
	mockSess := &ConcurrencyMockSession{
		RespondDelay: 20 * time.Millisecond,
	}
	mockEngine := &ConcurrencyMockEngine{
		MockSess: mockSess,
	}

	srv := NewServer("127.0.0.1", 8080, "Default", mockEngine)

	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			reqBody := ChatCompletionRequest{
				Messages: []ChatMessage{{Role: "user", Content: "Hi"}},
			}
			bodyBytes, _ := json.Marshal(reqBody)
			req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewReader(bodyBytes))
			rec := httptest.NewRecorder()
			srv.Router().ServeHTTP(rec, req)
		}()
	}

	wg.Wait()

	mockSess.mu.Lock()
	maxActive := mockSess.maxActive
	mockSess.mu.Unlock()

	if maxActive > 1 {
		t.Errorf("expected maximum concurrent engine requests to be 1 (serialized), got %d", maxActive)
	} else {
		t.Logf("concurrency queueing verified: maximum concurrent engine requests was %d", maxActive)
	}
}
