package ceph

type Risk string

const (
	RiskLow    Risk = "low"
	RiskMedium Risk = "medium"
	RiskHigh   Risk = "high"
)

type ActionResult struct {
	ResourceURL string `json:"resource_url,omitempty"`
	Details     any    `json:"details,omitempty"`
}

type ActionError struct {
	Code      string
	Message   string
	Retryable bool
	Details   map[string]any
}

func (e *ActionError) Error() string { return e.Message }
