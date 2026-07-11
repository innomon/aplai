package engine

import (
	"strings"
	"testing"
)

func TestAppleAISession(t *testing.T) {
	// Initialize a session
	session, err := NewSession("You are a helpful assistant.")
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}
	defer session.Close()

	// Test synchronous response
	resp, err := session.Respond("Hello")
	if err != nil {
		t.Fatalf("failed to respond: %v", err)
	}

	if resp == "" {
		t.Error("expected non-empty response")
	}

	if strings.Contains(resp, "ERROR:") {
		t.Errorf("received error response: %s", resp)
	}

	// Test streaming response
	tokens := []string{}
	err = session.RespondStream("Hello", func(token string) {
		tokens = append(tokens, token)
	})

	if err != nil {
		t.Fatalf("failed to respond stream: %v", err)
	}

	if len(tokens) == 0 {
		t.Error("expected to receive streamed tokens")
	}
}
