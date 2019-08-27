package stream

// Message describes a stream message
type Message struct {
	CheckID string `json:"check_id,omitempty"`
	AgentID string `json:"agent_id,omitempty"`
	ScanID  string `json:"scan_id,omitempty"`
	Action  string `json:"action"`
}
