package cli

import (
	"flag"

	"github.com/DishanRajapaksha/opc-ua-cli/internal/config"
)

type commonOptions struct {
	client     config.ClientConfig
	format     string
	configPath string
	profile    string
	verbose    bool
	debug      bool
}

func addCommonFlags(fs *flag.FlagSet, opts *commonOptions, defaultFormat string, formatHelp string) {
	cfg := config.DefaultClientConfig()
	opts.client = cfg
	opts.configPath = config.DefaultConfigPath

	fs.StringVar(&opts.configPath, "config", config.DefaultConfigPath, "YAML config file")
	fs.StringVar(&opts.profile, "profile", "", "config profile name")
	fs.StringVar(&opts.client.Endpoint, "endpoint", cfg.Endpoint, "OPC UA endpoint URL")
	fs.StringVar(&opts.client.Policy, "policy", cfg.Policy, "security policy")
	fs.StringVar(&opts.client.Mode, "mode", cfg.Mode, "security mode")
	fs.StringVar(&opts.client.Username, "username", cfg.Username, "username")
	fs.StringVar(&opts.client.Password, "password", cfg.Password, "password")
	fs.StringVar(&opts.client.CertFile, "cert", cfg.CertFile, "client certificate file")
	fs.StringVar(&opts.client.KeyFile, "key", cfg.KeyFile, "client private key file")
	fs.DurationVar(&opts.client.Timeout, "timeout", cfg.Timeout, "request timeout")
	fs.StringVar(&opts.format, "format", defaultFormat, formatHelp)
	fs.BoolVar(&opts.verbose, "verbose", false, "print high-level connection decisions")
	fs.BoolVar(&opts.debug, "debug", false, "enable lower-level OPC UA client debug logging")
}

func (opts *commonOptions) applyConfig(fs *flag.FlagSet) error {
	fileCfg, err := config.LoadClientConfigForProfile(opts.configPath, opts.profile)
	if err != nil {
		return err
	}

	cliCfg := opts.client
	opts.client = fileCfg
	visited := visitedFlags(fs)

	if visited["endpoint"] {
		opts.client.Endpoint = cliCfg.Endpoint
	}
	if visited["policy"] {
		opts.client.Policy = cliCfg.Policy
	}
	if visited["mode"] {
		opts.client.Mode = cliCfg.Mode
	}
	if visited["username"] {
		opts.client.Username = cliCfg.Username
	}
	if visited["password"] {
		opts.client.Password = cliCfg.Password
	}
	if visited["cert"] {
		opts.client.CertFile = cliCfg.CertFile
	}
	if visited["key"] {
		opts.client.KeyFile = cliCfg.KeyFile
	}
	if visited["timeout"] {
		opts.client.Timeout = cliCfg.Timeout
	}
	opts.client.Verbose = opts.verbose
	opts.client.Debug = opts.debug

	return nil
}

func visitedFlags(fs *flag.FlagSet) map[string]bool {
	visited := map[string]bool{}
	fs.Visit(func(f *flag.Flag) {
		visited[f.Name] = true
	})
	return visited
}
