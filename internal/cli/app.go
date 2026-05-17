package cli

import (
	"fmt"
	"io"
	"os"
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
	if len(args) == 0 || args[0] == "help" || args[0] == "--help" || args[0] == "-h" {
		a.printUsage()
		return 0
	}

	var err error
	switch args[0] {
	case "endpoints", "status":
		err = a.endpoints(args[1:])
	case "browse":
		err = a.browse(args[1:])
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
	default:
		a.printUsage()
		fmt.Fprintf(a.err, "unknown command %q\n", args[0])
		return exitGeneralError
	}

	if err != nil {
		fmt.Fprintln(a.err, err)
		return mapExitCode(err)
	}

	return exitSuccess
}

func (a *App) printUsage() {
	fmt.Fprintln(a.out, `opc-ua-cli is a small OPC UA command-line client.

Usage:
  opc-ua-cli endpoints --profile local
  opc-ua-cli browse --profile local --node i=84 --depth 1
  opc-ua-cli read --profile site-a --node 'ns=2;s=Demo.Static.Scalar.Int32'
  opc-ua-cli write --profile site-a --node 'ns=2;s=Demo.Static.Scalar.Int32' --type int32 --value 42
  opc-ua-cli monitor --profile site-a --node 'ns=2;s=Demo.Static.Scalar.Int32' --interval 1s
  opc-ua-cli watch --profile site-a --node 'ns=2;s=Demo.Static.Scalar.Int32' --interval 1s
  opc-ua-cli alarms --profile site-a --node i=2253 --min-severity 500 --interval 1s
  opc-ua-cli test-connection --profile site-a
  opc-ua-cli init-config
  opc-ua-cli init-config --output site-a.yaml
  opc-ua-cli init-config --force

Commands:
  endpoints, status   List server endpoints and security options
  browse              Browse child nodes
  read                Read a node value
  write               Write a scalar node value
  monitor             Subscribe to data changes
  watch               Poll node values without subscriptions
  alarms              Subscribe to OPC UA alarm and event notifications
  test-connection     Run connection/auth/security diagnostics
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

CLI flags override values loaded from --config and --profile.`)
}
