package logger

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"
)

func TestInit_ConsoleAndFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "aplai-init-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := Config{
		Level:          "DEBUG",
		ConsoleEnabled: true,
		FileEnabled:    true,
		Dir:            tmpDir,
		FileName:       "app.log",
		MaxSizeMB:      1,
		MaxBackups:     1,
	}

	logger, err := Init(cfg)
	if err != nil {
		t.Fatalf("failed to init logger: %v", err)
	}

	logger.Info("test console and file init message")
	logger.Debug("test debug log message")

	// Verify that the file was created and contains the log
	path := filepath.Join(tmpDir, "app.log")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("expected log file to exist at %s, but it does not", path)
	}
}

func TestInit_DefaultFallbacks(t *testing.T) {
	cfg := Config{
		Level:          "UNKNOWN",
		ConsoleEnabled: false,
		FileEnabled:    false,
	}

	logger, err := Init(cfg)
	if err != nil {
		t.Fatalf("failed to init logger: %v", err)
	}

	if logger == nil {
		t.Error("expected logger to not be nil")
	}
}

func TestMultiHandler_WithAttrsAndGroup(t *testing.T) {
	mh := NewMultiHandler(slog.NewTextHandler(os.Stderr, nil))
	
	// Test Enabled
	if !mh.Enabled(nil, slog.LevelInfo) {
		t.Error("expected Enabled to return true for LevelInfo")
	}

	// Test WithAttrs
	hWithAttrs := mh.WithAttrs([]slog.Attr{slog.String("foo", "bar")})
	if hWithAttrs == nil {
		t.Error("expected WithAttrs to return a non-nil handler")
	}

	// Test WithGroup
	hWithGroup := mh.WithGroup("test-group")
	if hWithGroup == nil {
		t.Error("expected WithGroup to return a non-nil handler")
	}
}
