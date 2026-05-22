package request

import (
	"time"

	commonReq "github.com/flipped-aurora/gin-vue-admin/server/model/common/request"
)

type FalcoHostSearch struct {
	commonReq.PageInfo
	Name     string `json:"name" form:"name"`
	Hostname string `json:"hostname" form:"hostname"`
	IP       string `json:"ip" form:"ip"`
	Status   string `json:"status" form:"status"`
	Provider string `json:"provider" form:"provider"`
	Region   string `json:"region" form:"region"`
}

type FalcoAgentSearch struct {
	commonReq.PageInfo
	AgentID string `json:"agentId" form:"agentId"`
	Status  string `json:"status" form:"status"`
	HostID  uint   `json:"hostId" form:"hostId"`
}

type FalcoTaskSearch struct {
	commonReq.PageInfo
	HostID   uint   `json:"hostId" form:"hostId"`
	TaskType string `json:"taskType" form:"taskType"`
	Status   string `json:"status" form:"status"`
}

type FalcoRulePackageSearch struct {
	commonReq.PageInfo
	Name   string `json:"name" form:"name"`
	Status string `json:"status" form:"status"`
}

type FalcoPublishSearch struct {
	commonReq.PageInfo
	HostID        uint   `json:"hostId" form:"hostId"`
	RulePackageID uint   `json:"rulePackageId" form:"rulePackageId"`
	Status        string `json:"status" form:"status"`
}

type FalcoEventSearch struct {
	commonReq.PageInfo
	HostID         uint       `json:"hostId" form:"hostId"`
	AgentID        string     `json:"agentId" form:"agentId"`
	Priority       string     `json:"priority" form:"priority"`
	Rule           string     `json:"rule" form:"rule"`
	StartEventTime *time.Time `json:"startEventTime" form:"startEventTime"`
	EndEventTime   *time.Time `json:"endEventTime" form:"endEventTime"`
}

type FalcoAgentRegister struct {
	EnrollKey  string `json:"enrollKey" form:"enrollKey"`
	AgentID    string `json:"agentId" form:"agentId"`
	Hostname   string `json:"hostname" form:"hostname"`
	IP         string `json:"ip" form:"ip"`
	InstanceID string `json:"instanceId" form:"instanceId"`
	Provider   string `json:"provider" form:"provider"`
	Region     string `json:"region" form:"region"`
	OS         string `json:"os" form:"os"`
	Arch       string `json:"arch" form:"arch"`
	Version    string `json:"version" form:"version"`
	Labels     string `json:"labels" form:"labels"`
	Metadata   string `json:"metadata" form:"metadata"`
}

type FalcoAgentHeartbeat struct {
	AgentID     string `json:"agentId" form:"agentId"`
	AccessToken string `json:"accessToken" form:"accessToken"`
	Status      string `json:"status" form:"status"`
}

type FalcoAgentStatusReport struct {
	AgentID       string  `json:"agentId" form:"agentId"`
	AccessToken   string  `json:"accessToken" form:"accessToken"`
	CPUPercent    float64 `json:"cpuPercent" form:"cpuPercent"`
	MemoryPercent float64 `json:"memoryPercent" form:"memoryPercent"`
	Load1         float64 `json:"load1" form:"load1"`
	FalcoStatus   string  `json:"falcoStatus" form:"falcoStatus"`
	EventCount    int64   `json:"eventCount" form:"eventCount"`
}

type FalcoAgentAuth struct {
	AgentID     string `json:"agentId" form:"agentId"`
	AccessToken string `json:"accessToken" form:"accessToken"`
}

type FalcoTaskCreate struct {
	HostID                uint   `json:"hostId" form:"hostId"`
	TaskType              string `json:"taskType" form:"taskType"`
	FalcoVersion          string `json:"falcoVersion" form:"falcoVersion"`
	RulePackageVersion    string `json:"rulePackageVersion" form:"rulePackageVersion"`
	ConfigTemplateVersion string `json:"configTemplateVersion" form:"configTemplateVersion"`
	InstallChannel        string `json:"installChannel" form:"installChannel"`
	DriverMode            string `json:"driverMode" form:"driverMode"`
	DownloadSource        string `json:"downloadSource" form:"downloadSource"`
	Checksum              string `json:"checksum" form:"checksum"`
	ServiceName           string `json:"serviceName" form:"serviceName"`
	ConfigPath            string `json:"configPath" form:"configPath"`
	LogPath               string `json:"logPath" form:"logPath"`
	Operator              string `json:"operator" form:"operator"`
	Payload               string `json:"payload" form:"payload"`
}

type FalcoTaskResult struct {
	TaskID             uint   `json:"taskId" form:"taskId"`
	AgentID            string `json:"agentId" form:"agentId"`
	AccessToken        string `json:"accessToken" form:"accessToken"`
	Status             string `json:"status" form:"status"`
	Stage              string `json:"stage" form:"stage"`
	Result             string `json:"result" form:"result"`
	StdoutSummary      string `json:"stdoutSummary" form:"stdoutSummary"`
	StderrSummary      string `json:"stderrSummary" form:"stderrSummary"`
	FalcoVersion       string `json:"falcoVersion" form:"falcoVersion"`
	RulePackageVersion string `json:"rulePackageVersion" form:"rulePackageVersion"`
	ServiceState       string `json:"serviceState" form:"serviceState"`
	ErrorCode          string `json:"errorCode" form:"errorCode"`
	ErrorMessage       string `json:"errorMessage" form:"errorMessage"`
}

type FalcoRulePackageCreate struct {
	Name        string `json:"name" form:"name"`
	Version     string `json:"version" form:"version"`
	Status      string `json:"status" form:"status"`
	Description string `json:"description" form:"description"`
	Content     string `json:"content" form:"content"`
	Checksum    string `json:"checksum" form:"checksum"`
}

type FalcoRulePublishCreate struct {
	RulePackageID uint   `json:"rulePackageId" form:"rulePackageId"`
	HostIDs       []uint `json:"hostIds" form:"hostIds"`
}

type FalcoEventItem struct {
	Rule         string    `json:"rule" form:"rule"`
	Priority     string    `json:"priority" form:"priority"`
	Source       string    `json:"source" form:"source"`
	Output       string    `json:"output" form:"output"`
	OutputFields string    `json:"outputFields" form:"outputFields"`
	EventTime    time.Time `json:"eventTime" form:"eventTime"`
}

type FalcoEventBulk struct {
	AgentID     string           `json:"agentId" form:"agentId"`
	AccessToken string           `json:"accessToken" form:"accessToken"`
	Events      []FalcoEventItem `json:"events" form:"events"`
}

type FalcoSettingsUpdate struct {
	EnrollKey     string `json:"enrollKey" form:"enrollKey"`
	EventKeepDays int    `json:"eventKeepDays" form:"eventKeepDays"`
	RuleSyncMode  string `json:"ruleSyncMode" form:"ruleSyncMode"`
}
