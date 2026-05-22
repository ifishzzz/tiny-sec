package response

type FalcoDashboard struct {
	HostTotal          int64 `json:"hostTotal"`
	OnlineHostTotal    int64 `json:"onlineHostTotal"`
	AgentTotal         int64 `json:"agentTotal"`
	PendingTaskTotal   int64 `json:"pendingTaskTotal"`
	RulePackageTotal   int64 `json:"rulePackageTotal"`
	EventTotal         int64 `json:"eventTotal"`
	CriticalEventTotal int64 `json:"criticalEventTotal"`
}

type FalcoSettings struct {
	EnrollKey     string `json:"enrollKey"`
	EventKeepDays int    `json:"eventKeepDays"`
	RuleSyncMode  string `json:"ruleSyncMode"`
}

type FalcoAgentRegisterResult struct {
	HostID               uint   `json:"hostId"`
	AgentID              string `json:"agentId"`
	AccessToken          string `json:"accessToken"`
	AgentStatus          string `json:"agentStatus"`
	HeartbeatInterval    int    `json:"heartbeatInterval"`
	TaskPullInterval     int    `json:"taskPullInterval"`
	EventUploadBatchSize int    `json:"eventUploadBatchSize"`
	WSPath               string `json:"wsPath"`
}

type FalcoTaskDispatch struct {
	TaskID    uint           `json:"taskId"`
	RequestID string         `json:"requestId"`
	TaskType  string         `json:"taskType"`
	Action    string         `json:"action"`
	Stage     string         `json:"stage"`
	Payload   map[string]any `json:"payload"`
}
