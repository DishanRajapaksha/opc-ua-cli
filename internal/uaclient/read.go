package uaclient

import (
	"context"
	"fmt"
	"time"

	"github.com/DishanRajapaksha/opc-ua-cli/internal/domain"
	"github.com/gopcua/opcua/ua"
)

func (s *Service) Read(ctx context.Context, node string) (domain.ReadResult, error) {
	nodeID, err := ua.ParseNodeID(node)
	if err != nil {
		return domain.ReadResult{}, err
	}

	response, err := s.client.Read(ctx, &ua.ReadRequest{
		MaxAge:             2000,
		TimestampsToReturn: ua.TimestampsToReturnBoth,
		NodesToRead: []*ua.ReadValueID{
			{NodeID: nodeID, AttributeID: ua.AttributeIDValue},
		},
	})
	if err != nil {
		return domain.ReadResult{}, err
	}
	if len(response.Results) == 0 {
		return domain.ReadResult{}, fmt.Errorf("read returned no result")
	}

	result := response.Results[0]
	if result.Status != ua.StatusOK {
		return domain.ReadResult{}, fmt.Errorf("read failed: %s", result.Status)
	}

	value := interface{}(nil)
	if result.Value != nil {
		value = result.Value.Value()
	}

	return domain.ReadResult{
		NodeID:          node,
		Value:           value,
		Status:          fmt.Sprint(result.Status),
		SourceTimestamp: formatTimestamp(result.SourceTimestamp),
		ServerTimestamp: formatTimestamp(result.ServerTimestamp),
	}, nil
}

func formatTimestamp(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.UTC().Format(time.RFC3339Nano)
}
