package cli

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

// RunRepl starts an interactive REPL chat session. It reads inputs from r,
// processes prompts with streaming responses, and prints them to w.
func RunRepl(session Session, r io.Reader, w io.Writer) error {
	fmt.Fprintln(w, "\033[1;36m=========================================\033[0m")
	fmt.Fprintln(w, "\033[1;36m       APLAI Interactive REPL Chat       \033[0m")
	fmt.Fprintln(w, "\033[1;36m=========================================\033[0m")
	fmt.Fprintln(w, "Type \033[1m/exit\033[0m or \033[1m/quit\033[0m to close the session.")
	fmt.Fprintln(w)

	scanner := bufio.NewScanner(r)

	for {
		fmt.Fprint(w, "\033[1;34mUser > \033[0m")
		
		if !scanner.Scan() {
			break
		}
		
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		if line == "/exit" || line == "/quit" || line == "exit" || line == "quit" {
			break
		}

		fmt.Fprint(w, "\033[1;32mAssistant > \033[0m")
		
		err := session.RespondStream(line, func(token string) {
			fmt.Fprint(w, token)
		})
		if err != nil {
			fmt.Fprintf(w, "\n\033[1;31mError:\033[0m %v\n", err)
			return err
		}
		fmt.Fprintln(w)
		fmt.Fprintln(w)
	}

	fmt.Fprintln(w, "\n\033[1;33mGoodbye!\033[0m")
	return nil
}
