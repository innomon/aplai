package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestRunRepl_CleanExit(t *testing.T) {
	mockSess := &MockSession{
		MockResult: "Hello back!",
	}

	// Input consists of:
	// 1. A prompt: "How are you?"
	// 2. An exit command: "/exit"
	input := bytes.NewBufferString("How are you?\n/exit\n")
	var output bytes.Buffer

	err := RunRepl(mockSess, input, &output)
	if err != nil {
		t.Fatalf("failed to run REPL: %v", err)
	}

	outStr := output.String()
	if !strings.Contains(outStr, "Hello back!") {
		t.Errorf("expected output to contain assistant response, got: %s", outStr)
	}
	if !strings.Contains(outStr, "Goodbye") {
		t.Errorf("expected output to contain goodbye message, got: %s", outStr)
	}
}

func TestRunRepl_QuitCommand(t *testing.T) {
	mockSess := &MockSession{
		MockResult: "response",
	}

	input := bytes.NewBufferString("/quit\n")
	var output bytes.Buffer

	err := RunRepl(mockSess, input, &output)
	if err != nil {
		t.Fatalf("failed to run REPL: %v", err)
	}

	outStr := output.String()
	if strings.Contains(outStr, "response") {
		t.Error("expected REPL to quit immediately without executing prompts")
	}
}
