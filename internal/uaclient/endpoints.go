package uaclient

import (
	"context"
	"fmt"
	"strings"

	"github.com/DishanRajapaksha/opc-ua-cli/internal/domain"
	"github.com/gopcua/opcua"
)

func ListEndpoints(ctx context.Context, endpoint string) ([]domain.Endpoint, error) {
	items, err := opcua.GetEndpoints(ctx, endpoint)
	if err != nil {
		return nil, err
	}

	rows := make([]domain.Endpoint, 0, len(items))
	for _, ep := range items {
		authModes := make([]string, 0, len(ep.UserIdentityTokens))
		for _, policy := range ep.UserIdentityTokens {
			authModes = append(authModes, fmt.Sprint(policy.TokenType))
		}

		rows = append(rows, domain.Endpoint{
			EndpointURL:    ep.EndpointURL,
			SecurityPolicy: shortSecurityPolicy(ep.SecurityPolicyURI),
			SecurityMode:   fmt.Sprint(ep.SecurityMode),
			UserTokens:     authModes,
		})
	}

	return rows, nil
}

func shortSecurityPolicy(uri string) string {
	parts := strings.Split(uri, "#")
	return parts[len(parts)-1]
}
