package uaclient

import (
	"context"
	"fmt"

	"github.com/DishanRajapaksha/opc-ua-cli/internal/domain"
	"github.com/gopcua/opcua"
	"github.com/gopcua/opcua/id"
	"github.com/gopcua/opcua/ua"
)

func (s *Service) Browse(ctx context.Context, root string, maxDepth int) ([]domain.Node, error) {
	nodeID, err := ua.ParseNodeID(root)
	if err != nil {
		return nil, err
	}

	rows := make([]domain.Node, 0)
	seen := map[string]bool{}
	if err := s.browse(ctx, s.client.Node(nodeID), "", 0, maxDepth, seen, &rows); err != nil {
		return nil, err
	}

	return rows, nil
}

func (s *Service) browse(ctx context.Context, node *opcua.Node, parentPath string, level int, maxDepth int, seen map[string]bool, rows *[]domain.Node) error {
	if level > maxDepth {
		return nil
	}

	for _, refType := range []uint32{id.HasComponent, id.Organizes, id.HasProperty} {
		children, err := node.ReferencedNodes(ctx, refType, ua.BrowseDirectionForward, ua.NodeClassAll, true)
		if err != nil {
			return err
		}

		for _, child := range children {
			if child == nil || child.ID == nil {
				continue
			}

			key := child.ID.String()
			if seen[key] {
				continue
			}
			seen[key] = true

			record, err := inspectNode(ctx, child, parentPath)
			if err != nil {
				return err
			}
			*rows = append(*rows, record)

			if level < maxDepth {
				if err := s.browse(ctx, child, record.Path, level+1, maxDepth, seen, rows); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func inspectNode(ctx context.Context, node *opcua.Node, parentPath string) (domain.Node, error) {
	attrs, err := node.Attributes(ctx, ua.AttributeIDNodeClass, ua.AttributeIDBrowseName, ua.AttributeIDDisplayName, ua.AttributeIDDataType, ua.AttributeIDAccessLevel)
	if err != nil {
		return domain.Node{}, err
	}

	record := domain.Node{NodeID: node.ID.String()}
	if attributeOK(attrs, 0) {
		record.NodeClass = fmt.Sprint(ua.NodeClass(attrs[0].Value.Int()))
	}
	if attributeOK(attrs, 1) {
		record.BrowseName = attrs[1].Value.String()
	}
	if attributeOK(attrs, 2) {
		record.DisplayName = attrs[2].Value.String()
	}
	if attributeOK(attrs, 3) {
		if dataType := attrs[3].Value.NodeID(); dataType != nil {
			record.DataType = dataTypeName(dataType)
		}
	}
	if attributeOK(attrs, 4) {
		access := ua.AccessLevelType(attrs[4].Value.Int())
		record.Writable = access&ua.AccessLevelTypeCurrentWrite == ua.AccessLevelTypeCurrentWrite
	}

	name := record.BrowseName
	if name == "" {
		name = record.DisplayName
	}
	record.Path = joinPath(parentPath, name)

	return record, nil
}

func attributeOK(attrs []*ua.DataValue, index int) bool {
	return index < len(attrs) && attrs[index] != nil && attrs[index].Status == ua.StatusOK && attrs[index].Value != nil
}

func dataTypeName(nodeID *ua.NodeID) string {
	switch nodeID.IntID() {
	case id.Boolean:
		return "Boolean"
	case id.Int16:
		return "Int16"
	case id.Int32:
		return "Int32"
	case id.Int64:
		return "Int64"
	case id.UInt16:
		return "UInt16"
	case id.UInt32:
		return "UInt32"
	case id.UInt64:
		return "UInt64"
	case id.Float:
		return "Float"
	case id.Double:
		return "Double"
	case id.String:
		return "String"
	case id.DateTime, id.UtcTime:
		return "DateTime"
	default:
		return nodeID.String()
	}
}

func joinPath(parent string, child string) string {
	if parent == "" {
		return child
	}
	if child == "" {
		return parent
	}
	return parent + "." + child
}
