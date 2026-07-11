package cli

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
