package domain

// Endpoint describes one OPC UA endpoint exposed by a server.
type Endpoint struct {
	EndpointURL    string   `json:"endpointUrl"`
	SecurityPolicy string   `json:"securityPolicy"`
	SecurityMode   string   `json:"securityMode"`
	UserTokens     []string `json:"userTokens"`
}

// Node describes a browsed OPC UA node.
type Node struct {
	NodeID      string `json:"nodeId"`
	NodeClass   string `json:"nodeClass"`
	BrowseName  string `json:"browseName,omitempty"`
	DisplayName string `json:"displayName,omitempty"`
	DataType    string `json:"dataType,omitempty"`
	Writable    bool   `json:"writable"`
	Path        string `json:"path,omitempty"`
}

// ReadResult is the result of reading a node value.
type ReadResult struct {
	NodeID          string      `json:"nodeId"`
	Value           interface{} `json:"value"`
	Status          string      `json:"status"`
	SourceTimestamp string      `json:"sourceTimestamp,omitempty"`
	ServerTimestamp string      `json:"serverTimestamp,omitempty"`
}

// WriteResult is the result of writing a node value.
type WriteResult struct {
	NodeID string `json:"nodeId"`
	Status string `json:"status"`
}

// DataChange is emitted by a subscription when a node value changes.
type DataChange struct {
	NodeID          string      `json:"nodeId"`
	Value           interface{} `json:"value"`
	SourceTimestamp string      `json:"sourceTimestamp,omitempty"`
}

// AlarmEvent is emitted by an OPC UA event subscription.
type AlarmEvent struct {
	NodeID     string                 `json:"nodeId"`
	EventID    string                 `json:"eventId,omitempty"`
	EventType  string                 `json:"eventType,omitempty"`
	SourceNode string                 `json:"sourceNode,omitempty"`
	SourceName string                 `json:"sourceName,omitempty"`
	Time       string                 `json:"time,omitempty"`
	ReceiveTime string               `json:"receiveTime,omitempty"`
	Message    string                 `json:"message,omitempty"`
	Severity   uint16                 `json:"severity,omitempty"`
	Fields     map[string]interface{} `json:"fields,omitempty"`
}
