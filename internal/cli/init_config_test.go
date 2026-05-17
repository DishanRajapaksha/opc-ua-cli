package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInitConfigWritesDefaultFile(t *testing.T) {
	wd := t.TempDir()
	previousWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	defer func() { _ = os.Chdir(previousWD) }()
	if err := os.Chdir(wd); err != nil {
		t.Fatalf("Chdir: %v", err)
	}

	var out bytes.Buffer
	var errOut bytes.Buffer
	app := NewApp(&out, &errOut)

	code := app.Run([]string{"init-config"})
	if code != 0 {
		t.Fatalf("Run returned code %d, stderr=%q", code, errOut.String())
	}

	path := filepath.Join(wd, "config.yaml")
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if !strings.Contains(string(contents), "endpoint:") {
		t.Fatalf("written config missing endpoint field")
	}
}

func TestInitConfigRefusesOverwriteWithoutForce(t *testing.T) {
	wd := t.TempDir()
	path := filepath.Join(wd, "site-a.yaml")
	if err := os.WriteFile(path, []byte("endpoint: existing"), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	var out bytes.Buffer
	var errOut bytes.Buffer
	app := NewApp(&out, &errOut)

	code := app.Run([]string{"init-config", "--output", path})
	if code == 0 {
		t.Fatalf("expected non-zero exit code")
	}
	if !strings.Contains(errOut.String(), "use --force") {
		t.Fatalf("stderr = %q", errOut.String())
	}
}

func TestInitConfigOverwritesWithForce(t *testing.T) {
	wd := t.TempDir()
	path := filepath.Join(wd, "site-a.yaml")
	if err := os.WriteFile(path, []byte("endpoint: existing"), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	var out bytes.Buffer
	var errOut bytes.Buffer
	app := NewApp(&out, &errOut)

	code := app.Run([]string{"init-config", "--output", path, "--force"})
	if code != 0 {
		t.Fatalf("Run returned code %d, stderr=%q", code, errOut.String())
	}

	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if strings.Contains(string(contents), "endpoint: existing") {
		t.Fatalf("expected file to be overwritten")
	}
}
