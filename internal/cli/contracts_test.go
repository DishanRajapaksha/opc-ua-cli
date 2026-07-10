package cli

import (
	"context"
	"testing"

	"github.com/DishanRajapaksha/opc-ua-cli/internal/output"
	"github.com/DishanRajapaksha/opc-ua-cli/internal/uaclient"
)

func TestSnapshotFormatContract(t *testing.T) {
	for _, format := range []string{"table", "text", "json", "csv"} {
		if err := validateSnapshotFormat(format); err != nil {
			t.Fatalf("snapshot format %q rejected: %v", format, err)
		}
	}
	if err := validateSnapshotFormat("jsonl"); err == nil {
		t.Fatal("snapshot commands must reject jsonl")
	}
}

func TestStreamFormatContract(t *testing.T) {
	for _, format := range []string{"text", "jsonl", "csv"} {
		if err := validateStreamFormat(format); err != nil {
			t.Fatalf("stream format %q rejected: %v", format, err)
		}
	}
	for _, format := range []string{"table", "json"} {
		if err := validateStreamFormat(format); err == nil {
			t.Fatalf("stream format %q must be rejected", format)
		}
	}
}

func TestSharedExitCodeContract(t *testing.T) {
	cases := []struct {
		err  error
		want int
	}{
		{uaclient.ErrConnection, 3},
		{uaclient.ErrBadStatusCode, 4},
		{uaclient.ErrAuthSecurity, 5},
		{uaclient.ErrNodeNotFound, 6},
		{uaclient.ErrWriteRejected, 7},
		{context.DeadlineExceeded, 8},
		{output.ErrOutput, 9},
	}
	for _, tc := range cases {
		if got := mapExitCode(tc.err); got != tc.want {
			t.Fatalf("mapExitCode(%v) = %d, want %d", tc.err, got, tc.want)
		}
	}
}
