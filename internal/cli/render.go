package cli

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
