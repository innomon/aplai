package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig_ExplicitPath(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "aplai-config-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	configContent := `
server:
  host: "127.0.0.1"
  port: 8080
logger:
  level: "DEBUG"
  console_enabled: true
  file_enabled: true
  dir: "/tmp/logs"
  filename: "aplai.log"
  max_size_mb: 5
  max_backups: 3
engine:
  instructions: "Test instructions"
`
	configPath := filepath.Join(tmpDir, "custom-config.yaml")
	err = os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if cfg.Server.Host != "127.0.0.1" {
		t.Errorf("expected Server.Host to be 127.0.0.1, got %s", cfg.Server.Host)
	}
	if cfg.Server.Port != 8080 {
		t.Errorf("expected Server.Port to be 8080, got %d", cfg.Server.Port)
	}
	if cfg.Logger.Level != "DEBUG" {
		t.Errorf("expected Logger.Level to be DEBUG, got %s", cfg.Logger.Level)
	}
	if cfg.Engine.Instructions != "Test instructions" {
		t.Errorf("expected Engine.Instructions to be 'Test instructions', got %s", cfg.Engine.Instructions)
	}
}

func TestLoadConfig_Defaults(t *testing.T) {
	// Loading with empty path should check default locations or fall back to defaults
	// We can also test Default() directly to verify default values.
	cfg := Default()

	// Verify defaults
	if cfg.Server.Port != 8080 {
		t.Errorf("expected default port 8080, got %d", cfg.Server.Port)
	}
	if cfg.Server.Host != "127.0.0.1" {
		t.Errorf("expected default host 127.0.0.1, got %s", cfg.Server.Host)
	}
	if cfg.Logger.Level != "INFO" {
		t.Errorf("expected default logger level INFO, got %s", cfg.Logger.Level)
	}
}

func TestLoadConfig_ExplicitPathError(t *testing.T) {
	_, err := Load("/non-existent-path/config.yaml")
	if err == nil {
		t.Error("expected error when loading explicit non-existent config path, got nil")
	}
}

func TestLoadConfig_CWDFallback(t *testing.T) {
	configContent := `
server:
  port: 9999
`
	cwdPath := "./config.yaml"
	err := os.WriteFile(cwdPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("failed to write CWD config: %v", err)
	}
	defer os.Remove(cwdPath)

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("failed to load CWD config: %v", err)
	}

	if cfg.Server.Port != 9999 {
		t.Errorf("expected Server.Port from CWD config (9999), got %d", cfg.Server.Port)
	}
}

func TestLoadConfig_MalformedYAML(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "aplai-malformed-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	configPath := filepath.Join(tmpDir, "config.yaml")
	err = os.WriteFile(configPath, []byte("server:\n  port: [invalid-yaml"), 0644)
	if err != nil {
		t.Fatalf("failed to write malformed config: %v", err)
	}

	_, err = Load(configPath)
	if err == nil {
		t.Error("expected error when parsing malformed YAML, got nil")
	}
}

func TestLoadConfig_DirectoryPathError(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "aplai-dir-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	_, err = Load(tmpDir)
	if err == nil {
		t.Error("expected error when trying to read a directory as a config file, got nil")
	}
}

func TestLoadConfig_ExeDirFallback(t *testing.T) {
	exe, err := os.Executable()
	if err != nil {
		t.Skip("skipping test: cannot resolve executable path")
	}
	exeDir := filepath.Dir(exe)

	configContent := `
server:
  port: 8888
`
	exePath := filepath.Join(exeDir, "config.yaml")
	err = os.WriteFile(exePath, []byte(configContent), 0644)
	if err != nil {
		t.Logf("skipping test: cannot write config to exe dir: %v", err)
		return
	}
	defer os.Remove(exePath)

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("failed to load exe dir config: %v", err)
	}

	if cfg.Server.Port != 8888 {
		t.Errorf("expected Server.Port from exe dir config (8888), got %d", cfg.Server.Port)
	}
}
