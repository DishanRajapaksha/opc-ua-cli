package uaclient

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/DishanRajapaksha/opc-ua-cli/internal/config"
	"github.com/gopcua/opcua"
	opcdebug "github.com/gopcua/opcua/debug"
	"github.com/gopcua/opcua/ua"
)

// Service owns the OPC UA session lifecycle and exposes application-level operations.
type Service struct {
	cfg    config.ClientConfig
	client *opcua.Client
}

func NewService(cfg config.ClientConfig) *Service {
	return &Service{cfg: cfg}
}

func (s *Service) Connect(ctx context.Context) error {
	if s.cfg.Debug {
		opcdebug.Enable = true
		opcdebug.Logger = log.New(os.Stderr, "opcua-debug: ", 0)
	}
	s.verbosef("connect start endpoint=%s timeout=%s policy=%s mode=%s auth=%s cert=%t key=%t", s.cfg.Endpoint, s.cfg.Timeout, s.cfg.Policy, s.cfg.Mode, authMode(s.cfg), s.hasCert(), s.hasKey())

	authType := ua.UserTokenTypeAnonymous
	auth := opcua.AuthAnonymous()
	if s.cfg.Username != "" {
		authType = ua.UserTokenTypeUserName
		auth = opcua.AuthUsername(s.cfg.Username, s.cfg.Password)
	}

	opts := []opcua.Option{
		opcua.ApplicationName("opc-ua-cli"),
		opcua.ApplicationURI("urn:github.com:DishanRajapaksha:opc-ua-cli"),
		opcua.AutoReconnect(true),
		opcua.RequestTimeout(s.cfg.Timeout),
		opcua.SecurityPolicy(s.cfg.Policy),
		opcua.SecurityModeString(s.cfg.Mode),
		auth,
	}

	if len(s.cfg.CertDER) > 0 {
		opts = append(opts, opcua.Certificate(s.cfg.CertDER))
	} else if s.cfg.CertFile != "" {
		opts = append(opts, opcua.CertificateFile(s.cfg.CertFile))
	}
	if s.cfg.PrivateKey != nil {
		opts = append(opts, opcua.PrivateKey(s.cfg.PrivateKey))
	} else if s.cfg.KeyFile != "" {
		opts = append(opts, opcua.PrivateKeyFile(s.cfg.KeyFile))
	}

	endpoint := s.cfg.Endpoint
	if s.usesEndpointSelection() {
		s.verbosef("endpoint selection enabled")
		endpoints, err := opcua.GetEndpoints(ctx, s.cfg.Endpoint)
		if err != nil {
			return fmt.Errorf("%w: cannot fetch server endpoints", ErrConnection)
		}

		selected, err := opcua.SelectEndpoint(endpoints, s.cfg.Policy, ua.MessageSecurityModeFromString(s.cfg.Mode))
		if err != nil {
			return fmt.Errorf("%w: no matching endpoint for requested security settings", ErrAuthSecurity)
		}

		endpoint = selected.EndpointURL
		s.verbosef("selected endpoint=%s", endpoint)
		opts = append(opts, opcua.SecurityFromEndpoint(selected, authType))
	}

	client, err := opcua.NewClient(endpoint, opts...)
	if err != nil {
		return fmt.Errorf("%w: invalid client configuration", ErrConnection)
	}
	if err := client.Connect(ctx); err != nil {
		if s.cfg.UsesSecurity() {
			return fmt.Errorf("%w: failed to establish secure session", ErrAuthSecurity)
		}
		return fmt.Errorf("%w: failed to connect to endpoint", ErrConnection)
	}
	s.verbosef("session established")

	s.client = client
	return nil
}

func (s *Service) verbosef(format string, args ...interface{}) {
	if !s.cfg.Verbose {
		return
	}
	fmt.Fprintf(os.Stderr, "verbose: "+format+"\n", args...)
}

func authMode(cfg config.ClientConfig) string {
	if cfg.Username != "" {
		return "username"
	}
	return "anonymous"
}

func (s *Service) hasCert() bool {
	return len(s.cfg.CertDER) > 0 || s.cfg.CertFile != ""
}

func (s *Service) hasKey() bool {
	return s.cfg.PrivateKey != nil || s.cfg.KeyFile != ""
}

func (s *Service) Close(ctx context.Context) {
	if s.client != nil {
		s.client.Close(ctx)
	}
}

func (s *Service) opcClient() *opcua.Client {
	return s.client
}

func (s *Service) usesEndpointSelection() bool {
	return s.cfg.Username != "" ||
		!strings.EqualFold(s.cfg.Policy, "None") ||
		!strings.EqualFold(s.cfg.Mode, "None") ||
		s.cfg.CertFile != "" ||
		s.cfg.KeyFile != "" ||
		len(s.cfg.CertDER) > 0 ||
		s.cfg.PrivateKey != nil
}
