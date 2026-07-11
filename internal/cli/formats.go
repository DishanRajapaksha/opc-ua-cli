package cli

import (
	"fmt"

	"github.com/DishanRajapaksha/opc-ua-cli/internal/config"
	"github.com/DishanRajapaksha/opc-ua-cli/internal/output"
)

func validateSnapshotFormat(format string) error {
	switch output.NormaliseFormat(format) {
	case output.FormatTable, output.FormatText, output.FormatJSON, output.FormatCSV:
		return nil
	case output.FormatJSONL:
		return fmt.Errorf("%w: snapshot commands produce one result; use --format json instead of --format jsonl", config.ErrConfig)
	default:
		return fmt.Errorf("%w: invalid snapshot output format %q; expected table, text, json, or csv", config.ErrConfig, format)
	}
}

func validateStreamFormat(format string) error {
	switch output.NormaliseFormat(format) {
	case output.FormatText, output.FormatJSONL, output.FormatCSV:
		return nil
	case output.FormatJSON:
		return fmt.Errorf("%w: stream commands use line-delimited output; use --format jsonl instead of --format json", config.ErrConfig)
	case output.FormatTable:
		return fmt.Errorf("%w: stream commands do not support table output; use text, jsonl, or csv", config.ErrConfig)
	default:
		return fmt.Errorf("%w: invalid stream output format %q; expected text, jsonl, or csv", config.ErrConfig, format)
	}
}
