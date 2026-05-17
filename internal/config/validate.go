package config

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"strings"
)

var allowedPolicies = map[string]bool{
	"none":           true,
	"basic128rsa15":  true,
	"basic256":       true,
	"basic256sha256": true,
	"aes128_sha256":  true,
	"aes256_sha256":  true,
}

var allowedModes = map[string]bool{
	"none":           true,
	"sign":           true,
	"signandencrypt": true,
}

func ValidateClientConfig(cfg ClientConfig) error {
	if strings.TrimSpace(cfg.Endpoint) == "" {
		return fmt.Errorf("%w: endpoint cannot be empty", ErrConfig)
	}
	if cfg.Timeout <= 0 {
		return fmt.Errorf("%w: timeout must be greater than zero", ErrConfig)
	}
	if !allowedPolicies[strings.ToLower(strings.TrimSpace(cfg.Policy))] {
		return fmt.Errorf("%w: unsupported security policy %q", ErrConfig, cfg.Policy)
	}
	if !allowedModes[strings.ToLower(strings.TrimSpace(cfg.Mode))] {
		return fmt.Errorf("%w: unsupported security mode %q", ErrConfig, cfg.Mode)
	}

	certDER := cfg.CertDER
	if len(certDER) == 0 && cfg.CertFile != "" {
		data, err := os.ReadFile(cfg.CertFile)
		if err != nil {
			return fmt.Errorf("%w: read cert file %q: %v", ErrConfig, cfg.CertFile, err)
		}
		if block, _ := pem.Decode(data); block != nil {
			data = block.Bytes
		}
		certDER = data
	}

	privateKey := cfg.PrivateKey
	if privateKey == nil && cfg.KeyFile != "" {
		data, err := os.ReadFile(cfg.KeyFile)
		if err != nil {
			return fmt.Errorf("%w: read key file %q: %v", ErrConfig, cfg.KeyFile, err)
		}
		privateKey, err = decodePrivateKey(string(data))
		if err != nil {
			return fmt.Errorf("%w: parse key file %q: %v", ErrConfig, cfg.KeyFile, err)
		}
	}

	if len(certDER) > 0 && privateKey != nil {
		cert, err := x509.ParseCertificate(certDER)
		if err != nil {
			return fmt.Errorf("%w: parse certificate: %v", ErrConfig, err)
		}
		certPub, ok := cert.PublicKey.(*rsa.PublicKey)
		if !ok {
			return fmt.Errorf("%w: certificate public key is not RSA", ErrConfig)
		}
		if certPub.N.Cmp(privateKey.N) != 0 || certPub.E != privateKey.E {
			return fmt.Errorf("%w: certificate and private key do not match", ErrConfig)
		}
	}
	return nil
}
