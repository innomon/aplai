package main

import (
	"aplai/internal/logger"
	"os"
	"path/filepath"
)

func main() {
	tmpDir, err := os.MkdirTemp("", "aplai-verify-*")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := logger.Config{
		Level:          "DEBUG",
		ConsoleEnabled: true,
		FileEnabled:    true,
		Dir:            tmpDir,
		FileName:       "aplai.log",
		MaxSizeMB:      1,
		MaxBackups:     2,
	}

	log, err := logger.Init(cfg)
	if err != nil {
		panic(err)
	}

	// Emit normal log
	log.Info("Normal log from main")

	// Emit via CustomLogger to verify caller preservation
	cl := logger.NewCustomLogger(log, "verify")
	cl.Infof("Hello from CustomLogger wrapper")

	// Read the generated JSONL log file
	path := filepath.Join(tmpDir, "aplai.log")
	data, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}

	println("\n--- Generated JSONL Log Content ---")
	print(string(data))
	println("-----------------------------------\n")
}
