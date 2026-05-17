package uaclient

import (
	"context"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/DishanRajapaksha/opc-ua-cli/internal/domain"
	"github.com/gopcua/opcua"
	"github.com/gopcua/opcua/id"
	"github.com/gopcua/opcua/ua"
)

// AlarmSubscription wraps an OPC UA event subscription and exposes alarm/event notifications.
type AlarmSubscription struct {
	Events <-chan domain.AlarmEvent
	Errors <-chan error
	close  func()
}

func (s AlarmSubscription) Close() {
	if s.close != nil {
		s.close()
	}
}

func (s *Service) SubscribeAlarms(ctx context.Context, node string, interval time.Duration, minSeverity uint16) (AlarmSubscription, error) {
	nodeID, err := ua.ParseNodeID(node)
	if err != nil {
		return AlarmSubscription{}, err
	}

	notifications := make(chan *opcua.PublishNotificationData, 32)
	subscription, err := s.client.Subscribe(ctx, &opcua.SubscriptionParameters{Interval: interval}, notifications)
	if err != nil {
		return AlarmSubscription{}, err
	}

	request, fieldNames := alarmEventRequest(nodeID, minSeverity)
	response, err := subscription.Monitor(ctx, ua.TimestampsToReturnBoth, request)
	if err != nil {
		subscription.Cancel(context.Background())
		return AlarmSubscription{}, err
	}
	if len(response.Results) == 0 || response.Results[0].StatusCode != ua.StatusOK {
		subscription.Cancel(context.Background())
		if len(response.Results) == 0 {
			return AlarmSubscription{}, fmt.Errorf("alarm subscription returned no result")
		}
		return AlarmSubscription{}, fmt.Errorf("alarm subscription failed: %s", response.Results[0].StatusCode)
	}

	events := make(chan domain.AlarmEvent, 32)
	errors := make(chan error, 8)

	go func() {
		defer close(events)
		defer close(errors)
		defer subscription.Cancel(context.Background())

		for {
			select {
			case <-ctx.Done():
				return
			case notification, ok := <-notifications:
				if !ok {
					return
				}
				if notification == nil {
					continue
				}
				if notification.Error != nil {
					select {
					case errors <- notification.Error:
					case <-ctx.Done():
					}
					continue
				}

				list, ok := notification.Value.(*ua.EventNotificationList)
				if !ok {
					continue
				}

				for _, item := range list.Events {
					event := alarmEventFromFields(node, fieldNames, item.EventFields)
					select {
					case events <- event:
					case <-ctx.Done():
						return
					}
				}
			}
		}
	}()

	return AlarmSubscription{
		Events: events,
		Errors: errors,
		close: func() {
			subscription.Cancel(context.Background())
		},
	}, nil
}

func alarmEventRequest(nodeID *ua.NodeID, minSeverity uint16) (*ua.MonitoredItemCreateRequest, []string) {
	fieldNames := []string{
		"EventId",
		"EventType",
		"SourceNode",
		"SourceName",
		"Time",
		"ReceiveTime",
		"Message",
		"Severity",
		"ConditionName",
		"ActiveState",
		"AckedState",
		"Retain",
	}

	selects := make([]*ua.SimpleAttributeOperand, 0, len(fieldNames))
	for _, name := range fieldNames {
		selects = append(selects, &ua.SimpleAttributeOperand{
			TypeDefinitionID: ua.NewNumericNodeID(0, id.BaseEventType),
			BrowsePath:       []*ua.QualifiedName{{NamespaceIndex: 0, Name: name}},
			AttributeID:      ua.AttributeIDValue,
		})
	}

	where := &ua.ContentFilter{
		Elements: []*ua.ContentFilterElement{
			{
				FilterOperator: ua.FilterOperatorGreaterThanOrEqual,
				FilterOperands: []*ua.ExtensionObject{
					{
						EncodingMask: ua.ExtensionObjectBinary,
						TypeID: &ua.ExpandedNodeID{
							NodeID: ua.NewNumericNodeID(0, id.SimpleAttributeOperand_Encoding_DefaultBinary),
						},
						Value: ua.SimpleAttributeOperand{
							TypeDefinitionID: ua.NewNumericNodeID(0, id.BaseEventType),
							BrowsePath:       []*ua.QualifiedName{{NamespaceIndex: 0, Name: "Severity"}},
							AttributeID:      ua.AttributeIDValue,
						},
					},
					{
						EncodingMask: ua.ExtensionObjectBinary,
						TypeID: &ua.ExpandedNodeID{
							NodeID: ua.NewNumericNodeID(0, id.LiteralOperand_Encoding_DefaultBinary),
						},
						Value: ua.LiteralOperand{Value: ua.MustVariant(minSeverity)},
					},
				},
			},
		},
	}

	filter := ua.EventFilter{
		SelectClauses: selects,
		WhereClause:   where,
	}

	filterExtObj := ua.ExtensionObject{
		EncodingMask: ua.ExtensionObjectBinary,
		TypeID: &ua.ExpandedNodeID{
			NodeID: ua.NewNumericNodeID(0, id.EventFilter_Encoding_DefaultBinary),
		},
		Value: filter,
	}

	return &ua.MonitoredItemCreateRequest{
		ItemToMonitor: &ua.ReadValueID{
			NodeID:       nodeID,
			AttributeID:  ua.AttributeIDEventNotifier,
			DataEncoding: &ua.QualifiedName{},
		},
		MonitoringMode: ua.MonitoringModeReporting,
		RequestedParameters: &ua.MonitoringParameters{
			ClientHandle:     42,
			DiscardOldest:    true,
			Filter:           &filterExtObj,
			QueueSize:        32,
			SamplingInterval: 0,
		},
	}, fieldNames
}

func alarmEventFromFields(node string, fieldNames []string, fields []*ua.Variant) domain.AlarmEvent {
	values := make(map[string]interface{}, len(fieldNames))
	for i, name := range fieldNames {
		if i >= len(fields) {
			continue
		}
		values[name] = eventFieldValue(fields[i])
	}

	event := domain.AlarmEvent{
		NodeID:  node,
		Fields:  values,
		EventID: stringValue(values["EventId"]),
		EventType: stringValue(values["EventType"]),
		SourceNode: stringValue(values["SourceNode"]),
		SourceName: stringValue(values["SourceName"]),
		Time: stringValue(values["Time"]),
		ReceiveTime: stringValue(values["ReceiveTime"]),
		Message: stringValue(values["Message"]),
		Severity: uint16Value(values["Severity"]),
	}
	return event
}

func eventFieldValue(value *ua.Variant) interface{} {
	if value == nil {
		return nil
	}

	raw := value.Value()
	switch typed := raw.(type) {
	case []byte:
		return hex.EncodeToString(typed)
	case time.Time:
		return formatTimestamp(typed)
	default:
		return typed
	}
}

func stringValue(value interface{}) string {
	if value == nil {
		return ""
	}
	return fmt.Sprint(value)
}

func uint16Value(value interface{}) uint16 {
	switch typed := value.(type) {
	case uint16:
		return typed
	case uint32:
		return uint16(typed)
	case uint64:
		return uint16(typed)
	case int:
		return uint16(typed)
	case int32:
		return uint16(typed)
	case int64:
		return uint16(typed)
	default:
		return 0
	}
}
