package output

import (
	shared "github.com/DishanRajapaksha/industrial-cli-kit/output"
	"io"
)

var ErrOutput = shared.ErrOutput

const (
	FormatTable = shared.FormatTable
	FormatText  = shared.FormatText
	FormatJSON  = shared.FormatJSON
	FormatJSONL = shared.FormatJSONL
	FormatCSV   = shared.FormatCSV
)

func NormaliseFormat(v string) string                        { return shared.NormaliseFormat(v) }
func WriteJSON(w io.Writer, v interface{}) error             { return shared.WriteJSON(w, v) }
func WriteJSONLine(w io.Writer, v interface{}) error         { return shared.WriteJSONLine(w, v) }
func WriteTable(w io.Writer, h []string, r [][]string) error { return shared.WriteTable(w, h, r) }
func WriteText(w io.Writer, v interface{}) error             { return shared.WriteText(w, v) }
func WriteCSV(w io.Writer, h []string, r [][]string) error   { return shared.WriteCSV(w, h, r) }
func WriteCSVRows(w io.Writer, r [][]string) error           { return shared.WriteCSVRows(w, r) }
