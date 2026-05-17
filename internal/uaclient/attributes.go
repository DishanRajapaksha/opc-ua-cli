package uaclient

import (
	"context"
	"fmt"

	"github.com/DishanRajapaksha/opc-ua-cli/internal/domain"
	"github.com/gopcua/opcua/ua"
)

func (s *Service) Attributes(ctx context.Context, node string) (domain.NodeAttributesResult, error) {
	nodeID, resolvedNode, err := s.ResolveNodeID(ctx, node)
	if err != nil {
		return domain.NodeAttributesResult{}, err
	}

	opcNode := s.client.Node(nodeID)
	attrIDs := []ua.AttributeID{
		ua.AttributeIDNodeID,
		ua.AttributeIDNodeClass,
		ua.AttributeIDBrowseName,
		ua.AttributeIDDisplayName,
		ua.AttributeIDDescription,
		ua.AttributeIDDataType,
		ua.AttributeIDValueRank,
		ua.AttributeIDAccessLevel,
		ua.AttributeIDUserAccessLevel,
	}
	raw, err := opcNode.Attributes(ctx, attrIDs...)
	if err != nil {
		return domain.NodeAttributesResult{}, fmt.Errorf("%w: failed to read attributes", ErrConnection)
	}

	names := []string{"NodeId", "NodeClass", "BrowseName", "DisplayName", "Description", "DataType", "ValueRank", "AccessLevel", "UserAccessLevel"}
	rows := make([]domain.NodeAttribute, 0, len(attrIDs))

	for i, name := range names {
		if i >= len(raw) || raw[i] == nil {
			rows = append(rows, domain.NodeAttribute{Name: name, Status: "StatusUnknown"})
			continue
		}
		entry := domain.NodeAttribute{Name: name, Status: fmt.Sprint(raw[i].Status)}
		if raw[i].Value != nil {
			entry.Value = attributeValue(name, raw[i].Value)
		}
		rows = append(rows, entry)
	}

	return domain.NodeAttributesResult{NodeID: resolvedNode, Attributes: rows}, nil
}

func attributeValue(name string, v *ua.Variant) interface{} {
	switch name {
	case "NodeClass":
		return fmt.Sprint(ua.NodeClass(v.Int()))
	case "DataType":
		if id := v.NodeID(); id != nil {
			return dataTypeName(id)
		}
	case "AccessLevel", "UserAccessLevel":
		return fmt.Sprintf("0x%X", uint8(v.Int()))
	}
	return v.Value()
}
