package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"aplai/internal/cli"
	"aplai/internal/config"
	"aplai/internal/engine"
	"aplai/internal/logger"
	"aplai/internal/server"
)

type RealEngine struct{}

type SessionAdapter struct {
	*engine.Session
}

func (s SessionAdapter) Close() error {
	s.Session.Close()
	return nil
}

func (RealEngine) NewSession(instructions string) (server.Session, error) {
	sess, err := engine.NewSession(instructions)
	if err != nil {
		return nil, err
	}
	return SessionAdapter{sess}, nil
}

func main() {
	configFlag := flag.String("config", "", "Path to configuration file")
	chatFlag := flag.Bool("chat", false, "Start interactive REPL chat session")
	serverFlag := flag.Bool("server", false, "Start OpenAI-compatible HTTP completions server")
	flag.Parse()

	// 1. Load config
	cfg, err := config.Load(*configFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	// 2. Initialize logger
	// If in pipeline mode or non-interactive mode, we want clean output to stdout.
	// But let's configure the logger. By default, logger prints console output to os.Stderr.
	_, err = logger.Init(logger.Config{
		Level:          cfg.Logger.Level,
		ConsoleEnabled: cfg.Logger.ConsoleEnabled,
		FileEnabled:    cfg.Logger.FileEnabled,
		Dir:            cfg.Logger.Dir,
		FileName:       cfg.Logger.FileName,
		MaxSizeMB:      cfg.Logger.MaxSizeMB,
		MaxBackups:     cfg.Logger.MaxBackups,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing logger: %v\n", err)
		os.Exit(1)
	}

	slog.Debug("Config loaded and logger initialized")

	// 3. Subcommand routing
	if *serverFlag {
		slog.Info("Starting OpenAI-compatible HTTP server", "host", cfg.Server.Host, "port", cfg.Server.Port)
		
		realEngine := RealEngine{}
		srv := server.NewServer(cfg.Server.Host, cfg.Server.Port, cfg.Engine.Instructions, realEngine)
		
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		go func() {
			<-sigChan
			cancel()
		}()
		
		err = srv.Start(ctx)
		if err != nil && err != http.ErrServerClosed {
			slog.Error("Server execution failed", "error", err)
			fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
			os.Exit(1)
		}
		slog.Info("Server stopped cleanly")
		os.Exit(0)
	}

	// Initialize engine session
	session, err := engine.NewSession(cfg.Engine.Instructions)
	if err != nil {
		slog.Error("Failed to initialize engine session", "error", err)
		fmt.Fprintf(os.Stderr, "Error initializing inference session: %v\n", err)
		os.Exit(1)
	}
	defer session.Close()

	if *chatFlag {
		err = cli.RunRepl(session, os.Stdin, os.Stdout)
		if err != nil {
			slog.Error("REPL chat session failed", "error", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	// 4. Default mode: Pipeline execution
	// Retrieve trailing positional arguments as the prompt
	promptArgs := flag.Args()
	prompt := strings.Join(promptArgs, " ")

	// Check if stdin is piped
	var isPiped bool
	stat, err := os.Stdin.Stat()
	if err == nil {
		isPiped = (stat.Mode() & os.ModeCharDevice) == 0
	}

	// If no prompt args and not piped, show usage and exit
	if prompt == "" && !isPiped {
		flag.Usage()
		os.Exit(0)
	}

	// Execute pipeline. In non-interactive mode, stdout must be strictly plain text.
	var inputReader io.Reader = os.Stdin
	if !isPiped {
		// If not piped, pass empty reader to avoid blocking on ReadAll
		inputReader = strings.NewReader("")
	}

	err = cli.RunPipeline(session, prompt, inputReader, os.Stdout)
	if err != nil {
		slog.Error("Pipeline execution failed", "error", err)
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
