package cli

import (
	"flag"
	"time"

	"github.com/DishanRajapaksha/opc-ua-cli/internal/config"
)

type commonOptions struct {
	client config.ClientConfig
	format string
}

func addCommonFlags(fs *flag.FlagSet, opts *commonOptions) {
	cfg := config.DefaultClientConfig()
	opts.client = cfg

	fs.StringVar(&opts.client.Endpoint, "endpoint", cfg.Endpoint, "OPC UA endpoint URL")
	fs.StringVar(&opts.client.Policy, "policy", cfg.Policy, "security policy")
	fs.StringVar(&opts.client.Mode, "mode", cfg.Mode, "security mode")
	fs.StringVar(&opts.client.Username, "username", cfg.Username, "username")
	fs.StringVar(&opts.client.Password, "password", cfg.Password, "password")
	fs.StringVar(&opts.client.CertFile, "cert", cfg.CertFile, "client certificate file")
	fs.StringVar(&opts.client.KeyFile, "key", cfg.KeyFile, "client private key file")
	fs.DurationVar(&opts.client.Timeout, "timeout", cfg.Timeout, "request timeout")
	fs.StringVar(&opts.format, "format", "table", "output format: table, text, json")
}

func operationContext(timeout time.Duration) (contextDone <-chan struct{}, cancel func()) {
	timer := time.NewTimer(timeout)
	closed := make(chan struct{})
	go func() {
		<-timer.C
		close(closed)
	}()
	return closed, func() {
		if timer.Stop() {
			close(closed)
		}
	}
}
