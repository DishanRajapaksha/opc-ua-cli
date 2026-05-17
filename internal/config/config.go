package config

import (
	"errors"
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

const DefaultEndpoint = "opc.tcp://localhost:4840"

// ClientConfig contains connection and security settings for an OPC UA client session.
type ClientConfig struct {
	Endpoint string
	Policy   string
	Mode     string
	Username string
	Password string
	CertFile string
	KeyFile  string
	Timeout  time.Duration
}

func DefaultClientConfig() ClientConfig {
	return ClientConfig{
		Endpoint: DefaultEndpoint,
		Policy:   "None",
		Mode:     "None",
		Timeout:  10 * time.Second,
	}
}

func LoadClientConfig(path string) (ClientConfig, error) {
	cfg := DefaultClientConfig()
	if path == "" {
		return cfg, nil
	}

	contents, err := os.ReadFile(path)
	if err != nil {
		return cfg, fmt.Errorf("read config %q: %w", path, err)
	}

	var file configFile
	if err := yaml.Unmarshal(contents, &file); err != nil {
		return cfg, fmt.Errorf("parse config %q: %w", path, err)
	}

	if file.Endpoint != "" {
		cfg.Endpoint = file.Endpoint
	}
	if file.Policy != "" {
		cfg.Policy = file.Policy
	}
	if file.Mode != "" {
		cfg.Mode = file.Mode
	}
	if file.Username != "" {
		cfg.Username = file.Username
	}
	if file.Password != "" {
		cfg.Password = file.Password
	}
	if file.CertFile != "" {
		cfg.CertFile = file.CertFile
	}
	if file.KeyFile != "" {
		cfg.KeyFile = file.KeyFile
	}
	if file.Timeout != "" {
		timeout, err := time.ParseDuration(file.Timeout)
		if err != nil {
			return cfg, fmt.Errorf("parse timeout %q: %w", file.Timeout, err)
		}
		cfg.Timeout = timeout
	}

	if cfg.Endpoint == "" {
		return cfg, errors.New("endpoint cannot be empty")
	}

	return cfg, nil
}

func (c ClientConfig) UsesSecurity() bool {
	return c.Username != "" || c.Policy != "None" || c.Mode != "None" || c.CertFile != "" || c.KeyFile != ""
}

type configFile struct {
	Endpoint string `yaml:"endpoint"`
	Policy   string `yaml:"policy"`
	Mode     string `yaml:"mode"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	CertFile string `yaml:"cert"`
	KeyFile  string `yaml:"key"`
	Timeout  string `yaml:"timeout"`
}
