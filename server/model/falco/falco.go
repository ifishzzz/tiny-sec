package falco

import (
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
)

type FalcoHost struct {
	global.GVA_MODEL
	Name         string     `json:"name" form:"name" gorm:"comment:主机名称"`
	Hostname     string     `json:"hostname" form:"hostname" gorm:"index;comment:主机名"`
	IP           string     `json:"ip" form:"ip" gorm:"index;comment:主机IP"`
	Provider     string     `json:"provider" form:"provider" gorm:"comment:云厂商"`
	Region       string     `json:"region" form:"region" gorm:"comment:区域"`
	InstanceID   string     `json:"instanceId" form:"instanceId" gorm:"index;comment:云实例ID"`
	OS           string     `json:"os" form:"os" gorm:"comment:操作系统"`
	Arch         string     `json:"arch" form:"arch" gorm:"comment:架构"`
	Status       string     `json:"status" form:"status" gorm:"default:offline;comment:主机状态"`
	AgentVersion string     `json:"agentVersion" form:"agentVersion" gorm:"comment:Agent版本"`
	LastSeenAt   *time.Time `json:"lastSeenAt" form:"lastSeenAt" gorm:"comment:最后在线时间"`
	Labels       string     `json:"labels" form:"labels" gorm:"type:text;comment:标签JSON"`
	Remarks      string     `json:"remarks" form:"remarks" gorm:"type:text;comment:备注"`
}

func (FalcoHost) TableName() string {
	return "falco_hosts"
}

type FalcoAgent struct {
	global.GVA_MODEL
	HostID           uint       `json:"hostId" form:"hostId" gorm:"index;comment:关联主机ID"`
	AgentID          string     `json:"agentId" form:"agentId" gorm:"uniqueIndex;comment:Agent标识"`
	AccessToken      string     `json:"-" gorm:"size:128;comment:接入令牌"`
	Version          string     `json:"version" form:"version" gorm:"comment:Agent版本"`
	Status           string     `json:"status" form:"status" gorm:"default:offline;comment:Agent状态"`
	LastHeartbeatAt  *time.Time `json:"lastHeartbeatAt" form:"lastHeartbeatAt" gorm:"comment:最后心跳时间"`
	LastReportedAt   *time.Time `json:"lastReportedAt" form:"lastReportedAt" gorm:"comment:最后状态上报时间"`
	LastConnectedAt  *time.Time `json:"lastConnectedAt" form:"lastConnectedAt" gorm:"comment:首次注册时间"`
	LastEventBatchAt *time.Time `json:"lastEventBatchAt" form:"lastEventBatchAt" gorm:"comment:最后事件上报时间"`
	Metadata         string     `json:"metadata" form:"metadata" gorm:"type:text;comment:元数据JSON"`
}

func (FalcoAgent) TableName() string {
	return "falco_agents"
}

type FalcoInstallTask struct {
	global.GVA_MODEL
	HostID        uint       `json:"hostId" form:"hostId" gorm:"index;comment:目标主机ID"`
	RequestID     string     `json:"requestId" form:"requestId" gorm:"index;comment:请求标识"`
	TaskType      string     `json:"taskType" form:"taskType" gorm:"index;comment:任务类型"`
	Action        string     `json:"action" form:"action" gorm:"index;comment:任务动作"`
	Status        string     `json:"status" form:"status" gorm:"default:pending;index;comment:任务状态"`
	Stage         string     `json:"stage" form:"stage" gorm:"comment:任务阶段"`
	Operator      string     `json:"operator" form:"operator" gorm:"comment:操作人"`
	Payload       string     `json:"payload" form:"payload" gorm:"type:text;comment:任务负载JSON"`
	Result        string     `json:"result" form:"result" gorm:"type:text;comment:任务结果"`
	StdoutSummary string     `json:"stdoutSummary" form:"stdoutSummary" gorm:"type:text;comment:标准输出摘要"`
	StderrSummary string     `json:"stderrSummary" form:"stderrSummary" gorm:"type:text;comment:标准错误摘要"`
	ErrorCode     string     `json:"errorCode" form:"errorCode" gorm:"comment:错误码"`
	ErrorMessage  string     `json:"errorMessage" form:"errorMessage" gorm:"type:text;comment:错误信息"`
	ExecutedAt    *time.Time `json:"executedAt" form:"executedAt" gorm:"comment:开始执行时间"`
	FinishedAt    *time.Time `json:"finishedAt" form:"finishedAt" gorm:"comment:完成时间"`
}

func (FalcoInstallTask) TableName() string {
	return "falco_install_tasks"
}

type FalcoRulePackage struct {
	global.GVA_MODEL
	Name        string `json:"name" form:"name" gorm:"index;comment:规则包名称"`
	Version     string `json:"version" form:"version" gorm:"comment:规则包版本"`
	Status      string `json:"status" form:"status" gorm:"default:draft;comment:规则包状态"`
	Description string `json:"description" form:"description" gorm:"type:text;comment:说明"`
	Content     string `json:"content" form:"content" gorm:"type:longtext;comment:规则内容"`
	Checksum    string `json:"checksum" form:"checksum" gorm:"comment:校验值"`
}

func (FalcoRulePackage) TableName() string {
	return "falco_rule_packages"
}

type FalcoRulePublish struct {
	global.GVA_MODEL
	RulePackageID uint       `json:"rulePackageId" form:"rulePackageId" gorm:"index;comment:规则包ID"`
	HostID        uint       `json:"hostId" form:"hostId" gorm:"index;comment:主机ID"`
	Status        string     `json:"status" form:"status" gorm:"default:pending;comment:发布状态"`
	Result        string     `json:"result" form:"result" gorm:"type:text;comment:发布结果"`
	PublishedAt   *time.Time `json:"publishedAt" form:"publishedAt" gorm:"comment:发布时间"`
}

func (FalcoRulePublish) TableName() string {
	return "falco_rule_publishes"
}

type FalcoRuntimeStatus struct {
	global.GVA_MODEL
	HostID        uint      `json:"hostId" form:"hostId" gorm:"index;comment:主机ID"`
	CPUPercent    float64   `json:"cpuPercent" form:"cpuPercent" gorm:"comment:CPU使用率"`
	MemoryPercent float64   `json:"memoryPercent" form:"memoryPercent" gorm:"comment:内存使用率"`
	Load1         float64   `json:"load1" form:"load1" gorm:"comment:1分钟负载"`
	FalcoStatus   string    `json:"falcoStatus" form:"falcoStatus" gorm:"comment:Falco状态"`
	EventCount    int64     `json:"eventCount" form:"eventCount" gorm:"comment:事件累计数"`
	ReportedAt    time.Time `json:"reportedAt" form:"reportedAt" gorm:"index;comment:上报时间"`
}

func (FalcoRuntimeStatus) TableName() string {
	return "falco_runtime_statuses"
}

type FalcoEvent struct {
	global.GVA_MODEL
	HostID       uint      `json:"hostId" form:"hostId" gorm:"index;comment:主机ID"`
	AgentID      string    `json:"agentId" form:"agentId" gorm:"index;comment:Agent标识"`
	Rule         string    `json:"rule" form:"rule" gorm:"index;comment:规则名"`
	Priority     string    `json:"priority" form:"priority" gorm:"index;comment:事件等级"`
	Source       string    `json:"source" form:"source" gorm:"comment:事件源"`
	Output       string    `json:"output" form:"output" gorm:"type:text;comment:事件输出"`
	OutputFields string    `json:"outputFields" form:"outputFields" gorm:"type:longtext;comment:结构化字段JSON"`
	EventTime    time.Time `json:"eventTime" form:"eventTime" gorm:"index;comment:事件时间"`
}

func (FalcoEvent) TableName() string {
	return "falco_events"
}
