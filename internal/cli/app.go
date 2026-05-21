package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
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
	fmt.Fprintln(a.out, `opc-ua-cli is a small OPC UA command-line client.

Usage:
  opc-ua-cli [global flags] <command> [flags]
  opc-ua-cli endpoints --profile local
  opc-ua-cli namespaces --profile local
  opc-ua-cli browse --profile local --node i=84 --depth 1
  opc-ua-cli tui --profile local --node i=84 --interval 1s
  opc-ua-cli attributes --profile local --node 'ns=2;s=Demo.Value'
  opc-ua-cli read --profile site-a --node 'ns=2;s=Demo.Static.Scalar.Int32'
  opc-ua-cli write --profile site-a --node 'ns=2;s=Demo.Static.Scalar.Int32' --type int32 --value 42
  opc-ua-cli monitor --profile site-a --node 'ns=2;s=Demo.Static.Scalar.Int32' --interval 1s
  opc-ua-cli watch --profile site-a --node 'ns=2;s=Demo.Static.Scalar.Int32' --interval 1s
  opc-ua-cli alarms --profile site-a --node i=2253 --min-severity 500 --interval 1s
  opc-ua-cli test-connection --profile site-a
  opc-ua-cli validate-config --profile site-a
  opc-ua-cli completions zsh
  opc-ua-cli init-config
  opc-ua-cli init-config --output site-a.yaml
  opc-ua-cli init-config --force
  opc-ua-cli version

Commands:
  version            Print version information
  endpoints           List server endpoints and security options
  status              Read server status
  namespaces          List namespace indexes and URIs
  browse              Browse child nodes
  tui                 Browse nodes interactively
  attributes          Inspect node metadata attributes
  read                Read a node value
  write               Write a scalar node value
  monitor             Subscribe to data changes
  watch               Poll node values without subscriptions
  alarms              Subscribe to OPC UA alarm and event notifications
  test-connection     Run connection/auth/security diagnostics
  validate-config     Validate local config without server connection
  completions         Generate shell completion scripts
  init-config         Write a starter YAML config file

Common flags:
  --config     YAML config file, defaults to config.yaml
  --profile    Config profile name; uses default_profile when omitted
  --endpoint   OPC UA endpoint URL
  --policy     Security policy
  --mode       Security mode
  --username   Username authentication
  --password   Password authentication
  --cert       Client certificate file
  --key        Client private key file
  --timeout    Request timeout
  --format     table, text, json, or jsonl
  --verbose    Print high-level connection decisions
  --debug      Enable lower-level OPC UA client debug logging

CLI flags override values loaded from --config and --profile.`)
}

func normaliseGlobalFlags(args []string) ([]string, error) {
	if len(args) == 0 {
		return args, nil
	}

	var globals []string
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--" {
			if i+1 >= len(args) {
				return nil, errors.New("command is required after --")
			}
			return appendCommandGlobals(args[i+1:], globals), nil
		}
		if !strings.HasPrefix(arg, "-") || arg == "-" {
			return appendCommandGlobals(args[i:], globals), nil
		}
		if arg == "--help" || arg == "-h" || arg == "--version" || arg == "-v" {
			return args[i:], nil
		}

		name, inlineValue, hasInlineValue := strings.Cut(arg, "=")
		switch name {
		case "--verbose", "--debug":
			if hasInlineValue {
				return nil, fmt.Errorf("%s does not take a value", name)
			}
			globals = append(globals, name)
		case "--config", "--profile", "--endpoint", "--policy", "--mode", "--username", "--password", "--cert", "--key", "--timeout", "--format":
			value := inlineValue
			if !hasInlineValue {
				i++
				if i >= len(args) || strings.HasPrefix(args[i], "-") {
					return nil, fmt.Errorf("%s requires a value", name)
				}
				value = args[i]
			}
			if value == "" {
				return nil, fmt.Errorf("%s requires a value", name)
			}
			globals = append(globals, name, value)
		default:
			return nil, fmt.Errorf("unknown global flag %q", name)
		}
	}

	return nil, errors.New("command is required")
}

func appendCommandGlobals(args []string, globals []string) []string {
	if len(args) == 0 || len(globals) == 0 {
		return args
	}
	command := args[0]
	filteredGlobals := filterGlobalsForCommand(command, globals)
	if len(filteredGlobals) == 0 {
		return args
	}
	out := make([]string, 0, len(args)+len(filteredGlobals))
	out = append(out, command)
	out = append(out, filteredGlobals...)
	out = append(out, args[1:]...)
	return out
}

func filterGlobalsForCommand(command string, globals []string) []string {
	out := make([]string, 0, len(globals))
	for i := 0; i < len(globals); i++ {
		name := globals[i]
		if !commandSupportsGlobalFlag(command, name) {
			if globalFlagTakesValue(name) {
				i++
			}
			continue
		}
		out = append(out, name)
		if globalFlagTakesValue(name) {
			i++
			if i < len(globals) {
				out = append(out, globals[i])
			}
		}
	}
	return out
}

func commandSupportsGlobalFlag(command string, name string) bool {
	switch command {
	case "validate-config":
		switch name {
		case "--config", "--profile", "--verbose", "--debug":
			return true
		default:
			return false
		}
	}
	switch command {
	case "endpoints", "status", "namespaces", "browse", "tui", "attributes", "read", "write", "monitor", "watch", "alarms", "test-connection":
		return true
	default:
		return false
	}
}

func globalFlagTakesValue(name string) bool {
	switch name {
	case "--verbose", "--debug":
		return false
	default:
		return true
	}
}
