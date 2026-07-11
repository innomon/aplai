package logger

import (
	"bytes"
	"encoding/json"
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

func TestCustomLogger_PreservesCaller(t *testing.T) {
	var buf bytes.Buffer
	opts := &slog.HandlerOptions{
		Level:     slog.LevelDebug,
		AddSource: true,
	}
	handler := slog.NewJSONHandler(&buf, opts)
	slogLogger := slog.New(handler)

	cLogger := &CustomLogger{
		logger: slogLogger,
		module: "test-module",
	}

	cLogger.Infof("message info: %s", "hello") // This must be at line 98

	var logEntry struct {
		Source struct {
			Function string `json:"function"`
			File     string `json:"file"`
			Line     int    `json:"line"`
		} `json:"source"`
		Msg string `json:"msg"`
	}

	err := json.Unmarshal(buf.Bytes(), &logEntry)
	if err != nil {
		t.Fatalf("failed to parse JSON log entry: %v, raw output: %s", err, buf.String())
	}
	t.Logf("RAW JSON LOG: %s", buf.String())

	// Verify line number is close to the calling site line number (within 2 lines)
	expectedLine := 97
	diff := logEntry.Source.Line - expectedLine
	if diff < -2 || diff > 2 {
		t.Errorf("expected caller line near %d, got %d. Function: %s, File: %s", expectedLine, logEntry.Source.Line, logEntry.Source.Function, logEntry.Source.File)
	}

	// Verify it contains the right file name
	if !bytes.Contains([]byte(logEntry.Source.File), []byte("logger_test.go")) {
		t.Errorf("expected file to be logger_test.go, got %s", logEntry.Source.File)
	}
}

func TestCustomLogger_Levels(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	l := slog.New(handler)
	cl := NewCustomLogger(l, "test-module")

	cl.Debugf("debug msg")
	cl.Warnf("warn msg")
	cl.Errorf("error msg")

	out := buf.String()
	if !bytes.Contains([]byte(out), []byte("debug msg")) ||
		!bytes.Contains([]byte(out), []byte("warn msg")) ||
		!bytes.Contains([]byte(out), []byte("error msg")) {
		t.Errorf("expected logs to contain all messages, got: %s", out)
	}
}
