package uaclient

import (
	"context"
	"strings"

	"github.com/DishanRajapaksha/opc-ua-cli/internal/config"
	"github.com/gopcua/opcua"
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

	if s.cfg.CertFile != "" {
		opts = append(opts, opcua.CertificateFile(s.cfg.CertFile))
	}
	if s.cfg.KeyFile != "" {
		opts = append(opts, opcua.PrivateKeyFile(s.cfg.KeyFile))
	}

	endpoint := s.cfg.Endpoint
	if s.usesEndpointSelection() {
		endpoints, err := opcua.GetEndpoints(ctx, s.cfg.Endpoint)
		if err != nil {
			return err
		}

		selected, err := opcua.SelectEndpoint(endpoints, s.cfg.Policy, ua.MessageSecurityModeFromString(s.cfg.Mode))
		if err != nil {
			return err
		}

		endpoint = selected.EndpointURL
		opts = append(opts, opcua.SecurityFromEndpoint(selected, authType))
	}

	client, err := opcua.NewClient(endpoint, opts...)
	if err != nil {
		return err
	}
	if err := client.Connect(ctx); err != nil {
		return err
	}

	s.client = client
	return nil
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
		s.cfg.KeyFile != ""
}
