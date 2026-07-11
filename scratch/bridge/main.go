package main

import (
	"aplai/internal/engine"
	"strings"
)

func main() {
	println("Initializing Apple AI session...")
	session, err := engine.NewSession("You are a helpful AI assistant.")
	if err != nil {
		panic(err)
	}
	defer session.Close()

	prompt := "Explain in one short sentence what is Apple Silicon."
	println("\nPrompt:", prompt)

	// Test synchronous response
	println("Running synchronous respond...")
	resp, err := session.Respond(prompt)
	if err != nil {
		panic(err)
	}
	println("Response:", resp)

	if strings.Contains(resp, "ERROR:") {
		panic("Engine responded with error: " + resp)
	}

	// Test streaming response
	println("\nRunning streaming respond...")
	print("Response stream: ")
	err = session.RespondStream(prompt, func(token string) {
		print(token)
	})
	if err != nil {
		panic(err)
	}
	println("\n\nStream completed successfully!")
}
