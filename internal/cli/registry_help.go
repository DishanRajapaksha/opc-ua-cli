package cli

import sharedhelp "github.com/DishanRajapaksha/industrial-cli-kit/help"

func (a *App) writeRegistryUsage() {
	_ = sharedhelp.Write(a.out, cliRegistry, sharedhelp.Options{
		Description: "opc-ua-cli is a small OPC UA command-line client.",
		Usage:       []string{"opc-ua-cli [global flags] <command> [flags]"},
		Examples: []string{
			"opc-ua-cli endpoints --profile local",
			"opc-ua-cli namespaces --profile local",
			"opc-ua-cli browse --profile local --node i=84 --depth 1",
			"opc-ua-cli tui --profile local --node i=84 --interval 1s",
			"opc-ua-cli attributes --profile local --node 'ns=2;s=Demo.Value'",
			"opc-ua-cli read --profile site-a --node 'ns=2;s=Demo.Static.Scalar.Int32'",
			"opc-ua-cli write --profile site-a --node 'ns=2;s=Demo.Static.Scalar.Int32' --type int32 --value 42 --yes",
			"opc-ua-cli monitor --profile site-a --node 'ns=2;s=Demo.Static.Scalar.Int32' --interval 1s",
			"opc-ua-cli alarms --profile site-a --node i=2253 --min-severity 500 --interval 1s",
			"opc-ua-cli test-connection --profile site-a",
			"opc-ua-cli completions zsh",
		},
	})
}
