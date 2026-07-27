package ceph

import "time"

type Risk string

const (
	RiskLow    Risk = "low"
	RiskMedium Risk = "medium"
	RiskHigh   Risk = "high"
)

type OperationResult struct {
	ResourceURL string `json:"resource_url,omitempty"`
	Details     any    `json:"details,omitempty"`
}
type OperationError struct {
	Code      string
	Message   string
	Retryable bool
	Details   map[string]any
}

func (e *OperationError) Error() string { return e.Message }

type OperationView struct {
	ID           string     `json:"operation_id"`
	Action       string     `json:"kind"`
	ClusterID    *uint64    `json:"cluster_id"`
	ResourceType string     `json:"resource_type"`
	ResourceKey  string     `json:"resource_key"`
	Status       string     `json:"status"`
	Stage        string     `json:"stage"`
	Progress     int        `json:"progress"`
	Risk         string     `json:"risk"`
	CreatedAt    time.Time  `json:"created_at"`
	StartedAt    *time.Time `json:"started_at,omitempty"`
	CompletedAt  *time.Time `json:"completed_at,omitempty"`
	Result       any        `json:"result,omitempty"`
	Error        any        `json:"error,omitempty"`
}
