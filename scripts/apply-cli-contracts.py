from pathlib import Path


def function_bounds(text: str, name: str) -> tuple[int, int]:
    start = text.find(f"func {name}(")
    if start < 0:
        raise SystemExit(f"function {name} not found")
    end = text.find("\nfunc ", start + 1)
    if end < 0:
        end = len(text)
    return start, end


def insert_after_apply(text: str, name: str) -> str:
    start, end = function_bounds(text, name)
    segment = text[start:end]
    marker = '''\tif err := common.applyConfig(fs); err != nil {
\t\treturn err
\t}
'''
    replacement = marker + '''\tif err := validateSnapshotFormat(common.format); err != nil {
\t\treturn err
\t}
'''
    if segment.count(marker) != 1:
        raise SystemExit(f"{name}: applyConfig marker count {segment.count(marker)}")
    segment = segment.replace(marker, replacement, 1)
    return text[:start] + segment + text[end:]


Path("internal/output/output.go").write_text(r'''package output

import (
    "encoding/csv"
    "encoding/json"
    "errors"
    "fmt"
    "io"
    "strings"
    "text/tabwriter"
)

var ErrOutput = errors.New("output error")

const (
    FormatTable = "table"
    FormatText  = "text"
    FormatJSON  = "json"
    FormatJSONL = "jsonl"
    FormatCSV   = "csv"
)

func NormaliseFormat(value string) string {
    switch strings.ToLower(strings.TrimSpace(value)) {
    case "", FormatTable:
        return FormatTable
    case FormatText:
        return FormatText
    case FormatJSON:
        return FormatJSON
    case FormatJSONL:
        return FormatJSONL
    case FormatCSV:
        return FormatCSV
    default:
        return value
    }
}

func WriteJSON(w io.Writer, value interface{}) error {
    encoder := json.NewEncoder(w)
    encoder.SetIndent("", "  ")
    if err := encoder.Encode(value); err != nil {
        return fmt.Errorf("%w: %v", ErrOutput, err)
    }
    return nil
}

func WriteJSONLine(w io.Writer, value interface{}) error {
    if err := json.NewEncoder(w).Encode(value); err != nil {
        return fmt.Errorf("%w: %v", ErrOutput, err)
    }
    return nil
}

func WriteTable(w io.Writer, headers []string, rows [][]string) error {
    tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
    fmt.Fprintln(tw, strings.Join(headers, "\t"))
    for _, row := range rows {
        fmt.Fprintln(tw, strings.Join(row, "\t"))
    }
    if err := tw.Flush(); err != nil {
        return fmt.Errorf("%w: %v", ErrOutput, err)
    }
    return nil
}

func WriteText(w io.Writer, value interface{}) error {
    if _, err := fmt.Fprintln(w, value); err != nil {
        return fmt.Errorf("%w: %v", ErrOutput, err)
    }
    return nil
}

func WriteCSV(w io.Writer, headers []string, rows [][]string) error {
    cw := csv.NewWriter(w)
    if len(headers) > 0 {
        if err := cw.Write(headers); err != nil {
            return fmt.Errorf("%w: %v", ErrOutput, err)
        }
    }
    if err := cw.WriteAll(rows); err != nil {
        return fmt.Errorf("%w: %v", ErrOutput, err)
    }
    if err := cw.Error(); err != nil {
        return fmt.Errorf("%w: %v", ErrOutput, err)
    }
    return nil
}

func WriteCSVRows(w io.Writer, rows [][]string) error {
    return WriteCSV(w, nil, rows)
}
''')

Path("internal/cli/exit_codes.go").write_text(r'''package cli

import (
    "context"
    "errors"
    "flag"
    "strings"

    "github.com/DishanRajapaksha/opc-ua-cli/internal/config"
    "github.com/DishanRajapaksha/opc-ua-cli/internal/output"
    "github.com/DishanRajapaksha/opc-ua-cli/internal/uaclient"
)

const (
    exitSuccess         = 0
    exitGeneralError    = 1
    exitConfigError     = 2
    exitConnection      = 3
    exitProtocolRequest = 4
    exitAuthSecurity    = 5
    exitNodeNotFound    = 6
    exitWriteReject     = 7
    exitTimeout         = 8
    exitOutputError     = 9
)

func mapExitCode(err error) int {
    switch {
    case err == nil:
        return exitSuccess
    case errors.Is(err, flag.ErrHelp):
        return exitSuccess
    case isFlagParseError(err):
        return exitConfigError
    case errors.Is(err, config.ErrConfig), errors.Is(err, uaclient.ErrValidation):
        return exitConfigError
    case errors.Is(err, context.DeadlineExceeded), strings.Contains(strings.ToLower(err.Error()), "timeout"):
        return exitTimeout
    case errors.Is(err, output.ErrOutput):
        return exitOutputError
    case errors.Is(err, uaclient.ErrAuthSecurity):
        return exitAuthSecurity
    case errors.Is(err, uaclient.ErrNodeNotFound):
        return exitNodeNotFound
    case errors.Is(err, uaclient.ErrBadStatusCode):
        return exitProtocolRequest
    case errors.Is(err, uaclient.ErrWriteRejected):
        return exitWriteReject
    case errors.Is(err, uaclient.ErrConnection):
        return exitConnection
    default:
        return exitGeneralError
    }
}

func isFlagParseError(err error) bool {
    msg := err.Error()
    return strings.Contains(msg, "flag provided but not defined") ||
        strings.Contains(msg, "flag needs an argument") ||
        strings.Contains(msg, "invalid value")
}
''')

Path("internal/cli/formats.go").write_text(r'''package cli

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
''')

Path("internal/cli/render.go").write_text(r'''package cli

import (
    "fmt"
    "strings"

    "github.com/DishanRajapaksha/opc-ua-cli/internal/domain"
    "github.com/DishanRajapaksha/opc-ua-cli/internal/output"
)

func (a *App) renderEndpoints(format string, rows []domain.Endpoint) error {
    tableRows := make([][]string, 0, len(rows))
    for _, row := range rows {
        tableRows = append(tableRows, []string{row.EndpointURL, row.SecurityPolicy, row.SecurityMode, strings.Join(row.UserTokens, ",")})
    }
    headers := []string{"Endpoint", "Policy", "Mode", "User tokens"}
    switch output.NormaliseFormat(format) {
    case output.FormatJSON:
        return output.WriteJSON(a.out, rows)
    case output.FormatCSV:
        return output.WriteCSV(a.out, headers, tableRows)
    case output.FormatTable, output.FormatText:
        return output.WriteTable(a.out, headers, tableRows)
    default:
        return validateSnapshotFormat(format)
    }
}

func (a *App) renderNamespaces(format string, rows [][]string) error {
    headers := []string{"Index", "URI"}
    switch output.NormaliseFormat(format) {
    case output.FormatJSON:
        values := make([]map[string]string, 0, len(rows))
        for _, row := range rows {
            values = append(values, map[string]string{"index": row[0], "uri": row[1]})
        }
        return output.WriteJSON(a.out, values)
    case output.FormatCSV:
        return output.WriteCSV(a.out, headers, rows)
    case output.FormatTable, output.FormatText:
        return output.WriteTable(a.out, headers, rows)
    default:
        return validateSnapshotFormat(format)
    }
}

func (a *App) renderNodes(format string, rows []domain.Node) error {
    tableRows := make([][]string, 0, len(rows))
    for _, row := range rows {
        tableRows = append(tableRows, []string{row.NodeID, row.NodeClass, row.BrowseName, row.DataType, fmt.Sprint(row.Writable), row.Path})
    }
    headers := []string{"NodeID", "Class", "Browse name", "Data type", "Writable", "Path"}
    switch output.NormaliseFormat(format) {
    case output.FormatJSON:
        return output.WriteJSON(a.out, rows)
    case output.FormatCSV:
        return output.WriteCSV(a.out, headers, tableRows)
    case output.FormatTable, output.FormatText:
        return output.WriteTable(a.out, headers, tableRows)
    default:
        return validateSnapshotFormat(format)
    }
}

func readHeaders() []string {
    return []string{"NodeID", "Status", "Value", "Source timestamp", "Server timestamp"}
}

func readRow(row domain.ReadResult) []string {
    return []string{row.NodeID, row.Status, fmt.Sprint(row.Value), row.SourceTimestamp, row.ServerTimestamp}
}

func (a *App) renderRead(format string, row domain.ReadResult) error {
    switch output.NormaliseFormat(format) {
    case output.FormatJSON:
        return output.WriteJSON(a.out, row)
    case output.FormatText:
        return output.WriteText(a.out, row.Value)
    case output.FormatCSV:
        return output.WriteCSV(a.out, readHeaders(), [][]string{readRow(row)})
    case output.FormatTable:
        return output.WriteTable(a.out, readHeaders(), [][]string{readRow(row)})
    default:
        return validateSnapshotFormat(format)
    }
}

func (a *App) renderWrite(format string, row domain.WriteResult) error {
    headers := []string{"NodeID", "Status"}
    values := [][]string{{row.NodeID, row.Status}}
    switch output.NormaliseFormat(format) {
    case output.FormatJSON:
        return output.WriteJSON(a.out, row)
    case output.FormatText:
        return output.WriteText(a.out, row.Status)
    case output.FormatCSV:
        return output.WriteCSV(a.out, headers, values)
    case output.FormatTable:
        return output.WriteTable(a.out, headers, values)
    default:
        return validateSnapshotFormat(format)
    }
}

func (a *App) renderAttributes(format string, row domain.NodeAttributesResult) error {
    tableRows := make([][]string, 0, len(row.Attributes))
    for _, attr := range row.Attributes {
        tableRows = append(tableRows, []string{row.NodeID, attr.Name, fmt.Sprint(attr.Value), attr.Status})
    }
    headers := []string{"NodeID", "Attribute", "Value", "Status"}
    switch output.NormaliseFormat(format) {
    case output.FormatJSON:
        return output.WriteJSON(a.out, row)
    case output.FormatText:
        if _, err := fmt.Fprintf(a.out, "NodeID: %s\n", row.NodeID); err != nil {
            return fmt.Errorf("%w: %v", output.ErrOutput, err)
        }
        for _, attr := range row.Attributes {
            if _, err := fmt.Fprintf(a.out, "%s: %v (%s)\n", attr.Name, attr.Value, attr.Status); err != nil {
                return fmt.Errorf("%w: %v", output.ErrOutput, err)
            }
        }
        return nil
    case output.FormatCSV:
        return output.WriteCSV(a.out, headers, tableRows)
    case output.FormatTable:
        return output.WriteTable(a.out, headers, tableRows)
    default:
        return validateSnapshotFormat(format)
    }
}

func renderReadMany(a *App, format string, rows []domain.ReadResult) error {
    values := make([][]string, 0, len(rows))
    for _, row := range rows {
        values = append(values, readRow(row))
    }
    switch output.NormaliseFormat(format) {
    case output.FormatJSON:
        return output.WriteJSON(a.out, rows)
    case output.FormatText:
        for _, row := range rows {
            if err := output.WriteText(a.out, row.Value); err != nil {
                return err
            }
        }
        return nil
    case output.FormatCSV:
        return output.WriteCSV(a.out, readHeaders(), values)
    case output.FormatTable:
        return output.WriteTable(a.out, readHeaders(), values)
    default:
        return validateSnapshotFormat(format)
    }
}

func renderWriteMany(a *App, format string, rows []domain.WriteResult) error {
    headers := []string{"NodeID", "Status"}
    values := make([][]string, 0, len(rows))
    for _, row := range rows {
        values = append(values, []string{row.NodeID, row.Status})
    }
    switch output.NormaliseFormat(format) {
    case output.FormatJSON:
        return output.WriteJSON(a.out, rows)
    case output.FormatText:
        for _, row := range rows {
            if err := output.WriteText(a.out, row.Status); err != nil {
                return err
            }
        }
        return nil
    case output.FormatCSV:
        return output.WriteCSV(a.out, headers, values)
    case output.FormatTable:
        return output.WriteTable(a.out, headers, values)
    default:
        return validateSnapshotFormat(format)
    }
}

func dataChangeHeaders() []string {
    return []string{"Source timestamp", "NodeID", "Value"}
}

func (a *App) renderDataChange(format string, row domain.DataChange) error {
    switch output.NormaliseFormat(format) {
    case output.FormatJSONL:
        return output.WriteJSONLine(a.out, row)
    case output.FormatCSV:
        return output.WriteCSVRows(a.out, [][]string{{row.SourceTimestamp, row.NodeID, fmt.Sprint(row.Value)}})
    case output.FormatText:
        if _, err := fmt.Fprintln(a.out, row.SourceTimestamp, row.NodeID, row.Value); err != nil {
            return fmt.Errorf("%w: %v", output.ErrOutput, err)
        }
        return nil
    default:
        return validateStreamFormat(format)
    }
}

func alarmHeaders() []string {
    return []string{"Time", "Severity", "Source", "Message", "Event type", "Event ID"}
}

func (a *App) renderAlarmEvent(format string, row domain.AlarmEvent) error {
    source := firstNonEmpty(row.SourceName, row.SourceNode, row.NodeID)
    switch output.NormaliseFormat(format) {
    case output.FormatJSONL:
        return output.WriteJSONLine(a.out, row)
    case output.FormatCSV:
        return output.WriteCSVRows(a.out, [][]string{{row.Time, fmt.Sprint(row.Severity), source, row.Message, row.EventType, row.EventID}})
    case output.FormatText:
        if _, err := fmt.Fprintf(a.out, "%s severity=%d source=%s message=%s eventType=%s eventId=%s\n", row.Time, row.Severity, source, row.Message, row.EventType, row.EventID); err != nil {
            return fmt.Errorf("%w: %v", output.ErrOutput, err)
        }
        return nil
    default:
        return validateStreamFormat(format)
    }
}

func firstNonEmpty(values ...string) string {
    for _, value := range values {
        if value != "" {
            return value
        }
    }
    return ""
}
''')

commands = Path("internal/cli/commands.go")
text = commands.read_text()
snapshot_token = "__SNAPSHOT_FORMAT_HELP__"
stream_token = "__STREAM_FORMAT_HELP__"
for old in (
    "output format: table, text, json, or jsonl",
    "output format: table, json, or jsonl",
    "output format: table, text, or json",
    "output format: table or json",
    "output format: table, json",
    "output format: table",
):
    text = text.replace(old, snapshot_token)
text = text.replace("output format: text or jsonl", stream_token)
text = text.replace(snapshot_token, "output format: table, text, json, or csv")
text = text.replace(stream_token, "output format: text, jsonl, or csv")

for name in (
    "(a *App) status",
    "(a *App) namespaces",
    "(a *App) attributes",
    "(a *App) browse",
    "(a *App) read",
    "(a *App) write",
):
    text = insert_after_apply(text, name)

start, end = function_bounds(text, "(a *App) endpoints")
segment = text[start:end]
marker = '''\tif err := fs.Parse(args); err != nil {
\t\treturn err
\t}
'''
replacement = marker + '''\tif err := validateSnapshotFormat(format); err != nil {
\t\treturn err
\t}
'''
if segment.count(marker) != 1:
    raise SystemExit("endpoints parse marker not unique")
segment = segment.replace(marker, replacement, 1)
text = text[:start] + segment + text[end:]

old = 'return output.WriteTable(a.out, []string{"Index", "URI"}, rows)'
if text.count(old) != 1:
    raise SystemExit("namespace table return not found")
text = text.replace(old, 'return a.renderNamespaces(common.format, rows)', 1)

start, end = function_bounds(text, "validateStreamFormat")
text = text[:start] + text[end:]

for name, header_call in (
    ("(a *App) monitor", "dataChangeHeaders()"),
    ("(a *App) watch", "dataChangeHeaders()"),
    ("(a *App) alarms", "alarmHeaders()"),
):
    start, end = function_bounds(text, name)
    segment = text[start:end]
    marker = '\tformat := output.NormaliseFormat(common.format)\n'
    addition = marker + f'''\tif format == output.FormatCSV {{
\t\tif err := output.WriteCSV(a.out, {header_call}, nil); err != nil {{
\t\t\treturn err
\t\t}}
\t}}
'''
    if segment.count(marker) != 1:
        raise SystemExit(f"{name}: format marker count {segment.count(marker)}")
    segment = segment.replace(marker, addition, 1)
    text = text[:start] + segment + text[end:]

commands.write_text(text)

app = Path("internal/cli/app.go")
text = app.read_text().replace(
    "  --format     table, text, json, or jsonl",
    "  --format     snapshots: table, text, json, csv; streams: text, jsonl, csv",
)
app.write_text(text)

readme = Path("README.md")
text = readme.read_text()
text = text.replace(
    '''Commands support formats where applicable:

- `table` (default)
- `text`
- `json`
- `jsonl`

Stream commands (`monitor`, `watch`, and `alarms`) use `jsonl` for machine-readable event output. They do not use `json`, because streams are not a single complete JSON document.''',
    '''Snapshot commands support:

- `table` (default)
- `text`
- `json`
- `csv`

Stream commands (`monitor`, `watch`, and `alarms`) support:

- `text` (default)
- `jsonl`
- `csv`

Streams reject `json` and `table`, because an unbounded stream is neither one complete JSON document nor a finite table.''',
    1,
)
text = text.replace(
    '''- `0`: success
- `1`: general error
- `2`: config error
- `3`: connection error
- `4`: authentication/security error
- `5`: node not found
- `6`: bad OPC UA status code
- `7`: write rejected''',
    '''- `0`: success
- `1`: general error
- `2`: usage or configuration error
- `3`: transport or connection error
- `4`: protocol or request error
- `5`: authentication or security error
- `6`: node not found
- `7`: write or control rejected
- `8`: operation timeout
- `9`: output or formatting error''',
    1,
)
readme.write_text(text)

Path("internal/cli/contracts_test.go").write_text(r'''package cli

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
''')
