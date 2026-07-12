package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/DishanRajapaksha/industrial-cli-kit/command"
)

type App struct {
	out io.Writer
	err io.Writer
}

func NewApp(out io.Writer, err io.Writer) *App {
	return &App{out: out, err: err}
}

func Main() {
	code := NewApp(os.Stdout, os.Stderr).Run(os.Args[1:])
	if code != 0 {
		os.Exit(code)
	}
}

func (a *App) Run(args []string) int {
	normalisedArgs, err := normaliseGlobalFlags(args)
	if err != nil {
		fmt.Fprintln(a.err, err)
		return exitConfigError
	}
	args = normalisedArgs

	if len(args) == 0 || args[0] == "help" || args[0] == "--help" || args[0] == "-h" {
		a.printUsage()
		return 0
	}

	err = nil
	switch args[0] {
	case "version", "--version", "-v":
		fmt.Fprintln(a.out, "opc-ua-cli development")
	case "endpoints":
		err = a.endpoints(args[1:])
	case "status":
		err = a.status(args[1:])
	case "namespaces":
		err = a.namespaces(args[1:])
	case "browse":
		err = a.browse(args[1:])
	case "tui":
		err = a.tui(args[1:])
	case "attributes":
		err = a.attributes(args[1:])
	case "read":
		err = a.read(args[1:])
	case "write":
		err = a.write(args[1:])
	case "monitor":
		err = a.monitor(args[1:])
	case "watch":
		err = a.watch(args[1:])
	case "alarms":
		err = a.alarms(args[1:])
	case "test-connection":
		err = a.testConnection(args[1:])
	case "init-config":
		err = a.initConfig(args[1:])
	case "validate-config":
		err = a.validateConfig(args[1:])
	case "completions":
		err = a.completions(args[1:])
	default:
		a.printUsage()
		fmt.Fprintf(a.err, "unknown command %q\n", args[0])
		return exitGeneralError
	}

	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return exitSuccess
		}
		fmt.Fprintln(a.err, err)
		return mapExitCode(err)
	}

	return exitSuccess
}

func (a *App) printUsage() {
	a.writeRegistryUsage()
}

func normaliseGlobalFlags(args []string) ([]string, error) {
	return command.NormalizeGlobalFlagsForRegistry(args, cliRegistry)
}
