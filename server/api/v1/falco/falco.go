package falco

import (
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	commonRes "github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	falcoModel "github.com/flipped-aurora/gin-vue-admin/server/model/falco"
	falcoReq "github.com/flipped-aurora/gin-vue-admin/server/model/falco/request"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type FalcoApi struct{}

func (f *FalcoApi) getAgentScriptPath(scriptName string) (string, error) {
	root := global.GVA_CONFIG.AutoCode.Root
	if root == "" {
		var err error
		root, err = os.Getwd()
		if err != nil {
			return "", err
		}
		if filepath.Base(root) == "server" {
			root = filepath.Dir(root)
		}
	}
	return filepath.Join(root, "deploy", "falco-agent", scriptName), nil
}

func (f *FalcoApi) createTypedTask(c *gin.Context, taskType string) {
	var task falcoReq.FalcoTaskCreate
	if err := c.ShouldBindJSON(&task); err != nil {
		commonRes.FailWithMessage(err.Error(), c)
		return
	}
	task.TaskType = taskType
	if err := falcoService.CreateTask(task); err != nil {
		global.GVA_LOG.Error("创建 Falco 任务失败", zap.Error(err))
		commonRes.FailWithMessage("创建失败: "+err.Error(), c)
		return
	}
	commonRes.OkWithMessage("创建成功", c)
}

func (f *FalcoApi) CreateHost(c *gin.Context) {
	var host falcoModel.FalcoHost
	if err := c.ShouldBindJSON(&host); err != nil {
		commonRes.FailWithMessage(err.Error(), c)
		return
	}
	if err := falcoService.CreateHost(&host); err != nil {
		global.GVA_LOG.Error("创建 Falco 主机失败", zap.Error(err))
		commonRes.FailWithMessage("创建失败: "+err.Error(), c)
		return
	}
	commonRes.OkWithMessage("创建成功", c)
}

func (f *FalcoApi) UpdateHost(c *gin.Context) {
	var host falcoModel.FalcoHost
	if err := c.ShouldBindJSON(&host); err != nil {
		commonRes.FailWithMessage(err.Error(), c)
		return
	}
	if err := falcoService.UpdateHost(host); err != nil {
		global.GVA_LOG.Error("更新 Falco 主机失败", zap.Error(err))
		commonRes.FailWithMessage("更新失败: "+err.Error(), c)
		return
	}
	commonRes.OkWithMessage("更新成功", c)
}

func (f *FalcoApi) DeleteHost(c *gin.Context) {
	id, _ := strconv.Atoi(c.Query("ID"))
	if err := falcoService.DeleteHost(uint(id)); err != nil {
		global.GVA_LOG.Error("删除 Falco 主机失败", zap.Error(err))
		commonRes.FailWithMessage("删除失败: "+err.Error(), c)
		return
	}
	commonRes.OkWithMessage("删除成功", c)
}

func (f *FalcoApi) GetHostList(c *gin.Context) {
	var search falcoReq.FalcoHostSearch
	if err := c.ShouldBindJSON(&search); err != nil {
		commonRes.FailWithMessage(err.Error(), c)
		return
	}
	list, total, err := falcoService.GetHostList(search)
	if err != nil {
		global.GVA_LOG.Error("获取 Falco 主机列表失败", zap.Error(err))
		commonRes.FailWithMessage("获取失败: "+err.Error(), c)
		return
	}
	commonRes.OkWithDetailed(commonRes.PageResult{
		List:     list,
		Total:    total,
		Page:     search.Page,
		PageSize: search.PageSize,
	}, "获取成功", c)
}

func (f *FalcoApi) GetAgentList(c *gin.Context) {
	var search falcoReq.FalcoAgentSearch
	if err := c.ShouldBindJSON(&search); err != nil {
		commonRes.FailWithMessage(err.Error(), c)
		return
	}
	list, total, err := falcoService.GetAgentList(search)
	if err != nil {
		global.GVA_LOG.Error("获取 Falco Agent 列表失败", zap.Error(err))
		commonRes.FailWithMessage("获取失败: "+err.Error(), c)
		return
	}
	commonRes.OkWithDetailed(commonRes.PageResult{
		List:     list,
		Total:    total,
		Page:     search.Page,
		PageSize: search.PageSize,
	}, "获取成功", c)
}

func (f *FalcoApi) CreateTask(c *gin.Context) {
	var task falcoReq.FalcoTaskCreate
	if err := c.ShouldBindJSON(&task); err != nil {
		commonRes.FailWithMessage(err.Error(), c)
		return
	}
	if err := falcoService.CreateTask(task); err != nil {
		global.GVA_LOG.Error("创建 Falco 任务失败", zap.Error(err))
		commonRes.FailWithMessage("创建失败: "+err.Error(), c)
		return
	}
	commonRes.OkWithMessage("创建成功", c)
}

func (f *FalcoApi) CreateInstallTask(c *gin.Context) {
	f.createTypedTask(c, "install")
}

func (f *FalcoApi) CreateUpgradeTask(c *gin.Context) {
	f.createTypedTask(c, "upgrade")
}

func (f *FalcoApi) CreateRollbackTask(c *gin.Context) {
	f.createTypedTask(c, "rollback")
}

func (f *FalcoApi) CreateReloadTask(c *gin.Context) {
	f.createTypedTask(c, "reload")
}

func (f *FalcoApi) CreateRestartTask(c *gin.Context) {
	f.createTypedTask(c, "restart")
}

func (f *FalcoApi) GetTaskList(c *gin.Context) {
	var search falcoReq.FalcoTaskSearch
	if err := c.ShouldBindJSON(&search); err != nil {
		commonRes.FailWithMessage(err.Error(), c)
		return
	}
	list, total, err := falcoService.GetTaskList(search)
	if err != nil {
		global.GVA_LOG.Error("获取 Falco 任务列表失败", zap.Error(err))
		commonRes.FailWithMessage("获取失败: "+err.Error(), c)
		return
	}
	commonRes.OkWithDetailed(commonRes.PageResult{
		List:     list,
		Total:    total,
		Page:     search.Page,
		PageSize: search.PageSize,
	}, "获取成功", c)
}

func (f *FalcoApi) CreateRulePackage(c *gin.Context) {
	var payload falcoReq.FalcoRulePackageCreate
	if err := c.ShouldBindJSON(&payload); err != nil {
		commonRes.FailWithMessage(err.Error(), c)
		return
	}
	if err := falcoService.CreateRulePackage(payload); err != nil {
		global.GVA_LOG.Error("创建 Falco 规则包失败", zap.Error(err))
		commonRes.FailWithMessage("创建失败: "+err.Error(), c)
		return
	}
	commonRes.OkWithMessage("创建成功", c)
}

func (f *FalcoApi) GetRulePackageList(c *gin.Context) {
	var search falcoReq.FalcoRulePackageSearch
	if err := c.ShouldBindJSON(&search); err != nil {
		commonRes.FailWithMessage(err.Error(), c)
		return
	}
	list, total, err := falcoService.GetRulePackageList(search)
	if err != nil {
		global.GVA_LOG.Error("获取 Falco 规则包列表失败", zap.Error(err))
		commonRes.FailWithMessage("获取失败: "+err.Error(), c)
		return
	}
	commonRes.OkWithDetailed(commonRes.PageResult{
		List:     list,
		Total:    total,
		Page:     search.Page,
		PageSize: search.PageSize,
	}, "获取成功", c)
}

func (f *FalcoApi) PublishRulePackage(c *gin.Context) {
	var payload falcoReq.FalcoRulePublishCreate
	if err := c.ShouldBindJSON(&payload); err != nil {
		commonRes.FailWithMessage(err.Error(), c)
		return
	}
	if err := falcoService.PublishRulePackage(payload); err != nil {
		global.GVA_LOG.Error("发布 Falco 规则包失败", zap.Error(err))
		commonRes.FailWithMessage("发布失败: "+err.Error(), c)
		return
	}
	commonRes.OkWithMessage("发布成功", c)
}

func (f *FalcoApi) GetPublishList(c *gin.Context) {
	var search falcoReq.FalcoPublishSearch
	if err := c.ShouldBindJSON(&search); err != nil {
		commonRes.FailWithMessage(err.Error(), c)
		return
	}
	list, total, err := falcoService.GetPublishList(search)
	if err != nil {
		global.GVA_LOG.Error("获取 Falco 发布记录失败", zap.Error(err))
		commonRes.FailWithMessage("获取失败: "+err.Error(), c)
		return
	}
	commonRes.OkWithDetailed(commonRes.PageResult{
		List:     list,
		Total:    total,
		Page:     search.Page,
		PageSize: search.PageSize,
	}, "获取成功", c)
}

func (f *FalcoApi) GetEventList(c *gin.Context) {
	var search falcoReq.FalcoEventSearch
	if err := c.ShouldBindJSON(&search); err != nil {
		commonRes.FailWithMessage(err.Error(), c)
		return
	}
	list, total, err := falcoService.GetEventList(search)
	if err != nil {
		global.GVA_LOG.Error("获取 Falco 事件列表失败", zap.Error(err))
		commonRes.FailWithMessage("获取失败: "+err.Error(), c)
		return
	}
	commonRes.OkWithDetailed(commonRes.PageResult{
		List:     list,
		Total:    total,
		Page:     search.Page,
		PageSize: search.PageSize,
	}, "获取成功", c)
}

func (f *FalcoApi) GetDashboard(c *gin.Context) {
	dashboard, err := falcoService.GetDashboard()
	if err != nil {
		global.GVA_LOG.Error("获取 Falco 仪表盘失败", zap.Error(err))
		commonRes.FailWithMessage("获取失败: "+err.Error(), c)
		return
	}
	commonRes.OkWithDetailed(dashboard, "获取成功", c)
}

func (f *FalcoApi) GetSettings(c *gin.Context) {
	commonRes.OkWithDetailed(falcoService.GetSettings(), "获取成功", c)
}

func (f *FalcoApi) UpdateSettings(c *gin.Context) {
	var payload falcoReq.FalcoSettingsUpdate
	if err := c.ShouldBindJSON(&payload); err != nil {
		commonRes.FailWithMessage(err.Error(), c)
		return
	}
	if err := falcoService.UpdateSettings(payload); err != nil {
		global.GVA_LOG.Error("更新 Falco 设置失败", zap.Error(err))
		commonRes.FailWithMessage("更新失败: "+err.Error(), c)
		return
	}
	commonRes.OkWithMessage("更新成功", c)
}

func (f *FalcoApi) RegisterAgent(c *gin.Context) {
	var payload falcoReq.FalcoAgentRegister
	if err := c.ShouldBindJSON(&payload); err != nil {
		commonRes.FailWithMessage(err.Error(), c)
		return
	}
	result, err := falcoService.RegisterAgent(payload)
	if err != nil {
		global.GVA_LOG.Error("注册 Falco Agent 失败", zap.Error(err))
		commonRes.FailWithMessage("注册失败: "+err.Error(), c)
		return
	}
	commonRes.OkWithDetailed(result, "注册成功", c)
}

func (f *FalcoApi) AgentHeartbeat(c *gin.Context) {
	var payload falcoReq.FalcoAgentHeartbeat
	if err := c.ShouldBindJSON(&payload); err != nil {
		commonRes.FailWithMessage(err.Error(), c)
		return
	}
	if err := falcoService.AgentHeartbeat(payload); err != nil {
		global.GVA_LOG.Error("Falco Agent 心跳失败", zap.Error(err))
		commonRes.FailWithMessage("心跳失败: "+err.Error(), c)
		return
	}
	commonRes.OkWithMessage("心跳成功", c)
}

func (f *FalcoApi) AgentStatusReport(c *gin.Context) {
	var payload falcoReq.FalcoAgentStatusReport
	if err := c.ShouldBindJSON(&payload); err != nil {
		commonRes.FailWithMessage(err.Error(), c)
		return
	}
	if err := falcoService.AgentStatusReport(payload); err != nil {
		global.GVA_LOG.Error("Falco Agent 状态上报失败", zap.Error(err))
		commonRes.FailWithMessage("状态上报失败: "+err.Error(), c)
		return
	}
	commonRes.OkWithMessage("状态上报成功", c)
}

func (f *FalcoApi) BulkEvents(c *gin.Context) {
	var payload falcoReq.FalcoEventBulk
	if err := c.ShouldBindJSON(&payload); err != nil {
		commonRes.FailWithMessage(err.Error(), c)
		return
	}
	if err := falcoService.BulkCreateEvents(payload); err != nil {
		global.GVA_LOG.Error("Falco 事件批量上报失败", zap.Error(err))
		commonRes.FailWithMessage("事件上报失败: "+err.Error(), c)
		return
	}
	commonRes.OkWithMessage("事件上报成功", c)
}

func (f *FalcoApi) PullAgentTasks(c *gin.Context) {
	var query falcoReq.FalcoAgentAuth
	if err := c.ShouldBindQuery(&query); err != nil {
		commonRes.FailWithMessage(err.Error(), c)
		return
	}
	list, err := falcoService.PullAgentTasks(query)
	if err != nil {
		global.GVA_LOG.Error("拉取 Falco 任务失败", zap.Error(err))
		commonRes.FailWithMessage("拉取失败: "+err.Error(), c)
		return
	}
	commonRes.OkWithDetailed(gin.H{"list": list}, "获取成功", c)
}

func (f *FalcoApi) SaveTaskResult(c *gin.Context) {
	var payload falcoReq.FalcoTaskResult
	if err := c.ShouldBindJSON(&payload); err != nil {
		commonRes.FailWithMessage(err.Error(), c)
		return
	}
	if err := falcoService.SaveTaskResult(payload); err != nil {
		global.GVA_LOG.Error("保存 Falco 任务结果失败", zap.Error(err))
		commonRes.FailWithMessage("保存失败: "+err.Error(), c)
		return
	}
	commonRes.OkWithMessage("保存成功", c)
}

func (f *FalcoApi) AgentWS(c *gin.Context) {
	commonRes.OkWithDetailed(gin.H{
		"enabled": false,
		"message": "一期暂未启用 WebSocket 控制通道",
	}, "获取成功", c)
}

func (f *FalcoApi) DownloadInstallerScript(c *gin.Context) {
	scriptPath, err := f.getAgentScriptPath("install-falco-agent.sh")
	if err != nil {
		global.GVA_LOG.Error("获取安装脚本路径失败", zap.Error(err))
		commonRes.FailWithMessage("获取安装脚本失败", c)
		return
	}
	content, err := os.ReadFile(scriptPath)
	if err != nil {
		global.GVA_LOG.Error("读取安装脚本失败", zap.Error(err))
		commonRes.FailWithMessage("读取安装脚本失败", c)
		return
	}
	c.Data(http.StatusOK, "text/plain; charset=utf-8", content)
}

func (f *FalcoApi) DownloadRuntimeScript(c *gin.Context) {
	scriptPath, err := f.getAgentScriptPath("falco-agent.sh")
	if err != nil {
		global.GVA_LOG.Error("获取 Agent 运行脚本路径失败", zap.Error(err))
		commonRes.FailWithMessage("获取 Agent 运行脚本失败", c)
		return
	}
	content, err := os.ReadFile(scriptPath)
	if err != nil {
		global.GVA_LOG.Error("读取 Agent 运行脚本失败", zap.Error(err))
		commonRes.FailWithMessage("读取 Agent 运行脚本失败", c)
		return
	}
	c.Data(http.StatusOK, "text/plain; charset=utf-8", content)
}
