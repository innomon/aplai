package logger

import (
	"bytes"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
)

func TestLogRotator_WriteAndRotate(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "aplai-logger-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	filename := "test.log"
	// Create a rotator with very small max size (e.g. 10 bytes) to force rotation easily.
	// Wait, the NewLogRotator parameter is in MB, but let's test with a small size.
	// Wait, if NewLogRotator takes MB, 1 MB is 1024*1024 bytes.
	// Let's modify NewLogRotator or expose internal fields so we can set maxSize directly in tests, or we can pass size in bytes/MB.
	// Let's implement NewLogRotator using MB but maybe allow configuring maxSize in bytes for testability, or just write enough MBs?
	// Actually, if we look at the spec code in aplai-prompt.md:
	// func NewLogRotator(dir, filename string, maxSizeMB int, maxBackups int) (*LogRotator, error)
	// it sets r.maxSize = int64(maxSizeMB) * 1024 * 1024.
	// In the test, we can manually overwrite r.maxSize to a small number of bytes! E.g. 20 bytes.
	r, err := NewLogRotator(tmpDir, filename, 1, 2)
	if err != nil {
		t.Fatalf("failed to create LogRotator: %v", err)
	}
	defer r.Close()

	// Overwrite maxSize to 20 bytes for easy testing of rotation
	r.maxSize = 20

	// Write first log
	msg1 := []byte("hello world 1\n") // 14 bytes
	n, err := r.Write(msg1)
	if err != nil {
		t.Fatalf("failed to write: %v", err)
	}
	if n != len(msg1) {
		t.Errorf("expected to write %d bytes, wrote %d", len(msg1), n)
	}

	// Write second log - this should trigger rotation because 14 + 14 = 28 > 20
	msg2 := []byte("hello world 2\n") // 14 bytes
	n, err = r.Write(msg2)
	if err != nil {
		t.Fatalf("failed to write second: %v", err)
	}
	if n != len(msg2) {
		t.Errorf("expected to write %d bytes, wrote %d", len(msg2), n)
	}

	// Verify that the original file exists and has the second message
	currentPath := filepath.Join(tmpDir, filename)
	content, err := os.ReadFile(currentPath)
	if err != nil {
		t.Fatalf("failed to read current log file: %v", err)
	}
	if !bytes.Contains(content, msg2) {
		t.Errorf("current log should contain msg2, got: %s", content)
	}

	// Verify that a backup file was created
	files, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatalf("failed to read dir: %v", err)
	}

	var backupFound bool
	for _, f := range files {
		if f.Name() != filename {
			backupFound = true
			backupContent, err := os.ReadFile(filepath.Join(tmpDir, f.Name()))
			if err != nil {
				t.Fatalf("failed to read backup file: %v", err)
			}
			if !bytes.Contains(backupContent, msg1) {
				t.Errorf("backup log should contain msg1, got: %s", backupContent)
			}
		}
	}
	if !backupFound {
		t.Error("expected to find a backup log file, but none was found")
	}
}

func TestLogRotator_Pruning(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "aplai-logger-prune-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	filename := "test.log"
	// Max backups = 2
	r, err := NewLogRotator(tmpDir, filename, 1, 2)
	if err != nil {
		t.Fatalf("failed to create LogRotator: %v", err)
	}
	defer r.Close()

	r.maxSize = 10 // Rotate on every write of >10 bytes

	// Write 4 messages of 15 bytes each. This should trigger 3 rotations.
	// With maxBackups = 2, we should have the active file + at most 2 backups (total 3 files max).
	for i := 0; i < 4; i++ {
		_, err := r.Write([]byte("msg exceeding 10\n"))
		if err != nil {
			t.Fatalf("failed to write: %v", err)
		}
	}

	files, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatalf("failed to read dir: %v", err)
	}

	// Active file + max 2 backups = 3 files max
	if len(files) > 3 {
		t.Errorf("expected at most 3 files, got %d", len(files))
	}
}

func TestMultiHandler(t *testing.T) {
	var buf1, buf2 bytes.Buffer
	h1 := slog.NewTextHandler(&buf1, nil)
	h2 := slog.NewJSONHandler(&buf2, nil)

	mh := NewMultiHandler(h1, h2)
	logger := slog.New(mh)

	logger.Info("test message", slog.String("key", "val"))

	out1 := buf1.String()
	out2 := buf2.String()

	if !bytes.Contains([]byte(out1), []byte("test message")) {
		t.Errorf("expected text handler output to contain message, got: %s", out1)
	}
	if !bytes.Contains([]byte(out2), []byte("test message")) || !bytes.Contains([]byte(out2), []byte(`"key":"val"`)) {
		t.Errorf("expected json handler output to contain message and attributes, got: %s", out2)
	}
}

func TestNewLogRotator_InvalidDir(t *testing.T) {
	// A path that is invalid / impossible to create (e.g. empty or permission-denied directory)
	// We can try to use a path with null character or one under a root folder we don't have permission to create
	_, err := NewLogRotator("/invalid-path-that-fails-mkdir/one/two/three", "app.log", 1, 1)
	if err == nil {
		t.Error("expected error when directory creation fails, got nil")
	}
}

func TestNewLogRotator_OpenError(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "aplai-open-fail-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a directory named "app.log"
	err = os.Mkdir(filepath.Join(tmpDir, "app.log"), 0755)
	if err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}

	// Try to create a LogRotator with "app.log" as filename. Since it's a directory, open() should fail.
	_, err = NewLogRotator(tmpDir, "app.log", 1, 1)
	if err == nil {
		t.Error("expected error when trying to open a directory as a file, got nil")
	}
}

