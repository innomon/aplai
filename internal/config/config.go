package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// ServerConfig defines configuration parameters for the HTTP server.
type ServerConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

// LoggerConfig defines parameters for the dual console/file structured logger.
type LoggerConfig struct {
	Level          string `yaml:"level"`
	ConsoleEnabled bool   `yaml:"console_enabled"`
	FileEnabled    bool   `yaml:"file_enabled"`
	Dir            string `yaml:"dir"`
	FileName       string `yaml:"filename"`
	MaxSizeMB      int    `yaml:"max_size_mb"`
	MaxBackups     int    `yaml:"max_backups"`
}

// EngineConfig defines configuration parameters for the AI inference engine.
type EngineConfig struct {
	Instructions string `yaml:"instructions"`
}

// Config represents the global configuration for APLAI.
type Config struct {
	Server ServerConfig `yaml:"server"`
	Logger LoggerConfig `yaml:"logger"`
	Engine EngineConfig `yaml:"engine"`
}

// Default returns a Config initialized with sensible default values.
func Default() *Config {
	return &Config{
		Server: ServerConfig{
			Host: "127.0.0.1",
			Port: 8080,
		},
		Logger: LoggerConfig{
			Level:          "INFO",
			ConsoleEnabled: true,
			FileEnabled:    true,
			Dir:            "logs",
			FileName:       "aplai.log",
			MaxSizeMB:      10,
			MaxBackups:     5,
		},
		Engine: EngineConfig{
			Instructions: "You are a helpful on-device AI assistant.",
		},
	}
}

// Load loads the configuration from the specified custom path, or searches standard paths:
// 1. Current directory (./config.yaml)
// 2. Executable directory (<exe>/config.yaml)
// If no configuration file is found, it returns the default configuration.
func Load(customPath string) (*Config, error) {
	cfg := Default()

	var resolvedPath string

	if customPath != "" {
		resolvedPath = customPath
		// If customPath is explicitly set, failure to find it is considered an error.
		if _, err := os.Stat(resolvedPath); os.IsNotExist(err) {
			return nil, fmt.Errorf("explicit config file not found: %s", customPath)
		}
	} else {
		// 1. Check current directory
		cwdPath := "./config.yaml"
		if _, err := os.Stat(cwdPath); err == nil {
			resolvedPath = cwdPath
		} else {
			// 2. Check executable directory
			exe, err := os.Executable()
			if err == nil {
				exePath := filepath.Join(filepath.Dir(exe), "config.yaml")
				if _, err := os.Stat(exePath); err == nil {
					resolvedPath = exePath
				}
			}
		}
	}

	// If no config file was found, return defaults.
	if resolvedPath == "" {
		return cfg, nil
	}

	// Read and parse config file
	data, err := os.ReadFile(resolvedPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	err = yaml.Unmarshal(data, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config YAML: %w", err)
	}

	return cfg, nil
}
