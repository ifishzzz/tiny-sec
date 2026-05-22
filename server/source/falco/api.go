package falco

import (
	"context"

	sysModel "github.com/flipped-aurora/gin-vue-admin/server/model/system"
	"github.com/flipped-aurora/gin-vue-admin/server/service/system"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

const initOrderFalcoAPI = system.InitOrderExternal + 10

type initAPI struct{}

func init() {
	system.RegisterInit(initOrderFalcoAPI, &initAPI{})
}

func (i *initAPI) InitializerName() string {
	return "falco_api_seed"
}

func (i *initAPI) MigrateTable(ctx context.Context) (context.Context, error) {
	db, ok := ctx.Value("db").(*gorm.DB)
	if !ok {
		return ctx, system.ErrMissingDBContext
	}
	return ctx, db.AutoMigrate(&sysModel.SysApi{})
}

func (i *initAPI) TableCreated(ctx context.Context) bool {
	db, ok := ctx.Value("db").(*gorm.DB)
	if !ok {
		return false
	}
	return db.Migrator().HasTable(&sysModel.SysApi{})
}

func (i *initAPI) InitializeData(ctx context.Context) (context.Context, error) {
	db, ok := ctx.Value("db").(*gorm.DB)
	if !ok {
		return ctx, system.ErrMissingDBContext
	}
	entities := []sysModel.SysApi{
		{ApiGroup: "Falco", Method: "POST", Path: "/falco/dashboard", Description: "Falco 仪表盘"},
		{ApiGroup: "Falco", Method: "POST", Path: "/falco/host/create", Description: "创建 Falco 主机"},
		{ApiGroup: "Falco", Method: "PUT", Path: "/falco/host/update", Description: "更新 Falco 主机"},
		{ApiGroup: "Falco", Method: "DELETE", Path: "/falco/host/delete", Description: "删除 Falco 主机"},
		{ApiGroup: "Falco", Method: "POST", Path: "/falco/host/list", Description: "获取 Falco 主机列表"},
		{ApiGroup: "Falco", Method: "POST", Path: "/falco/agent/list", Description: "获取 Falco Agent 列表"},
		{ApiGroup: "Falco", Method: "POST", Path: "/falco/task/create", Description: "创建 Falco 任务"},
		{ApiGroup: "Falco", Method: "POST", Path: "/falco/task/install", Description: "创建 Falco 安装任务"},
		{ApiGroup: "Falco", Method: "POST", Path: "/falco/task/upgrade", Description: "创建 Falco 升级任务"},
		{ApiGroup: "Falco", Method: "POST", Path: "/falco/task/rollback", Description: "创建 Falco 回滚任务"},
		{ApiGroup: "Falco", Method: "POST", Path: "/falco/task/reload", Description: "创建 Falco 重载任务"},
		{ApiGroup: "Falco", Method: "POST", Path: "/falco/task/restart", Description: "创建 Falco 重启任务"},
		{ApiGroup: "Falco", Method: "POST", Path: "/falco/task/list", Description: "获取 Falco 任务列表"},
		{ApiGroup: "Falco", Method: "POST", Path: "/falco/rule/package/create", Description: "创建 Falco 规则包"},
		{ApiGroup: "Falco", Method: "POST", Path: "/falco/rule/package/list", Description: "获取 Falco 规则包列表"},
		{ApiGroup: "Falco", Method: "POST", Path: "/falco/rule/publish", Description: "发布 Falco 规则包"},
		{ApiGroup: "Falco", Method: "POST", Path: "/falco/rule/publish/list", Description: "获取 Falco 规则发布记录"},
		{ApiGroup: "Falco", Method: "POST", Path: "/falco/event/list", Description: "获取 Falco 事件列表"},
		{ApiGroup: "Falco", Method: "GET", Path: "/falco/settings", Description: "获取 Falco 设置"},
		{ApiGroup: "Falco", Method: "PUT", Path: "/falco/settings", Description: "更新 Falco 设置"},
	}
	if err := db.Create(&entities).Error; err != nil {
		return ctx, errors.Wrap(err, "Falco API 初始化失败")
	}
	return context.WithValue(ctx, i.InitializerName(), entities), nil
}

func (i *initAPI) DataInserted(ctx context.Context) bool {
	db, ok := ctx.Value("db").(*gorm.DB)
	if !ok {
		return false
	}
	if errors.Is(db.Where("path = ? AND method = ?", "/falco/dashboard", "POST").First(&sysModel.SysApi{}).Error, gorm.ErrRecordNotFound) {
		return false
	}
	return true
}
