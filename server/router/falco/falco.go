package falco

import (
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

type FalcoRouter struct{}

func (r *FalcoRouter) InitFalcoRouter(privateRouter *gin.RouterGroup, publicRouter *gin.RouterGroup) {
	falcoRouter := privateRouter.Group("falco").Use(middleware.OperationRecord())
	falcoRouterWithoutRecord := privateRouter.Group("falco")
	falcoAgentRouter := publicRouter.Group("falco/agent")
	falcoAgentInstallRouter := publicRouter.Group("falco/agent/install")

	{
		falcoRouter.POST("host/create", falcoApi.CreateHost)
		falcoRouter.PUT("host/update", falcoApi.UpdateHost)
		falcoRouter.DELETE("host/delete", falcoApi.DeleteHost)
		falcoRouter.POST("task/create", falcoApi.CreateTask)
		falcoRouter.POST("task/install", falcoApi.CreateInstallTask)
		falcoRouter.POST("task/upgrade", falcoApi.CreateUpgradeTask)
		falcoRouter.POST("task/rollback", falcoApi.CreateRollbackTask)
		falcoRouter.POST("task/reload", falcoApi.CreateReloadTask)
		falcoRouter.POST("task/restart", falcoApi.CreateRestartTask)
		falcoRouter.POST("rule/package/create", falcoApi.CreateRulePackage)
		falcoRouter.POST("rule/publish", falcoApi.PublishRulePackage)
		falcoRouter.PUT("settings", falcoApi.UpdateSettings)
	}
	{
		falcoRouterWithoutRecord.POST("dashboard", falcoApi.GetDashboard)
		falcoRouterWithoutRecord.POST("host/list", falcoApi.GetHostList)
		falcoRouterWithoutRecord.POST("agent/list", falcoApi.GetAgentList)
		falcoRouterWithoutRecord.POST("task/list", falcoApi.GetTaskList)
		falcoRouterWithoutRecord.POST("rule/package/list", falcoApi.GetRulePackageList)
		falcoRouterWithoutRecord.POST("rule/publish/list", falcoApi.GetPublishList)
		falcoRouterWithoutRecord.POST("event/list", falcoApi.GetEventList)
		falcoRouterWithoutRecord.GET("settings", falcoApi.GetSettings)
	}
	{
		falcoAgentRouter.POST("register", falcoApi.RegisterAgent)
		falcoAgentRouter.POST("heartbeat", falcoApi.AgentHeartbeat)
		falcoAgentRouter.POST("status/report", falcoApi.AgentStatusReport)
		falcoAgentRouter.POST("events/bulk", falcoApi.BulkEvents)
		falcoAgentRouter.GET("task/pull", falcoApi.PullAgentTasks)
		falcoAgentRouter.POST("task/result", falcoApi.SaveTaskResult)
		falcoAgentRouter.GET("ws", falcoApi.AgentWS)
	}
	{
		falcoAgentInstallRouter.GET("installer", falcoApi.DownloadInstallerScript)
		falcoAgentInstallRouter.GET("runtime", falcoApi.DownloadRuntimeScript)
	}
}
