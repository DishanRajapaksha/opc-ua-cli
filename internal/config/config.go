package config

import "time"

const DefaultEndpoint = "opc.tcp://localhost:4840"

// ClientConfig contains connection and security settings for an OPC UA client session.
type ClientConfig struct {
	Endpoint string
	Policy   string
	Mode     string
	Username string
	Password string
	CertFile string
	KeyFile  string
	Timeout  time.Duration
}

func DefaultClientConfig() ClientConfig {
	return ClientConfig{
		Endpoint: DefaultEndpoint,
		Policy:   "None",
		Mode:     "None",
		Timeout:  10 * time.Second,
	}
}

func (c ClientConfig) UsesSecurity() bool {
	return c.Username != "" || c.Policy != "None" || c.Mode != "None" || c.CertFile != "" || c.KeyFile != ""
}
