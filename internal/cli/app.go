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
	case "endpoints":
		err = a.endpoints(args[1:])
	case "browse":
		err = a.browse(args[1:])
	case "read":
		err = a.read(args[1:])
	case "write":
		err = a.write(args[1:])
	case "monitor":
		err = a.monitor(args[1:])
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
  opc-ua-cli endpoints --endpoint opc.tcp://localhost:4840
  opc-ua-cli browse    --endpoint opc.tcp://localhost:4840 --node i=84 --depth 1
  opc-ua-cli read      --endpoint opc.tcp://localhost:4840 --node 'ns=2;s=Demo.Static.Scalar.Int32'
  opc-ua-cli write     --endpoint opc.tcp://localhost:4840 --node 'ns=2;s=Demo.Static.Scalar.Int32' --type int32 --value 42
  opc-ua-cli monitor   --endpoint opc.tcp://localhost:4840 --node 'ns=2;s=Demo.Static.Scalar.Int32' --interval 1s

Commands:
  endpoints   List server endpoints and security options
  browse      Browse child nodes
  read        Read a node value
  write       Write a scalar node value
  monitor     Subscribe to data changes

Common flags:
  --endpoint   OPC UA endpoint URL
  --policy     Security policy: None, Basic128Rsa15, Basic256, Basic256Sha256
  --mode       Security mode: None, Sign, SignAndEncrypt
  --username   Username authentication
  --password   Password authentication
  --cert       Client certificate file
  --key        Client private key file
  --timeout    Request timeout, for example 10s
  --format     table, text, or json`)
}
