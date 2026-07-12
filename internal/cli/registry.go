package cli

import "github.com/DishanRajapaksha/industrial-cli-kit/command"

var cliRegistry = command.Registry{
	Binary: "opc-ua-cli",
	GlobalFlags: []command.Flag{
		{Name: "config", TakesValue: true, Summary: "YAML config file"},
		{Name: "profile", TakesValue: true, Summary: "config profile name"},
		{Name: "endpoint", TakesValue: true, Summary: "OPC UA endpoint URL"},
		{Name: "policy", TakesValue: true, Summary: "security policy"},
		{Name: "mode", TakesValue: true, Summary: "security mode"},
		{Name: "username", TakesValue: true, Summary: "username authentication"},
		{Name: "password", TakesValue: true, Summary: "password authentication"},
		{Name: "cert", TakesValue: true, Summary: "client certificate file"},
		{Name: "key", TakesValue: true, Summary: "client private key file"},
		{Name: "timeout", TakesValue: true, Summary: "request timeout"},
		{Name: "format", TakesValue: true, Summary: "output format"},
		{Name: "verbose", Summary: "print connection decisions"},
		{Name: "debug", Summary: "enable client diagnostics"},
	},
	Commands: []command.Command{
		{Name: "endpoints", Summary: "List server endpoints and security options"},
		{Name: "status", Summary: "Read server status"},
		{Name: "namespaces", Summary: "List namespace indexes and URIs"},
		{Name: "browse", Summary: "Browse child nodes", Flags: registryFlags("node", "depth")},
		{Name: "tui", Summary: "Browse nodes interactively", Flags: registryFlags("node", "interval")},
		{Name: "attributes", Summary: "Inspect node metadata attributes", Flags: registryFlags("node")},
		{Name: "read", Summary: "Read node values", Flags: registryFlags("node", "nodes")},
		{Name: "write", Summary: "Write node values", Flags: registryFlags("node", "type", "value", "item", "dry-run", "yes")},
		{Name: "monitor", Summary: "Subscribe to data changes", Flags: registryFlags("node", "interval", "duration")},
		{Name: "watch", Summary: "Poll node values", Flags: registryFlags("node", "interval", "duration")},
		{Name: "alarms", Summary: "Subscribe to alarms and events", Flags: registryFlags("node", "interval", "duration", "min-severity")},
		{Name: "test-connection", Summary: "Run connection diagnostics"},
		{Name: "validate-config", Summary: "Validate local config"},
		{Name: "init-config", Summary: "Write a starter YAML config", Flags: registryFlags("output", "force")},
		{Name: "completions", Summary: "Generate shell completion scripts"},
		{Name: "help", Summary: "Print help"},
		{Name: "version", Summary: "Print version information"},
	},
}

func registryFlags(names ...string) []command.Flag {
	flags := make([]command.Flag, 0, len(names))
	for _, name := range names {
		takesValue := name != "force" && name != "dry-run" && name != "yes"
		flags = append(flags, command.Flag{Name: name, TakesValue: takesValue})
	}
	return flags
}
