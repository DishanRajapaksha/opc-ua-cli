package uaclient

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/gopcua/opcua/ua"
)

func (s *Service) NamespaceArray(ctx context.Context) ([]string, error) {
	row, err := s.Read(ctx, "i=2255")
	if err != nil {
		return nil, err
	}
	values, ok := row.Value.([]string)
	if !ok {
		return nil, fmt.Errorf("%w: namespace array has unexpected type", ErrBadStatusCode)
	}
	return values, nil
}

func (s *Service) ResolveNodeID(ctx context.Context, raw string) (*ua.NodeID, string, error) {
	trimmed := strings.TrimSpace(raw)
	if !strings.HasPrefix(trimmed, "nsu=") {
		nodeID, err := ua.ParseNodeID(trimmed)
		if err != nil {
			return nil, "", fmt.Errorf("%w: invalid node id", ErrValidation)
		}
		return nodeID, trimmed, nil
	}
	parts := strings.SplitN(trimmed, ";", 2)
	if len(parts) != 2 {
		return nil, "", fmt.Errorf("%w: invalid nsu node id", ErrValidation)
	}
	nsValue := strings.TrimPrefix(parts[0], "nsu=")
	uri := nsValue
	if mapped, ok := s.cfg.Namespaces[nsValue]; ok {
		uri = mapped
	}
	array, err := s.NamespaceArray(ctx)
	if err != nil {
		return nil, "", err
	}
	index := -1
	for i, candidate := range array {
		if candidate == uri {
			index = i
			break
		}
	}
	if index < 0 {
		return nil, "", fmt.Errorf("%w: namespace URI %q not present on server", ErrNodeNotFound, uri)
	}
	resolved := "ns=" + strconv.Itoa(index) + ";" + parts[1]
	nodeID, err := ua.ParseNodeID(resolved)
	if err != nil {
		return nil, "", fmt.Errorf("%w: invalid resolved node id", ErrValidation)
	}
	return nodeID, resolved, nil
}
