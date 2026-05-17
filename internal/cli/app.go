package cli

import (
	"fmt"
	"io"
	"os"
	"strings"
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

	if strings.HasPrefix(args[0], "-") {
		if err := a.root(args); err != nil {
			fmt.Fprintln(a.err, err)
			return exitFailure
		}
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

XML-DA compatible flag style:
  opc-ua-cli -endpoint opc.tcp://localhost:4840
  opc-ua-cli -endpoint opc.tcp://localhost:4840 -browse-path i=84 -browse-depth 1
  opc-ua-cli -endpoint opc.tcp://localhost:4840 -read-path 'ns=2;s=Demo.Static.Scalar.Int32'
  opc-ua-cli -endpoint opc.tcp://localhost:4840 -write-path 'ns=2;s=Demo.Static.Scalar.Int32' -write-type int32 -write-value 42
  opc-ua-cli -endpoint opc.tcp://localhost:4840 -monitor-path 'ns=2;s=Demo.Static.Scalar.Int32' -monitor-interval 1s

Subcommand style:
  opc-ua-cli endpoints --endpoint opc.tcp://localhost:4840
  opc-ua-cli browse    --endpoint opc.tcp://localhost:4840 --node i=84 --depth 1
  opc-ua-cli read      --endpoint opc.tcp://localhost:4840 --node 'ns=2;s=Demo.Static.Scalar.Int32'
  opc-ua-cli write     --endpoint opc.tcp://localhost:4840 --node 'ns=2;s=Demo.Static.Scalar.Int32' --type int32 --value 42
  opc-ua-cli monitor   --endpoint opc.tcp://localhost:4840 --node 'ns=2;s=Demo.Static.Scalar.Int32' --interval 1s

Root flags:
  -endpoint           OPC UA endpoint URL
  -browse-path        OPC UA node id to browse
  -browse-depth       Browse recursion depth
  -read-path          OPC UA node id to read
  -write-path         OPC UA node id to write
  -write-value        Value to write
  -write-type         Scalar value type
  -monitor-path       OPC UA node id to monitor
  -monitor-interval   Subscription interval
  -duration           Monitor duration; zero runs until interrupted
  -format             table, text, or json

Common connection flags:
  -policy / --policy     Security policy: None, Basic128Rsa15, Basic256, Basic256Sha256
  -mode / --mode         Security mode: None, Sign, SignAndEncrypt
  -username / --username Username authentication
  -password / --password Password authentication
  -cert / --cert         Client certificate file
  -key / --key           Client private key file
  -timeout / --timeout   Request timeout, for example 10s

Commands:
  endpoints, status   List server endpoints and security options
  browse              Browse child nodes
  read                Read a node value
  write               Write a scalar node value
  monitor             Subscribe to data changes`)
}
