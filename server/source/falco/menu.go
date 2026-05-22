package falco

import (
	"context"

	sysModel "github.com/flipped-aurora/gin-vue-admin/server/model/system"
	"github.com/flipped-aurora/gin-vue-admin/server/service/system"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

const initOrderFalcoMenu = initOrderFalcoAPI + 1

type initMenu struct{}

func init() {
	system.RegisterInit(initOrderFalcoMenu, &initMenu{})
}

func (i *initMenu) InitializerName() string {
	return "falco_menu_seed"
}

func (i *initMenu) MigrateTable(ctx context.Context) (context.Context, error) {
	db, ok := ctx.Value("db").(*gorm.DB)
	if !ok {
		return ctx, system.ErrMissingDBContext
	}
	return ctx, db.AutoMigrate(&sysModel.SysBaseMenu{}, &sysModel.SysBaseMenuParameter{}, &sysModel.SysBaseMenuBtn{})
}

func (i *initMenu) TableCreated(ctx context.Context) bool {
	db, ok := ctx.Value("db").(*gorm.DB)
	if !ok {
		return false
	}
	return db.Migrator().HasTable(&sysModel.SysBaseMenu{})
}

func (i *initMenu) InitializeData(ctx context.Context) (context.Context, error) {
	db, ok := ctx.Value("db").(*gorm.DB)
	if !ok {
		return ctx, system.ErrMissingDBContext
	}

	parent := sysModel.SysBaseMenu{
		MenuLevel: 0,
		Hidden:    false,
		ParentId:  0,
		Path:      "falco",
		Name:      "falco",
		Component: "view/routerHolder.vue",
		Sort:      11,
		Meta:      sysModel.Meta{Title: "Falco 管理", Icon: "monitor"},
	}
	if err := db.Create(&parent).Error; err != nil {
		return ctx, errors.Wrap(err, "Falco 父菜单初始化失败")
	}

	children := []sysModel.SysBaseMenu{
		{MenuLevel: 1, Hidden: false, ParentId: parent.ID, Path: "dashboard", Name: "falcoDashboard", Component: "view/falco/dashboard/index.vue", Sort: 1, Meta: sysModel.Meta{Title: "仪表盘", Icon: "data-analysis", KeepAlive: true}},
		{MenuLevel: 1, Hidden: false, ParentId: parent.ID, Path: "hosts", Name: "falcoHosts", Component: "view/falco/hosts/index.vue", Sort: 2, Meta: sysModel.Meta{Title: "主机管理", Icon: "monitor", KeepAlive: true}},
		{MenuLevel: 1, Hidden: false, ParentId: parent.ID, Path: "agents", Name: "falcoAgents", Component: "view/falco/agents/index.vue", Sort: 3, Meta: sysModel.Meta{Title: "Agent 管理", Icon: "cpu", KeepAlive: true}},
		{MenuLevel: 1, Hidden: false, ParentId: parent.ID, Path: "rules", Name: "falcoRules", Component: "view/falco/rules/index.vue", Sort: 4, Meta: sysModel.Meta{Title: "规则中心", Icon: "document", KeepAlive: true}},
		{MenuLevel: 1, Hidden: false, ParentId: parent.ID, Path: "tasks", Name: "falcoTasks", Component: "view/falco/tasks/index.vue", Sort: 5, Meta: sysModel.Meta{Title: "任务中心", Icon: "list", KeepAlive: true}},
		{MenuLevel: 1, Hidden: false, ParentId: parent.ID, Path: "events", Name: "falcoEvents", Component: "view/falco/events/index.vue", Sort: 6, Meta: sysModel.Meta{Title: "事件中心", Icon: "bell", KeepAlive: true}},
		{MenuLevel: 1, Hidden: false, ParentId: parent.ID, Path: "settings", Name: "falcoSettings", Component: "view/falco/settings/index.vue", Sort: 7, Meta: sysModel.Meta{Title: "系统设置", Icon: "setting", KeepAlive: true}},
	}
	if err := db.Create(&children).Error; err != nil {
		return ctx, errors.Wrap(err, "Falco 子菜单初始化失败")
	}
	return context.WithValue(ctx, i.InitializerName(), append([]sysModel.SysBaseMenu{parent}, children...)), nil
}

func (i *initMenu) DataInserted(ctx context.Context) bool {
	db, ok := ctx.Value("db").(*gorm.DB)
	if !ok {
		return false
	}
	if errors.Is(db.Where("name = ?", "falco").First(&sysModel.SysBaseMenu{}).Error, gorm.ErrRecordNotFound) {
		return false
	}
	return true
}
