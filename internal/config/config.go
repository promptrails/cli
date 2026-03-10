package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config holds CLI configuration persisted in ~/.promptrails/config.json.
type Config struct {
	APIURL        string `json:"api_url"`
	WorkspaceID   string `json:"workspace_id,omitempty"`
	WorkspaceName string `json:"workspace_name,omitempty"`
	OutputFormat  string `json:"output_format"`
}

// Credentials holds authentication credentials persisted in ~/.promptrails/credentials.json.
type Credentials struct {
	APIKey string `json:"api_key,omitempty"`
}

// Dir returns the config directory path (~/.promptrails).
func Dir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".promptrails")
}

// Load reads config from disk, returning defaults if not found.
func Load() (*Config, error) {
	path := filepath.Join(Dir(), "config.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{
				APIURL:       "https://api.promptrails.ai",
				OutputFormat: "table",
			}, nil
		}
		return nil, fmt.Errorf("read config: %w", err)
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	return &cfg, nil
}

// Save writes config to disk.
func (c *Config) Save() error {
	dir := Dir()
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	return os.WriteFile(filepath.Join(dir, "config.json"), data, 0644)
}

// LoadCredentials reads credentials from disk.
func LoadCredentials() (*Credentials, error) {
	path := filepath.Join(Dir(), "credentials.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Credentials{}, nil
		}
		return nil, fmt.Errorf("read credentials: %w", err)
	}
	var creds Credentials
	if err := json.Unmarshal(data, &creds); err != nil {
		return nil, fmt.Errorf("parse credentials: %w", err)
	}
	return &creds, nil
}

// Save writes credentials to disk with restricted permissions.
func (c *Credentials) Save() error {
	dir := Dir()
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal credentials: %w", err)
	}
	return os.WriteFile(filepath.Join(dir, "credentials.json"), data, 0600)
}

// Clear removes credentials from disk.
func (c *Credentials) Clear() error {
	path := filepath.Join(Dir(), "credentials.json")
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove credentials: %w", err)
	}
	return nil
}

// IsLoggedIn returns true if an API key is configured.
func (c *Credentials) IsLoggedIn() bool {
	return c.APIKey != ""
}
