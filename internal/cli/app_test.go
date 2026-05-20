package cli

import (
	"bytes"
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
