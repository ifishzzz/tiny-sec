package falco

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	falcoModel "github.com/flipped-aurora/gin-vue-admin/server/model/falco"
	falcoReq "github.com/flipped-aurora/gin-vue-admin/server/model/falco/request"
	falcoRes "github.com/flipped-aurora/gin-vue-admin/server/model/falco/response"
	systemModel "github.com/flipped-aurora/gin-vue-admin/server/model/system"
	"github.com/flipped-aurora/gin-vue-admin/server/utils"
	"gorm.io/gorm"
)

const (
	falcoDefaultEnrollKey     = "falco-enroll-key"
	falcoDefaultEventKeepDays = 7
	falcoDefaultRuleSyncMode  = "manual"
	falcoDefaultFalcoVersion  = "0.40.0"
	falcoDefaultInstallSource = "https://download.falco.org/packages/rpm"
	falcoDefaultServiceName   = "falco"
	falcoDefaultConfigPath    = "/etc/falco/falco.yaml"
	falcoDefaultLogPath       = "/var/log/falco.log"
	falcoDefaultArch          = "x86_64"
	falcoDefaultOS            = "Amazon Linux 2023"

	falcoDefaultHeartbeatInterval    = 30
	falcoDefaultTaskPullInterval     = 15
	falcoDefaultEventUploadBatchSize = 200
)

type FalcoService struct{}

func (s *FalcoService) CreateHost(host *falcoModel.FalcoHost) error {
	host.Name = strings.TrimSpace(host.Name)
	host.Hostname = strings.TrimSpace(host.Hostname)
	host.IP = strings.TrimSpace(host.IP)
	if host.Status == "" {
		host.Status = "offline"
	}
	return global.GVA_DB.Create(host).Error
}

func (s *FalcoService) UpdateHost(host falcoModel.FalcoHost) error {
	return global.GVA_DB.Model(&falcoModel.FalcoHost{}).Where("id = ?", host.ID).Updates(map[string]any{
		"name":          strings.TrimSpace(host.Name),
		"hostname":      strings.TrimSpace(host.Hostname),
		"ip":            strings.TrimSpace(host.IP),
		"provider":      strings.TrimSpace(host.Provider),
		"region":        strings.TrimSpace(host.Region),
		"instance_id":   strings.TrimSpace(host.InstanceID),
		"os":            strings.TrimSpace(host.OS),
		"arch":          strings.TrimSpace(host.Arch),
		"status":        strings.TrimSpace(host.Status),
		"agent_version": strings.TrimSpace(host.AgentVersion),
		"labels":        host.Labels,
		"remarks":       host.Remarks,
	}).Error
}

func (s *FalcoService) DeleteHost(id uint) error {
	return global.GVA_DB.Delete(&falcoModel.FalcoHost{}, id).Error
}

func (s *FalcoService) GetHostList(info falcoReq.FalcoHostSearch) (list []falcoModel.FalcoHost, total int64, err error) {
	db := global.GVA_DB.Model(&falcoModel.FalcoHost{})
	if keyword := strings.TrimSpace(info.Keyword); keyword != "" {
		like := "%" + keyword + "%"
		db = db.Where("name LIKE ? OR hostname LIKE ? OR ip LIKE ? OR instance_id LIKE ?", like, like, like, like)
	}
	if name := strings.TrimSpace(info.Name); name != "" {
		db = db.Where("name LIKE ?", "%"+name+"%")
	}
	if hostname := strings.TrimSpace(info.Hostname); hostname != "" {
		db = db.Where("hostname LIKE ?", "%"+hostname+"%")
	}
	if ip := strings.TrimSpace(info.IP); ip != "" {
		db = db.Where("ip LIKE ?", "%"+ip+"%")
	}
	if status := strings.TrimSpace(info.Status); status != "" {
		db = db.Where("status = ?", status)
	}
	if provider := strings.TrimSpace(info.Provider); provider != "" {
		db = db.Where("provider = ?", provider)
	}
	if region := strings.TrimSpace(info.Region); region != "" {
		db = db.Where("region = ?", region)
	}
	err = db.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	err = db.Scopes((&info.PageInfo).Paginate()).Order("id desc").Find(&list).Error
	return list, total, err
}

func (s *FalcoService) GetAgentList(info falcoReq.FalcoAgentSearch) (list []falcoModel.FalcoAgent, total int64, err error) {
	db := global.GVA_DB.Model(&falcoModel.FalcoAgent{})
	if keyword := strings.TrimSpace(info.Keyword); keyword != "" {
		db = db.Where("agent_id LIKE ?", "%"+keyword+"%")
	}
	if agentID := strings.TrimSpace(info.AgentID); agentID != "" {
		db = db.Where("agent_id LIKE ?", "%"+agentID+"%")
	}
	if status := strings.TrimSpace(info.Status); status != "" {
		db = db.Where("status = ?", status)
	}
	if info.HostID > 0 {
		db = db.Where("host_id = ?", info.HostID)
	}
	err = db.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	err = db.Scopes((&info.PageInfo).Paginate()).Order("id desc").Find(&list).Error
	return list, total, err
}

func (s *FalcoService) RegisterAgent(info falcoReq.FalcoAgentRegister) (falcoRes.FalcoAgentRegisterResult, error) {
	var result falcoRes.FalcoAgentRegisterResult
	expectedEnrollKey := s.getSettingValue("falco_enroll_key", falcoDefaultEnrollKey)
	if strings.TrimSpace(info.EnrollKey) != expectedEnrollKey {
		return result, errors.New("enrollKey 校验失败")
	}
	if strings.TrimSpace(info.AgentID) == "" {
		return result, errors.New("agentId 不能为空")
	}

	now := time.Now()
	host, err := s.findOrCreateHost(info, now)
	if err != nil {
		return result, err
	}

	var agent falcoModel.FalcoAgent
	err = global.GVA_DB.Where("agent_id = ?", strings.TrimSpace(info.AgentID)).First(&agent).Error
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		agent = falcoModel.FalcoAgent{
			HostID:          host.ID,
			AgentID:         strings.TrimSpace(info.AgentID),
			AccessToken:     utils.RandomString(48),
			Version:         strings.TrimSpace(info.Version),
			Status:          "online",
			LastHeartbeatAt: &now,
			LastReportedAt:  &now,
			LastConnectedAt: &now,
			Metadata:        info.Metadata,
		}
		err = global.GVA_DB.Create(&agent).Error
	case err == nil:
		updates := map[string]any{
			"host_id":           host.ID,
			"version":           strings.TrimSpace(info.Version),
			"status":            "online",
			"last_heartbeat_at": now,
			"last_reported_at":  now,
			"metadata":          info.Metadata,
		}
		if strings.TrimSpace(agent.AccessToken) == "" {
			agent.AccessToken = utils.RandomString(48)
			updates["access_token"] = agent.AccessToken
		}
		err = global.GVA_DB.Model(&agent).Updates(updates).Error
	}
	if err != nil {
		return result, err
	}

	result = falcoRes.FalcoAgentRegisterResult{
		HostID:               host.ID,
		AgentID:              agent.AgentID,
		AccessToken:          agent.AccessToken,
		AgentStatus:          "online",
		HeartbeatInterval:    s.getSettingInt("falco_agent_heartbeat_interval", falcoDefaultHeartbeatInterval),
		TaskPullInterval:     s.getSettingInt("falco_agent_task_pull_interval", falcoDefaultTaskPullInterval),
		EventUploadBatchSize: s.getSettingInt("falco_agent_event_upload_batch_size", falcoDefaultEventUploadBatchSize),
		WSPath:               "/falco/agent/ws",
	}
	return result, nil
}

func (s *FalcoService) AgentHeartbeat(info falcoReq.FalcoAgentHeartbeat) error {
	agent, host, err := s.authAgent(info.AgentID, info.AccessToken)
	if err != nil {
		return err
	}
	now := time.Now()
	status := strings.TrimSpace(info.Status)
	if status == "" {
		status = "online"
	}
	if err = global.GVA_DB.Model(&agent).Updates(map[string]any{
		"status":            status,
		"last_heartbeat_at": now,
	}).Error; err != nil {
		return err
	}
	return global.GVA_DB.Model(&host).Updates(map[string]any{
		"status":       status,
		"last_seen_at": now,
	}).Error
}

func (s *FalcoService) AgentStatusReport(info falcoReq.FalcoAgentStatusReport) error {
	agent, host, err := s.authAgent(info.AgentID, info.AccessToken)
	if err != nil {
		return err
	}
	now := time.Now()

	var runtimeStatus falcoModel.FalcoRuntimeStatus
	err = global.GVA_DB.Where("host_id = ?", host.ID).First(&runtimeStatus).Error
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		runtimeStatus = falcoModel.FalcoRuntimeStatus{
			HostID:        host.ID,
			CPUPercent:    info.CPUPercent,
			MemoryPercent: info.MemoryPercent,
			Load1:         info.Load1,
			FalcoStatus:   strings.TrimSpace(info.FalcoStatus),
			EventCount:    info.EventCount,
			ReportedAt:    now,
		}
		err = global.GVA_DB.Create(&runtimeStatus).Error
	case err == nil:
		err = global.GVA_DB.Model(&runtimeStatus).Updates(map[string]any{
			"cpu_percent":    info.CPUPercent,
			"memory_percent": info.MemoryPercent,
			"load1":          info.Load1,
			"falco_status":   strings.TrimSpace(info.FalcoStatus),
			"event_count":    info.EventCount,
			"reported_at":    now,
		}).Error
	}
	if err != nil {
		return err
	}
	if err = global.GVA_DB.Model(&agent).Updates(map[string]any{
		"status":            "online",
		"last_reported_at":  now,
		"last_heartbeat_at": now,
	}).Error; err != nil {
		return err
	}
	return global.GVA_DB.Model(&host).Updates(map[string]any{
		"status":       "online",
		"last_seen_at": now,
	}).Error
}

func (s *FalcoService) BulkCreateEvents(info falcoReq.FalcoEventBulk) error {
	agent, host, err := s.authAgent(info.AgentID, info.AccessToken)
	if err != nil {
		return err
	}
	if len(info.Events) == 0 {
		return nil
	}

	now := time.Now()
	events := make([]falcoModel.FalcoEvent, 0, len(info.Events))
	for _, item := range info.Events {
		eventTime := item.EventTime
		if eventTime.IsZero() {
			eventTime = now
		}
		events = append(events, falcoModel.FalcoEvent{
			HostID:       host.ID,
			AgentID:      agent.AgentID,
			Rule:         strings.TrimSpace(item.Rule),
			Priority:     strings.TrimSpace(item.Priority),
			Source:       strings.TrimSpace(item.Source),
			Output:       item.Output,
			OutputFields: item.OutputFields,
			EventTime:    eventTime,
		})
	}
	if err = global.GVA_DB.Create(&events).Error; err != nil {
		return err
	}
	if err = global.GVA_DB.Model(&agent).Update("last_event_batch_at", now).Error; err != nil {
		return err
	}
	return global.GVA_DB.Model(&host).Updates(map[string]any{
		"status":       "online",
		"last_seen_at": now,
	}).Error
}

func (s *FalcoService) CreateTask(info falcoReq.FalcoTaskCreate) error {
	var host falcoModel.FalcoHost
	if err := global.GVA_DB.First(&host, info.HostID).Error; err != nil {
		return err
	}

	taskType := normalizeTaskType(info.TaskType)
	action, err := taskTypeToAction(taskType)
	if err != nil {
		return err
	}

	task := falcoModel.FalcoInstallTask{
		HostID:    info.HostID,
		RequestID: "req_" + utils.RandomString(16),
		TaskType:  taskType,
		Action:    action,
		Status:    "pending",
		Stage:     "created",
		Operator:  firstNonEmpty(strings.TrimSpace(info.Operator), "console"),
	}
	if err = global.GVA_DB.Create(&task).Error; err != nil {
		return err
	}

	payload, err := s.buildTaskPayload(task, host, info)
	if err != nil {
		return err
	}
	return global.GVA_DB.Model(&task).Update("payload", payload).Error
}

func (s *FalcoService) GetTaskList(info falcoReq.FalcoTaskSearch) (list []falcoModel.FalcoInstallTask, total int64, err error) {
	db := global.GVA_DB.Model(&falcoModel.FalcoInstallTask{})
	if info.HostID > 0 {
		db = db.Where("host_id = ?", info.HostID)
	}
	if taskType := strings.TrimSpace(info.TaskType); taskType != "" {
		db = db.Where("task_type = ?", taskType)
	}
	if status := strings.TrimSpace(info.Status); status != "" {
		db = db.Where("status = ?", status)
	}
	err = db.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	err = db.Scopes((&info.PageInfo).Paginate()).Order("id desc").Find(&list).Error
	return list, total, err
}

func (s *FalcoService) PullAgentTasks(info falcoReq.FalcoAgentAuth) ([]falcoRes.FalcoTaskDispatch, error) {
	agent, _, err := s.authAgent(info.AgentID, info.AccessToken)
	if err != nil {
		return nil, err
	}
	var list []falcoModel.FalcoInstallTask
	if err = global.GVA_DB.Where("host_id = ? AND status = ?", agent.HostID, "pending").Order("id asc").Limit(20).Find(&list).Error; err != nil {
		return nil, err
	}
	if len(list) == 0 {
		return []falcoRes.FalcoTaskDispatch{}, nil
	}

	now := time.Now()
	dispatches := make([]falcoRes.FalcoTaskDispatch, 0, len(list))
	for _, task := range list {
		_ = global.GVA_DB.Model(&falcoModel.FalcoInstallTask{}).Where("id = ?", task.ID).Updates(map[string]any{
			"status":      "dispatched",
			"stage":       "delivered",
			"executed_at": now,
		}).Error

		payload := map[string]any{}
		if strings.TrimSpace(task.Payload) != "" {
			_ = json.Unmarshal([]byte(task.Payload), &payload)
		}
		dispatches = append(dispatches, falcoRes.FalcoTaskDispatch{
			TaskID:    task.ID,
			RequestID: task.RequestID,
			TaskType:  task.TaskType,
			Action:    task.Action,
			Stage:     "delivered",
			Payload:   payload,
		})
	}
	return dispatches, nil
}

func (s *FalcoService) SaveTaskResult(info falcoReq.FalcoTaskResult) error {
	_, host, err := s.authAgent(info.AgentID, info.AccessToken)
	if err != nil {
		return err
	}
	status := strings.TrimSpace(info.Status)
	if status == "" {
		status = "succeeded"
	}
	stage := firstNonEmpty(strings.TrimSpace(info.Stage), "finished")
	now := time.Now()

	result := strings.TrimSpace(info.Result)
	if result == "" {
		resultJSON, marshalErr := json.Marshal(map[string]any{
			"falcoVersion":       strings.TrimSpace(info.FalcoVersion),
			"rulePackageVersion": strings.TrimSpace(info.RulePackageVersion),
			"serviceState":       strings.TrimSpace(info.ServiceState),
		})
		if marshalErr == nil {
			result = string(resultJSON)
		}
	}

	if err = global.GVA_DB.Model(&falcoModel.FalcoInstallTask{}).Where("id = ?", info.TaskID).Updates(map[string]any{
		"status":         status,
		"stage":          stage,
		"result":         result,
		"stdout_summary": info.StdoutSummary,
		"stderr_summary": info.StderrSummary,
		"error_code":     strings.TrimSpace(info.ErrorCode),
		"error_message":  info.ErrorMessage,
		"finished_at":    now,
	}).Error; err != nil {
		return err
	}

	if strings.TrimSpace(info.ServiceState) != "" {
		var runtimeStatus falcoModel.FalcoRuntimeStatus
		err = global.GVA_DB.Where("host_id = ?", host.ID).First(&runtimeStatus).Error
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			runtimeStatus = falcoModel.FalcoRuntimeStatus{
				HostID:      host.ID,
				FalcoStatus: strings.TrimSpace(info.ServiceState),
				ReportedAt:  now,
			}
			err = global.GVA_DB.Create(&runtimeStatus).Error
		case err == nil:
			err = global.GVA_DB.Model(&runtimeStatus).Updates(map[string]any{
				"falco_status": strings.TrimSpace(info.ServiceState),
				"reported_at":  now,
			}).Error
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *FalcoService) CreateRulePackage(info falcoReq.FalcoRulePackageCreate) error {
	pkg := falcoModel.FalcoRulePackage{
		Name:        strings.TrimSpace(info.Name),
		Version:     strings.TrimSpace(info.Version),
		Status:      firstNonEmpty(strings.TrimSpace(info.Status), "draft"),
		Description: info.Description,
		Content:     info.Content,
		Checksum:    strings.TrimSpace(info.Checksum),
	}
	return global.GVA_DB.Create(&pkg).Error
}

func (s *FalcoService) GetRulePackageList(info falcoReq.FalcoRulePackageSearch) (list []falcoModel.FalcoRulePackage, total int64, err error) {
	db := global.GVA_DB.Model(&falcoModel.FalcoRulePackage{})
	if keyword := strings.TrimSpace(info.Keyword); keyword != "" {
		db = db.Where("name LIKE ? OR version LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	if name := strings.TrimSpace(info.Name); name != "" {
		db = db.Where("name LIKE ?", "%"+name+"%")
	}
	if status := strings.TrimSpace(info.Status); status != "" {
		db = db.Where("status = ?", status)
	}
	err = db.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	err = db.Scopes((&info.PageInfo).Paginate()).Order("id desc").Find(&list).Error
	return list, total, err
}

func (s *FalcoService) PublishRulePackage(info falcoReq.FalcoRulePublishCreate) error {
	if info.RulePackageID == 0 || len(info.HostIDs) == 0 {
		return errors.New("规则发布参数不完整")
	}

	now := time.Now()
	return global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		for _, hostID := range info.HostIDs {
			publish := falcoModel.FalcoRulePublish{
				RulePackageID: info.RulePackageID,
				HostID:        hostID,
				Status:        "pending",
				PublishedAt:   &now,
			}
			if err := tx.Create(&publish).Error; err != nil {
				return err
			}

			task := falcoModel.FalcoInstallTask{
				HostID:    hostID,
				RequestID: "req_" + utils.RandomString(16),
				TaskType:  "rule_publish",
				Action:    "falco.rule_publish",
				Status:    "pending",
				Stage:     "created",
				Operator:  "console",
				Payload:   fmt.Sprintf("{\"rulePackageId\":%d}", info.RulePackageID),
			}
			if err := tx.Create(&task).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *FalcoService) GetPublishList(info falcoReq.FalcoPublishSearch) (list []falcoModel.FalcoRulePublish, total int64, err error) {
	db := global.GVA_DB.Model(&falcoModel.FalcoRulePublish{})
	if info.HostID > 0 {
		db = db.Where("host_id = ?", info.HostID)
	}
	if info.RulePackageID > 0 {
		db = db.Where("rule_package_id = ?", info.RulePackageID)
	}
	if status := strings.TrimSpace(info.Status); status != "" {
		db = db.Where("status = ?", status)
	}
	err = db.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	err = db.Scopes((&info.PageInfo).Paginate()).Order("id desc").Find(&list).Error
	return list, total, err
}

func (s *FalcoService) GetEventList(info falcoReq.FalcoEventSearch) (list []falcoModel.FalcoEvent, total int64, err error) {
	db := global.GVA_DB.Model(&falcoModel.FalcoEvent{})
	if info.HostID > 0 {
		db = db.Where("host_id = ?", info.HostID)
	}
	if agentID := strings.TrimSpace(info.AgentID); agentID != "" {
		db = db.Where("agent_id = ?", agentID)
	}
	if priority := strings.TrimSpace(info.Priority); priority != "" {
		db = db.Where("priority = ?", priority)
	}
	if rule := strings.TrimSpace(info.Rule); rule != "" {
		db = db.Where("rule LIKE ?", "%"+rule+"%")
	}
	if info.StartEventTime != nil {
		db = db.Where("event_time >= ?", *info.StartEventTime)
	}
	if info.EndEventTime != nil {
		db = db.Where("event_time <= ?", *info.EndEventTime)
	}
	err = db.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	err = db.Scopes((&info.PageInfo).Paginate()).Order("event_time desc").Find(&list).Error
	return list, total, err
}

func (s *FalcoService) GetDashboard() (falcoRes.FalcoDashboard, error) {
	dashboard := falcoRes.FalcoDashboard{}
	if err := global.GVA_DB.Model(&falcoModel.FalcoHost{}).Count(&dashboard.HostTotal).Error; err != nil {
		return dashboard, err
	}
	if err := global.GVA_DB.Model(&falcoModel.FalcoHost{}).Where("status = ?", "online").Count(&dashboard.OnlineHostTotal).Error; err != nil {
		return dashboard, err
	}
	if err := global.GVA_DB.Model(&falcoModel.FalcoAgent{}).Count(&dashboard.AgentTotal).Error; err != nil {
		return dashboard, err
	}
	if err := global.GVA_DB.Model(&falcoModel.FalcoInstallTask{}).Where("status IN ?", []string{"pending", "running"}).Count(&dashboard.PendingTaskTotal).Error; err != nil {
		return dashboard, err
	}
	if err := global.GVA_DB.Model(&falcoModel.FalcoRulePackage{}).Count(&dashboard.RulePackageTotal).Error; err != nil {
		return dashboard, err
	}
	if err := global.GVA_DB.Model(&falcoModel.FalcoEvent{}).Count(&dashboard.EventTotal).Error; err != nil {
		return dashboard, err
	}
	if err := global.GVA_DB.Model(&falcoModel.FalcoEvent{}).Where("priority IN ?", []string{"critical", "emergency"}).Count(&dashboard.CriticalEventTotal).Error; err != nil {
		return dashboard, err
	}
	return dashboard, nil
}

func (s *FalcoService) GetSettings() falcoRes.FalcoSettings {
	return falcoRes.FalcoSettings{
		EnrollKey:     s.getSettingValue("falco_enroll_key", falcoDefaultEnrollKey),
		EventKeepDays: s.getSettingInt("falco_event_keep_days", falcoDefaultEventKeepDays),
		RuleSyncMode:  s.getSettingValue("falco_rule_sync_mode", falcoDefaultRuleSyncMode),
	}
}

func (s *FalcoService) UpdateSettings(info falcoReq.FalcoSettingsUpdate) error {
	if strings.TrimSpace(info.EnrollKey) == "" {
		info.EnrollKey = falcoDefaultEnrollKey
	}
	if info.EventKeepDays <= 0 {
		info.EventKeepDays = falcoDefaultEventKeepDays
	}
	if strings.TrimSpace(info.RuleSyncMode) == "" {
		info.RuleSyncMode = falcoDefaultRuleSyncMode
	}

	if err := s.upsertSetting("falco_enroll_key", "Falco 注册密钥", strings.TrimSpace(info.EnrollKey), "Falco Agent 首次注册使用的 enrollKey"); err != nil {
		return err
	}
	if err := s.upsertSetting("falco_event_keep_days", "Falco 事件保留天数", strconv.Itoa(info.EventKeepDays), "Falco 事件保留天数"); err != nil {
		return err
	}
	if err := s.upsertSetting("falco_rule_sync_mode", "Falco 规则同步模式", strings.TrimSpace(info.RuleSyncMode), "Falco 规则同步模式"); err != nil {
		return err
	}
	return nil
}

func (s *FalcoService) findOrCreateHost(info falcoReq.FalcoAgentRegister, now time.Time) (falcoModel.FalcoHost, error) {
	db := global.GVA_DB.Model(&falcoModel.FalcoHost{})
	var host falcoModel.FalcoHost
	var err error

	if instanceID := strings.TrimSpace(info.InstanceID); instanceID != "" {
		err = db.Where("instance_id = ?", instanceID).First(&host).Error
	} else {
		err = db.Where("hostname = ? AND ip = ?", strings.TrimSpace(info.Hostname), strings.TrimSpace(info.IP)).First(&host).Error
	}

	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		host = falcoModel.FalcoHost{
			Name:         firstNonEmpty(strings.TrimSpace(info.Hostname), strings.TrimSpace(info.InstanceID), strings.TrimSpace(info.AgentID)),
			Hostname:     strings.TrimSpace(info.Hostname),
			IP:           strings.TrimSpace(info.IP),
			Provider:     strings.TrimSpace(info.Provider),
			Region:       strings.TrimSpace(info.Region),
			InstanceID:   strings.TrimSpace(info.InstanceID),
			OS:           strings.TrimSpace(info.OS),
			Arch:         strings.TrimSpace(info.Arch),
			Status:       "online",
			AgentVersion: strings.TrimSpace(info.Version),
			LastSeenAt:   &now,
			Labels:       info.Labels,
		}
		err = global.GVA_DB.Create(&host).Error
	case err == nil:
		err = global.GVA_DB.Model(&host).Updates(map[string]any{
			"name":          firstNonEmpty(strings.TrimSpace(info.Hostname), strings.TrimSpace(host.Name), strings.TrimSpace(info.AgentID)),
			"hostname":      strings.TrimSpace(info.Hostname),
			"ip":            strings.TrimSpace(info.IP),
			"provider":      strings.TrimSpace(info.Provider),
			"region":        strings.TrimSpace(info.Region),
			"instance_id":   strings.TrimSpace(info.InstanceID),
			"os":            strings.TrimSpace(info.OS),
			"arch":          strings.TrimSpace(info.Arch),
			"status":        "online",
			"agent_version": strings.TrimSpace(info.Version),
			"last_seen_at":  now,
			"labels":        info.Labels,
		}).Error
	}
	return host, err
}

func (s *FalcoService) authAgent(agentID string, accessToken string) (falcoModel.FalcoAgent, falcoModel.FalcoHost, error) {
	var agent falcoModel.FalcoAgent
	var host falcoModel.FalcoHost
	err := global.GVA_DB.Where("agent_id = ? AND access_token = ?", strings.TrimSpace(agentID), strings.TrimSpace(accessToken)).First(&agent).Error
	if err != nil {
		return agent, host, errors.New("agent 鉴权失败")
	}
	err = global.GVA_DB.First(&host, agent.HostID).Error
	return agent, host, err
}

func (s *FalcoService) getSettingValue(key string, defaultValue string) string {
	var param systemModel.SysParams
	err := global.GVA_DB.Where("`key` = ?", key).First(&param).Error
	if err != nil || strings.TrimSpace(param.Value) == "" {
		return defaultValue
	}
	return strings.TrimSpace(param.Value)
}

func (s *FalcoService) getSettingInt(key string, defaultValue int) int {
	value := s.getSettingValue(key, "")
	if value == "" {
		return defaultValue
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return parsed
}

func (s *FalcoService) upsertSetting(key string, name string, value string, desc string) error {
	var param systemModel.SysParams
	err := global.GVA_DB.Where("`key` = ?", key).First(&param).Error
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		param = systemModel.SysParams{Name: name, Key: key, Value: value, Desc: desc}
		return global.GVA_DB.Create(&param).Error
	case err != nil:
		return err
	default:
		return global.GVA_DB.Model(&param).Updates(map[string]any{
			"name":  name,
			"value": value,
			"desc":  desc,
		}).Error
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func resolveHostOS(osName string) (string, string) {
	value := strings.TrimSpace(strings.ToLower(osName))
	switch {
	case strings.Contains(value, "amazon linux 2023"):
		return "amzn", "2023"
	case strings.Contains(value, "amazon linux"):
		return "amzn", strings.TrimSpace(strings.TrimPrefix(value, "amazon linux"))
	case value == "":
		return "amzn", "2023"
	default:
		return value, value
	}
}

func normalizeTaskType(taskType string) string {
	value := strings.ToLower(strings.TrimSpace(taskType))
	switch value {
	case "falco.install":
		return "install"
	case "falco.upgrade":
		return "upgrade"
	case "falco.rollback":
		return "rollback"
	case "falco.reload":
		return "reload"
	case "falco.restart":
		return "restart"
	default:
		return value
	}
}

func taskTypeToAction(taskType string) (string, error) {
	switch normalizeTaskType(taskType) {
	case "install":
		return "falco.install", nil
	case "upgrade":
		return "falco.upgrade", nil
	case "rollback":
		return "falco.rollback", nil
	case "reload":
		return "falco.reload", nil
	case "restart":
		return "falco.restart", nil
	default:
		return "", errors.New("暂不支持的任务类型")
	}
}

func (s *FalcoService) buildTaskPayload(task falcoModel.FalcoInstallTask, host falcoModel.FalcoHost, info falcoReq.FalcoTaskCreate) (string, error) {
	if strings.TrimSpace(info.Payload) != "" {
		return info.Payload, nil
	}

	osFamily, osVersion := resolveHostOS(host.OS)
	payload := map[string]any{
		"taskId":                task.ID,
		"requestId":             task.RequestID,
		"action":                task.Action,
		"targetType":            "host",
		"targetId":              task.HostID,
		"hostname":              host.Hostname,
		"cloudProvider":         firstNonEmpty(strings.TrimSpace(host.Provider), "aws"),
		"cloudRegion":           host.Region,
		"osFamily":              osFamily,
		"osVersion":             osVersion,
		"arch":                  firstNonEmpty(strings.TrimSpace(host.Arch), falcoDefaultArch),
		"packageManager":        "dnf",
		"falcoVersion":          firstNonEmpty(strings.TrimSpace(info.FalcoVersion), falcoDefaultFalcoVersion),
		"installChannel":        firstNonEmpty(strings.TrimSpace(info.InstallChannel), "stable"),
		"driverMode":            firstNonEmpty(strings.TrimSpace(info.DriverMode), "modern_ebpf"),
		"configTemplateVersion": firstNonEmpty(strings.TrimSpace(info.ConfigTemplateVersion), "v1"),
		"rulePackageVersion":    firstNonEmpty(strings.TrimSpace(info.RulePackageVersion), "default"),
		"downloadSource":        firstNonEmpty(strings.TrimSpace(info.DownloadSource), falcoDefaultInstallSource),
		"checksum":              strings.TrimSpace(info.Checksum),
		"serviceName":           firstNonEmpty(strings.TrimSpace(info.ServiceName), falcoDefaultServiceName),
		"configPath":            firstNonEmpty(strings.TrimSpace(info.ConfigPath), falcoDefaultConfigPath),
		"logPath":               firstNonEmpty(strings.TrimSpace(info.LogPath), falcoDefaultLogPath),
	}

	bytes, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}
