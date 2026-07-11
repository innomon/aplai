package cli

import (
	"bytes"
	"errors"
	"testing"
)

type MockSession struct {
	LastPrompt string
	MockResult string
	MockErr    error
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
	onToken(m.MockResult)
	return nil
}

func TestRunPipeline_PipedInput(t *testing.T) {
	mockSess := &MockSession{
		MockResult: "This is a response to piped input",
	}

	input := bytes.NewBufferString("Hello from pipe")
	var output bytes.Buffer

	err := RunPipeline(mockSess, "", input, &output)
	if err != nil {
		t.Fatalf("failed to run pipeline: %v", err)
	}

	if mockSess.LastPrompt != "Hello from pipe" {
		t.Errorf("expected prompt 'Hello from pipe', got '%s'", mockSess.LastPrompt)
	}

	if output.String() != "This is a response to piped input\n" {
		t.Errorf("expected output 'This is a response to piped input\\n', got '%s'", output.String())
	}
}

func TestRunPipeline_PromptArg(t *testing.T) {
	mockSess := &MockSession{
		MockResult: "This is a response to prompt arg",
	}

	var input bytes.Buffer // Empty stdin
	var output bytes.Buffer

	err := RunPipeline(mockSess, "Hello from arg", &input, &output)
	if err != nil {
		t.Fatalf("failed to run pipeline: %v", err)
	}

	if mockSess.LastPrompt != "Hello from arg" {
		t.Errorf("expected prompt 'Hello from arg', got '%s'", mockSess.LastPrompt)
	}

	if output.String() != "This is a response to prompt arg\n" {
		t.Errorf("expected output 'This is a response to prompt arg\\n', got '%s'", output.String())
	}
}

func TestRunPipeline_Error(t *testing.T) {
	mockSess := &MockSession{
		MockErr: errors.New("engine failure"),
	}

	var input bytes.Buffer
	var output bytes.Buffer

	err := RunPipeline(mockSess, "Hello", &input, &output)
	if err == nil {
		t.Error("expected error from pipeline, got nil")
	}
}

func TestRunPipeline_EmptyPrompt(t *testing.T) {
	mockSess := &MockSession{}
	var input bytes.Buffer
	var output bytes.Buffer

	err := RunPipeline(mockSess, "", &input, &output)
	if err != nil {
		t.Fatalf("expected nil error on empty prompt, got %v", err)
	}
}

type errorReader struct{}

func (errorReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("read error")
}

func TestRunPipeline_ReadError(t *testing.T) {
	mockSess := &MockSession{}
	var output bytes.Buffer

	err := RunPipeline(mockSess, "", errorReader{}, &output)
	if err == nil {
		t.Error("expected error when reading input fails, got nil")
	}
}

type errorWriter struct{}

func (errorWriter) Write(p []byte) (n int, err error) {
	return 0, errors.New("write error")
}

func TestRunPipeline_WriteError(t *testing.T) {
	mockSess := &MockSession{
		MockResult: "response",
	}
	var input bytes.Buffer

	err := RunPipeline(mockSess, "Hello", &input, errorWriter{})
	if err == nil {
		t.Error("expected error when writing output fails, got nil")
	}
}
