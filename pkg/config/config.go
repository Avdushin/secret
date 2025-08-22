// pkg/config/config.go
package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Backend     string   `yaml:"backend"`
	GPGKey      string   `yaml:"gpg_key,omitempty"`
	ProjectName string   `yaml:"project_name,omitempty"`
	SecretFiles []string `yaml:"secret_files,omitempty"`
	SecretDir   string   `yaml:"secret_dir,omitempty"`
}

var DefaultSecretFiles = []string{".env", "dev.env", "config.json", ".config.yaml"}

func LoadConfig() (*Config, error) {
	configPath := filepath.Join(".secret", "config.yaml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var cfg Config
	err = yaml.Unmarshal(data, &cfg)
	return &cfg, err
}

func SaveConfig(cfg *Config) error {
	if err := os.MkdirAll(".secret", 0700); err != nil {
		return err
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(".secret", "config.yaml"), data, 0600)
}
