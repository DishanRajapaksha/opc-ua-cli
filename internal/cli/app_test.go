package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunVersion(t *testing.T) {
	var out, errOut bytes.Buffer
	code := NewApp(&out, &errOut).Run([]string{"version"})
	if code != exitSuccess {
		t.Fatalf("Run(version) = %d, want %d", code, exitSuccess)
	}
	if !strings.Contains(out.String(), "opc-ua-cli development") {
		t.Fatalf("stdout = %q", out.String())
	}
	if errOut.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", errOut.String())
	}
}

func TestValidateConfigRequiresConfigFile(t *testing.T) {
	var out, errOut bytes.Buffer
	code := NewApp(&out, &errOut).Run([]string{"validate-config", "--config", "missing-test-config.yaml"})
	if code != exitConfigError {
		t.Fatalf("Run(validate-config missing) = %d, want %d; stderr=%q", code, exitConfigError, errOut.String())
	}
	if !strings.Contains(errOut.String(), "run opc-ua-cli init-config") {
		t.Fatalf("stderr = %q", errOut.String())
	}
}

func TestValidateConfigAcceptsGlobalConfigFlag(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte(`endpoint: opc.tcp://localhost:4840
policy: None
mode: None
timeout: 10s
`), 0o600); err != nil {
		t.Fatal(err)
	}

	var out, errOut bytes.Buffer
	code := NewApp(&out, &errOut).Run([]string{"--config", path, "validate-config"})
	if code != exitSuccess {
		t.Fatalf("Run(global config validate-config) = %d, want %d; stderr=%q", code, exitSuccess, errOut.String())
	}
	if !strings.Contains(out.String(), "PASS") {
		t.Fatalf("stdout = %q", out.String())
	}
}

func TestNormaliseGlobalFlagsPreservesCommandOverride(t *testing.T) {
	got, err := normaliseGlobalFlags([]string{"--format", "json", "read", "--format", "text", "--node", "i=2258"})
	if err != nil {
		t.Fatalf("normaliseGlobalFlags returned error: %v", err)
	}
	want := []string{"read", "--format", "json", "--format", "text", "--node", "i=2258"}
	if strings.Join(got, "\x00") != strings.Join(want, "\x00") {
		t.Fatalf("normalised args = %#v, want %#v", got, want)
	}
}
