package uaclient

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/DishanRajapaksha/opc-ua-cli/internal/domain"
	"github.com/gopcua/opcua/ua"
)

func (s *Service) Write(ctx context.Context, node string, valueType string, rawValue string) (domain.WriteResult, error) {
	nodeID, resolvedNode, err := s.ResolveNodeID(ctx, node)
	if err != nil {
		return domain.WriteResult{}, err
	}

	parsed, err := parseScalar(valueType, rawValue)
	if err != nil {
		return domain.WriteResult{}, fmt.Errorf("%w: invalid value for type %q", ErrValidation, valueType)
	}

	variant, err := ua.NewVariant(parsed)
	if err != nil {
		return domain.WriteResult{}, fmt.Errorf("%w: invalid OPC UA value", ErrValidation)
	}

	response, err := s.client.Write(ctx, &ua.WriteRequest{
		NodesToWrite: []*ua.WriteValue{
			{
				NodeID:      nodeID,
				AttributeID: ua.AttributeIDValue,
				Value: &ua.DataValue{
					EncodingMask: ua.DataValueValue,
					Value:        variant,
				},
			},
		},
	})
	if err != nil {
		return domain.WriteResult{}, fmt.Errorf("%w: write request failed", ErrConnection)
	}
	if len(response.Results) == 0 {
		return domain.WriteResult{}, fmt.Errorf("%w: write returned no result", ErrWriteRejected)
	}
	if response.Results[0] != ua.StatusOK {
		if response.Results[0] == ua.StatusBadNodeIDUnknown {
			return domain.WriteResult{}, fmt.Errorf("%w: %s", ErrNodeNotFound, resolvedNode)
		}
		return domain.WriteResult{}, fmt.Errorf("%w: %s", ErrWriteRejected, response.Results[0])
	}

	return domain.WriteResult{NodeID: resolvedNode, Status: fmt.Sprint(response.Results[0])}, nil
}

func parseScalar(valueType string, raw string) (interface{}, error) {
	switch strings.ToLower(strings.TrimSpace(valueType)) {
	case "", "string":
		return raw, nil
	case "bool", "boolean":
		return strconv.ParseBool(raw)
	case "int8", "sbyte":
		value, err := strconv.ParseInt(raw, 10, 8)
		return int8(value), err
	case "int16":
		value, err := strconv.ParseInt(raw, 10, 16)
		return int16(value), err
	case "int", "int32":
		value, err := strconv.ParseInt(raw, 10, 32)
		return int32(value), err
	case "int64":
		return strconv.ParseInt(raw, 10, 64)
	case "uint8", "byte":
		value, err := strconv.ParseUint(raw, 10, 8)
		return byte(value), err
	case "uint16":
		value, err := strconv.ParseUint(raw, 10, 16)
		return uint16(value), err
	case "uint", "uint32":
		value, err := strconv.ParseUint(raw, 10, 32)
		return uint32(value), err
	case "uint64":
		return strconv.ParseUint(raw, 10, 64)
	case "float", "float32":
		value, err := strconv.ParseFloat(raw, 32)
		return float32(value), err
	case "double", "float64":
		return strconv.ParseFloat(raw, 64)
	default:
		return nil, fmt.Errorf("unsupported scalar type %q", valueType)
	}
}
