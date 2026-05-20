package cli

import (
	"errors"
	"flag"
	"strings"

	"github.com/DishanRajapaksha/opc-ua-cli/internal/config"
	"github.com/DishanRajapaksha/opc-ua-cli/internal/uaclient"
)

const (
	exitSuccess      = 0
	exitGeneralError = 1
	exitConfigError  = 2
	exitConnection   = 3
	exitAuthSecurity = 4
	exitNodeNotFound = 5
	exitBadStatus    = 6
	exitWriteReject  = 7
)

func mapExitCode(err error) int {
	switch {
	case err == nil:
		return exitSuccess
	case errors.Is(err, flag.ErrHelp):
		return exitSuccess
	case isFlagParseError(err):
		return exitConfigError
	case errors.Is(err, config.ErrConfig):
		return exitConfigError
	case errors.Is(err, uaclient.ErrAuthSecurity):
		return exitAuthSecurity
	case errors.Is(err, uaclient.ErrNodeNotFound):
		return exitNodeNotFound
	case errors.Is(err, uaclient.ErrBadStatusCode):
		return exitBadStatus
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
