package cli

import (
	"fmt"
	"strings"

	"github.com/DishanRajapaksha/opc-ua-cli/internal/domain"
	"github.com/DishanRajapaksha/opc-ua-cli/internal/output"
)

func (a *App) renderEndpoints(format string, rows []domain.Endpoint) error {
	if output.NormaliseFormat(format) == output.FormatJSON {
		return output.WriteJSON(a.out, rows)
	}

	tableRows := make([][]string, 0, len(rows))
	for _, row := range rows {
		tableRows = append(tableRows, []string{row.EndpointURL, row.SecurityPolicy, row.SecurityMode, strings.Join(row.UserTokens, ",")})
	}
	return output.WriteTable(a.out, []string{"Endpoint", "Policy", "Mode", "User tokens"}, tableRows)
}

func (a *App) renderNodes(format string, rows []domain.Node) error {
	if output.NormaliseFormat(format) == output.FormatJSON {
		return output.WriteJSON(a.out, rows)
	}

	tableRows := make([][]string, 0, len(rows))
	for _, row := range rows {
		tableRows = append(tableRows, []string{row.NodeID, row.NodeClass, row.BrowseName, row.DataType, fmt.Sprint(row.Writable), row.Path})
	}
	return output.WriteTable(a.out, []string{"NodeID", "Class", "Browse name", "Data type", "Writable", "Path"}, tableRows)
}

func (a *App) renderRead(format string, row domain.ReadResult) error {
	switch output.NormaliseFormat(format) {
	case output.FormatJSON:
		return output.WriteJSON(a.out, row)
	case output.FormatText:
		return output.WriteText(a.out, row.Value)
	default:
		return output.WriteTable(a.out, []string{"NodeID", "Status", "Value", "Source timestamp", "Server timestamp"}, [][]string{{row.NodeID, row.Status, fmt.Sprint(row.Value), row.SourceTimestamp, row.ServerTimestamp}})
	}
}

func (a *App) renderWrite(format string, row domain.WriteResult) error {
	if output.NormaliseFormat(format) == output.FormatJSON {
		return output.WriteJSON(a.out, row)
	}
	return output.WriteTable(a.out, []string{"NodeID", "Status"}, [][]string{{row.NodeID, row.Status}})
}

func (a *App) renderDataChange(format string, row domain.DataChange) error {
	if output.NormaliseFormat(format) == output.FormatJSON {
		return output.WriteJSONLine(a.out, row)
	}
	_, err := fmt.Fprintln(a.out, row.SourceTimestamp, row.NodeID, row.Value)
	return err
}
