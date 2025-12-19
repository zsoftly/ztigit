// Package config handles configuration management for ztigit
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
	"github.com/zalando/go-keyring"
)

const (
	// KeyringService is the service name used for keychain storage
	KeyringService = "ztigit"
)

// Config holds the application configuration
type Config struct {
	// Default provider to use
	DefaultProvider string `mapstructure:"default_provider"`

	// GitLab configuration
	GitLab GitLabConfig `mapstructure:"gitlab"`

	// GitHub configuration
	GitHub GitHubConfig `mapstructure:"github"`

	// Mirror configuration
	Mirror MirrorConfig `mapstructure:"mirror"`

	// Debug mode
	Debug bool `mapstructure:"debug"`
}

// GitLabConfig holds GitLab-specific configuration
type GitLabConfig struct {
	Token   string `mapstructure:"token"`
	BaseURL string `mapstructure:"base_url"`
}

// GitHubConfig holds GitHub-specific configuration
type GitHubConfig struct {
	Token   string `mapstructure:"token"`
	BaseURL string `mapstructure:"base_url"`
}

// MirrorConfig holds mirror operation configuration
type MirrorConfig struct {
	// Base directory for cloned repositories
	BaseDir string `mapstructure:"base_dir"`

	// Number of parallel clone/pull operations
	Parallel int `mapstructure:"parallel"`

	// Skip archived repositories
	SkipArchived bool `mapstructure:"skip_archived"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	homeDir, err := os.UserHomeDir()
	if err != nil || homeDir == "" {
		homeDir = "." // Fallback to current directory
	}

	return &Config{
		DefaultProvider: "gitlab",
		GitLab: GitLabConfig{
			BaseURL: "https://gitlab.com",
		},
		GitHub: GitHubConfig{
			BaseURL: "https://github.com",
		},
		Mirror: MirrorConfig{
			BaseDir:      filepath.Join(homeDir, "git-repos"),
			Parallel:     4,
			SkipArchived: true,
		},
		Debug: false,
	}
}

// Load loads configuration from file and environment variables
func Load() (*Config, error) {
	cfg := DefaultConfig()

	// Set up viper
	viper.SetConfigName("ztigit")
	viper.SetConfigType("yaml")

	// Config file locations
	homeDir, err := os.UserHomeDir()
	if err != nil || homeDir == "" {
		homeDir = "." // Fallback to current directory
	}
	viper.AddConfigPath(filepath.Join(homeDir, ".config", "ztigit"))
	viper.AddConfigPath(filepath.Join(homeDir, ".ztigit"))
	viper.AddConfigPath(".")

	// Environment variable prefix
	viper.SetEnvPrefix("ZTIGIT")
	viper.AutomaticEnv()

	// Map environment variables
	viper.BindEnv("gitlab.token", "GITLAB_TOKEN", "ZTIGIT_GITLAB_TOKEN")
	viper.BindEnv("gitlab.base_url", "GITLAB_URL", "ZTIGIT_GITLAB_URL")
	viper.BindEnv("github.token", "GITHUB_TOKEN", "ZTIGIT_GITHUB_TOKEN")
	viper.BindEnv("github.base_url", "GITHUB_URL", "ZTIGIT_GITHUB_URL")

	// Try to read config file (not required, ignore errors)
	_ = viper.ReadInConfig()

	// Unmarshal into config struct
	if err := viper.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	return cfg, nil
}

// GetConfigDir returns the configuration directory path
func GetConfigDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil || homeDir == "" {
		homeDir = "." // Fallback to current directory
	}
	return filepath.Join(homeDir, ".config", "ztigit")
}

// GetConfigFile returns the configuration file path
func GetConfigFile() string {
	return filepath.Join(GetConfigDir(), "ztigit.yaml")
}

// Save saves the configuration to file
// Tokens are stored in system keychain when available, otherwise in config file
func Save(cfg *Config) error {
	configDir := GetConfigDir()

	// Create config directory if it doesn't exist (0700 for security)
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Try to store tokens in keychain (secure storage)
	keyringWorked := false
	if cfg.GitLab.Token != "" {
		if err := SetTokenSecure("gitlab", cfg.GitLab.Token); err == nil && keyringAvailable {
			keyringWorked = true
		}
	}
	if cfg.GitHub.Token != "" {
		if err := SetTokenSecure("github", cfg.GitHub.Token); err == nil && keyringAvailable {
			keyringWorked = true
		}
	}

	// Set values in viper (tokens only if keychain not available)
	viper.Set("default_provider", cfg.DefaultProvider)
	viper.Set("gitlab.base_url", cfg.GitLab.BaseURL)
	viper.Set("github.base_url", cfg.GitHub.BaseURL)
	viper.Set("mirror.base_dir", cfg.Mirror.BaseDir)
	viper.Set("mirror.parallel", cfg.Mirror.Parallel)
	viper.Set("mirror.skip_archived", cfg.Mirror.SkipArchived)
	viper.Set("debug", cfg.Debug)

	// Only store tokens in config file if keychain is not available
	if !keyringWorked {
		viper.Set("gitlab.token", cfg.GitLab.Token)
		viper.Set("github.token", cfg.GitHub.Token)
	} else {
		// Clear tokens from config file if using keychain
		viper.Set("gitlab.token", "")
		viper.Set("github.token", "")
	}

	// Write config file
	configFile := GetConfigFile()
	if err := viper.WriteConfigAs(configFile); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	// Set restrictive permissions on config file
	if err := os.Chmod(configFile, 0600); err != nil {
		return fmt.Errorf("failed to set config file permissions: %w", err)
	}

	return nil
}

// GetToken returns the token for the specified provider
// Checks keychain first, then falls back to config file/env vars
func (c *Config) GetToken(provider string) string {
	// Try keychain first (most secure)
	if token := GetTokenSecure(provider); token != "" {
		return token
	}

	// Fall back to config file / environment variable
	switch provider {
	case "gitlab":
		return c.GitLab.Token
	case "github":
		return c.GitHub.Token
	default:
		return ""
	}
}

// GetBaseURL returns the base URL for the specified provider
func (c *Config) GetBaseURL(provider string) string {
	switch provider {
	case "gitlab":
		return c.GitLab.BaseURL
	case "github":
		return c.GitHub.BaseURL
	default:
		return ""
	}
}

// keyringAvailable checks if the system keyring is available
var keyringAvailable = true

// SetTokenSecure stores a token securely in the system keychain
// Falls back to config file if keychain is unavailable
func SetTokenSecure(provider, token string) error {
	if !keyringAvailable {
		return nil // Will use config file fallback
	}

	key := fmt.Sprintf("%s-token", provider)
	err := keyring.Set(KeyringService, key, token)
	if err != nil {
		// Keyring not available (e.g., headless server), mark as unavailable
		keyringAvailable = false
		return nil // Fall back to config file storage
	}
	return nil
}

// GetTokenSecure retrieves a token from the system keychain
// Returns empty string if not found or keychain unavailable
func GetTokenSecure(provider string) string {
	if !keyringAvailable {
		return ""
	}

	key := fmt.Sprintf("%s-token", provider)
	token, err := keyring.Get(KeyringService, key)
	if err != nil {
		// Keyring not available or token not found
		if err == keyring.ErrNotFound {
			return ""
		}
		// Mark keyring as unavailable for future calls
		keyringAvailable = false
		return ""
	}
	return token
}

// DeleteTokenSecure removes a token from the system keychain
func DeleteTokenSecure(provider string) error {
	if !keyringAvailable {
		return nil
	}

	key := fmt.Sprintf("%s-token", provider)
	err := keyring.Delete(KeyringService, key)
	if err != nil && err != keyring.ErrNotFound {
		return err
	}
	return nil
}

// IsKeyringAvailable returns whether the system keyring is available
func IsKeyringAvailable() bool {
	return keyringAvailable
}
