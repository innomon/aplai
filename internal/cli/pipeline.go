package cli

import (
	"fmt"
	"io"
	"strings"
)

// Session represents the engine session interface used by the CLI.
type Session interface {
	Respond(prompt string) (string, error)
	RespondStream(prompt string, onToken func(string)) error
}

// RunPipeline runs the prompt pipeline. It reads piped input from r,
// combines it with any trailing prompt, executes it against the session,
// and writes the result to w.
func RunPipeline(session Session, prompt string, r io.Reader, w io.Writer) error {
	// Read all from r (mock or real stdin)
	pipedBytes, err := io.ReadAll(r)
	if err != nil {
		return fmt.Errorf("failed to read from input: %w", err)
	}

	pipedStr := strings.TrimSpace(string(pipedBytes))
	var finalPrompt string

	if pipedStr != "" && prompt != "" {
		finalPrompt = pipedStr + "\n" + prompt
	} else if pipedStr != "" {
		finalPrompt = pipedStr
	} else {
		finalPrompt = prompt
	}

	if finalPrompt == "" {
		return nil
	}

	resp, err := session.Respond(finalPrompt)
	if err != nil {
		return fmt.Errorf("inference failed: %w", err)
	}

	_, err = fmt.Fprintln(w, resp)
	if err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	return nil
}
