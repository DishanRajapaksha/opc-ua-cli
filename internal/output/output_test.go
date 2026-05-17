package output

import "testing"

func TestNormaliseFormatSupportsJSONL(t *testing.T) {
	if got := NormaliseFormat("jsonl"); got != FormatJSONL {
		t.Fatalf("NormaliseFormat(jsonl) = %q", got)
	}
	if got := NormaliseFormat("JSONL"); got != FormatJSONL {
		t.Fatalf("NormaliseFormat(JSONL) = %q", got)
	}
}
