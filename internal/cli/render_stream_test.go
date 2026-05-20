package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/DishanRajapaksha/opc-ua-cli/internal/domain"
)

func TestRenderDataChangeJSONL(t *testing.T) {
	var out bytes.Buffer
	app := NewApp(&out, &bytes.Buffer{})
	row := domain.DataChange{NodeID: "ns=2;s=Demo.Value", Value: 42}

	if err := app.renderDataChange("jsonl", row); err != nil {
		t.Fatalf("renderDataChange returned error: %v", err)
	}

	line := strings.TrimSpace(out.String())
	if !strings.Contains(line, `"nodeId":"ns=2;s=Demo.Value"`) {
		t.Fatalf("unexpected output: %q", line)
	}
	if !strings.Contains(line, `"value":42`) {
		t.Fatalf("unexpected output: %q", line)
	}
}

func TestRenderAlarmEventJSONL(t *testing.T) {
	var out bytes.Buffer
	app := NewApp(&out, &bytes.Buffer{})
	row := domain.AlarmEvent{NodeID: "i=2253", Severity: 600, Message: "High temperature"}

	if err := app.renderAlarmEvent("jsonl", row); err != nil {
		t.Fatalf("renderAlarmEvent returned error: %v", err)
	}

	line := strings.TrimSpace(out.String())
	if !strings.Contains(line, `"nodeId":"i=2253"`) {
		t.Fatalf("unexpected output: %q", line)
	}
	if !strings.Contains(line, `"severity":600`) {
		t.Fatalf("unexpected output: %q", line)
	}
}

func TestValidateStreamFormatRejectsJSON(t *testing.T) {
	if err := validateStreamFormat("json"); err == nil {
		t.Fatal("expected json stream format to be rejected")
	}
	if err := validateStreamFormat("jsonl"); err != nil {
		t.Fatalf("validateStreamFormat(jsonl) returned error: %v", err)
	}
}
