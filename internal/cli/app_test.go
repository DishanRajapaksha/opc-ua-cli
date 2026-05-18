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
