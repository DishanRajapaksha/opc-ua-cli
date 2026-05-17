package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadClientConfigForProfileUsesDefaultProfile(t *testing.T) {
	path := writeTempConfig(t, `default_profile: local
profiles:
  local:
    endpoint: opc.tcp://localhost:4840
    timeout: 5s
  site-a:
    endpoint: opc.tcp://192.168.1.50:4840
    timeout: 15s
`)

	cfg, err := LoadClientConfigForProfile(path, "")
	if err != nil {
		t.Fatalf("LoadClientConfigForProfile returned error: %v", err)
	}
	if cfg.Endpoint != "opc.tcp://localhost:4840" {
		t.Fatalf("Endpoint = %q", cfg.Endpoint)
	}
	if cfg.Timeout != 5*time.Second {
		t.Fatalf("Timeout = %s", cfg.Timeout)
	}
}

func TestLoadClientConfigForProfileUsesSelectedProfile(t *testing.T) {
	path := writeTempConfig(t, `default_profile: local
profiles:
  local:
    endpoint: opc.tcp://localhost:4840
  site-a:
    endpoint: opc.tcp://192.168.1.50:4840
    policy: Basic256Sha256
    mode: SignAndEncrypt
`)

	cfg, err := LoadClientConfigForProfile(path, "site-a")
	if err != nil {
		t.Fatalf("LoadClientConfigForProfile returned error: %v", err)
	}
	if cfg.Endpoint != "opc.tcp://192.168.1.50:4840" {
		t.Fatalf("Endpoint = %q", cfg.Endpoint)
	}
	if cfg.Policy != "Basic256Sha256" {
		t.Fatalf("Policy = %q", cfg.Policy)
	}
	if cfg.Mode != "SignAndEncrypt" {
		t.Fatalf("Mode = %q", cfg.Mode)
	}
}

func TestLoadClientConfigForProfileAllowsTopLevelDefaults(t *testing.T) {
	path := writeTempConfig(t, `timeout: 20s
policy: None
default_profile: site-a
profiles:
  site-a:
    endpoint: opc.tcp://192.168.1.50:4840
`)

	cfg, err := LoadClientConfigForProfile(path, "")
	if err != nil {
		t.Fatalf("LoadClientConfigForProfile returned error: %v", err)
	}
	if cfg.Endpoint != "opc.tcp://192.168.1.50:4840" {
		t.Fatalf("Endpoint = %q", cfg.Endpoint)
	}
	if cfg.Timeout != 20*time.Second {
		t.Fatalf("Timeout = %s", cfg.Timeout)
	}
}

func TestLoadClientConfigForProfileFailsForMissingProfile(t *testing.T) {
	path := writeTempConfig(t, `profiles:
  local:
    endpoint: opc.tcp://localhost:4840
`)

	_, err := LoadClientConfigForProfile(path, "missing")
	if err == nil {
		t.Fatal("expected error")
	}
}

func writeTempConfig(t *testing.T, contents string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatalf("write temp config: %v", err)
	}
	return path
}
