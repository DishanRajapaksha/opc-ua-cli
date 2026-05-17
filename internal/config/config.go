package config

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	DefaultEndpoint   = "opc.tcp://localhost:4840"
	DefaultConfigPath = "config.yaml"
)

// ClientConfig contains connection and security settings for an OPC UA client session.
type ClientConfig struct {
	Endpoint   string
	Policy     string
	Mode       string
	Username   string
	Password   string
	CertFile   string
	KeyFile    string
	CertDER    []byte
	PrivateKey *rsa.PrivateKey
	Timeout    time.Duration
	Namespaces map[string]string
	Verbose    bool
	Debug      bool
}

func DefaultClientConfig() ClientConfig {
	return ClientConfig{
		Endpoint: DefaultEndpoint,
		Policy:   "None",
		Mode:     "None",
		Timeout:  10 * time.Second,
	}
}

func LoadClientConfig(path string) (ClientConfig, error) {
	return LoadClientConfigForProfile(path, "")
}

func LoadClientConfigForProfile(path string, profile string) (ClientConfig, error) {
	cfg := DefaultClientConfig()
	if path == "" {
		return cfg, nil
	}

	contents, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) && path == DefaultConfigPath {
			return cfg, nil
		}
		return cfg, fmt.Errorf("%w: read config %q: %v", ErrConfig, path, err)
	}

	var file configFile
	if err := yaml.Unmarshal(contents, &file); err != nil {
		return cfg, fmt.Errorf("%w: parse config %q: %v", ErrConfig, path, err)
	}

	if err := applySettings(&cfg, file.settings); err != nil {
		return cfg, err
	}

	if len(file.Profiles) > 0 {
		selectedProfile := profile
		if selectedProfile == "" {
			selectedProfile = file.DefaultProfile
		}
		if selectedProfile == "" {
			return cfg, fmt.Errorf("%w: config has profiles but no profile was selected and default_profile is empty", ErrConfig)
		}

		settings, ok := file.Profiles[selectedProfile]
		if !ok {
			return cfg, fmt.Errorf("%w: profile %q not found", ErrConfig, selectedProfile)
		}
		if err := applySettings(&cfg, settings); err != nil {
			return cfg, fmt.Errorf("%w: profile %q: %v", ErrConfig, selectedProfile, err)
		}
	}

	if cfg.Endpoint == "" {
		return cfg, fmt.Errorf("%w: endpoint cannot be empty", ErrConfig)
	}
	if err := ValidateClientConfig(cfg); err != nil {
		return cfg, err
	}

	return cfg, nil
}

func StarterConfigYAML() ([]byte, error) {
	defaults := DefaultClientConfig()
	root := &yaml.Node{Kind: yaml.MappingNode}
	root.Content = append(root.Content,
		scalar("endpoint"), commentedValue(defaults.Endpoint, "OPC UA endpoint URL"),
		scalar("policy"), commentedValue(defaults.Policy, "Security policy (None, Basic256Sha256, ...)"),
		scalar("mode"), commentedValue(defaults.Mode, "Security mode (None, Sign, SignAndEncrypt)"),
		scalar("timeout"), commentedValue(defaults.Timeout.String(), "Request timeout (for example 10s, 30s)"),
		scalar("username"), commentedValue("", "Optional username for user/password auth"),
		scalar("password"), commentedValue("", "Optional password for user/password auth"),
		scalar("cert_base64"), commentedValue("", "Optional base64 client certificate (DER or PEM bytes)"),
		scalar("key_base64"), commentedValue("", "Optional base64 RSA private key (PKCS#1 or PKCS#8 PEM/DER)"),
	)

	doc := &yaml.Node{Kind: yaml.DocumentNode, Content: []*yaml.Node{root}}
	out, err := yaml.Marshal(doc)
	if err != nil {
		return nil, fmt.Errorf("marshal starter config: %w", err)
	}
	return out, nil
}

func (c ClientConfig) UsesSecurity() bool {
	return c.Username != "" || c.Policy != "None" || c.Mode != "None" || c.CertFile != "" || c.KeyFile != "" || len(c.CertDER) > 0 || c.PrivateKey != nil
}

type configFile struct {
	settings       `yaml:",inline"`
	DefaultProfile string              `yaml:"default_profile"`
	Profiles       map[string]settings `yaml:"profiles"`
}

type settings struct {
	Endpoint   string            `yaml:"endpoint"`
	Policy     string            `yaml:"policy"`
	Mode       string            `yaml:"mode"`
	Username   string            `yaml:"username"`
	Password   string            `yaml:"password"`
	CertFile   string            `yaml:"cert"`
	KeyFile    string            `yaml:"key"`
	CertBase64 string            `yaml:"cert_base64"`
	KeyBase64  string            `yaml:"key_base64"`
	Timeout    string            `yaml:"timeout"`
	Namespaces map[string]string `yaml:"namespaces"`
}

func applySettings(cfg *ClientConfig, file settings) error {
	if file.Endpoint != "" {
		cfg.Endpoint = file.Endpoint
	}
	if file.Policy != "" {
		cfg.Policy = file.Policy
	}
	if file.Mode != "" {
		cfg.Mode = file.Mode
	}
	if file.Username != "" {
		cfg.Username = file.Username
	}
	if file.Password != "" {
		cfg.Password = file.Password
	}
	if file.CertFile != "" {
		cfg.CertFile = file.CertFile
	}
	if file.KeyFile != "" {
		cfg.KeyFile = file.KeyFile
	}
	if file.CertBase64 != "" {
		certDER, err := decodeCertificate(file.CertBase64)
		if err != nil {
			return fmt.Errorf("%w: parse cert_base64: %v", ErrConfig, err)
		}
		cfg.CertDER = certDER
	}
	if file.KeyBase64 != "" {
		privateKey, err := decodePrivateKey(file.KeyBase64)
		if err != nil {
			return fmt.Errorf("%w: parse key_base64: %v", ErrConfig, err)
		}
		cfg.PrivateKey = privateKey
	}
	if file.Timeout != "" {
		timeout, err := time.ParseDuration(file.Timeout)
		if err != nil {
			return fmt.Errorf("%w: parse timeout %q: %v", ErrConfig, file.Timeout, err)
		}
		cfg.Timeout = timeout
	}
	if len(file.Namespaces) > 0 {
		if cfg.Namespaces == nil {
			cfg.Namespaces = map[string]string{}
		}
		for k, v := range file.Namespaces {
			cfg.Namespaces[k] = v
		}
	}

	return nil
}

func decodeCertificate(value string) ([]byte, error) {
	decoded, err := decodeBase64(value)
	if err != nil {
		return nil, err
	}
	if block, _ := pem.Decode(decoded); block != nil {
		if block.Type != "CERTIFICATE" {
			return nil, fmt.Errorf("expected CERTIFICATE PEM block, got %s", block.Type)
		}
		decoded = block.Bytes
	}
	if _, err := x509.ParseCertificate(decoded); err != nil {
		return nil, err
	}
	return decoded, nil
}

func decodePrivateKey(value string) (*rsa.PrivateKey, error) {
	decoded, err := decodeBase64(value)
	if err != nil {
		return nil, err
	}
	if block, _ := pem.Decode(decoded); block != nil {
		decoded = block.Bytes
	}
	if key, err := x509.ParsePKCS1PrivateKey(decoded); err == nil {
		return key, nil
	}
	parsed, err := x509.ParsePKCS8PrivateKey(decoded)
	if err != nil {
		return nil, err
	}
	key, ok := parsed.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("private key is not RSA")
	}
	return key, nil
}

func decodeBase64(value string) ([]byte, error) {
	compact := strings.Join(strings.Fields(value), "")
	decoded, err := base64.StdEncoding.DecodeString(compact)
	if err == nil {
		return decoded, nil
	}
	return base64.RawStdEncoding.DecodeString(compact)
}

func scalar(value string) *yaml.Node {
	return &yaml.Node{Kind: yaml.ScalarNode, Value: value}
}

func commentedValue(value string, comment string) *yaml.Node {
	return &yaml.Node{Kind: yaml.ScalarNode, Value: value, HeadComment: comment}
}
