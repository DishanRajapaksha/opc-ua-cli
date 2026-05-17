package cli

import (
	"fmt"
	"io"
	"os"
)

const exitFailure = 1

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
	case "alarms":
		err = a.alarms(args[1:])
	default:
		a.printUsage()
		fmt.Fprintf(a.err, "unknown command %q\n", args[0])
		return exitFailure
	}

	if err != nil {
		fmt.Fprintln(a.err, err)
		return exitFailure
	}

	return 0
}

func (a *App) printUsage() {
	fmt.Fprintln(a.out, `opc-ua-cli is a small OPC UA command-line client.

Usage:
  opc-ua-cli endpoints --config config.yaml
  opc-ua-cli browse --config config.yaml --node i=84 --depth 1
  opc-ua-cli read --config config.yaml --node 'ns=2;s=Demo.Static.Scalar.Int32'
  opc-ua-cli write --config config.yaml --node 'ns=2;s=Demo.Static.Scalar.Int32' --type int32 --value 42
  opc-ua-cli monitor --config config.yaml --node 'ns=2;s=Demo.Static.Scalar.Int32' --interval 1s
  opc-ua-cli alarms --config config.yaml --node i=2253 --min-severity 500 --interval 1s

Commands:
  endpoints, status   List server endpoints and security options
  browse              Browse child nodes
  read                Read a node value
  write               Write a scalar node value
  monitor             Subscribe to data changes
  alarms              Subscribe to OPC UA alarm and event notifications

Common flags:
  --config     YAML config file
  --endpoint   OPC UA endpoint URL
  --policy     Security policy
  --mode       Security mode
  --username   Username authentication
  --password   Password authentication
  --cert       Client certificate file
  --key        Client private key file
  --timeout    Request timeout
  --format     table, text, or json

CLI flags override values loaded from --config.`)
}
