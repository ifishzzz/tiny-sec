package falco

import (
	"context"

	adapter "github.com/casbin/gorm-adapter/v3"
	"github.com/flipped-aurora/gin-vue-admin/server/service/system"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

const initOrderFalcoCasbin = initOrderFalcoMenuAuthority + 1

type initCasbin struct{}

func init() {
	system.RegisterInit(initOrderFalcoCasbin, &initCasbin{})
}

func (i *initCasbin) InitializerName() string {
	return "falco_casbin_seed"
}

func (i *initCasbin) MigrateTable(ctx context.Context) (context.Context, error) {
	db, ok := ctx.Value("db").(*gorm.DB)
	if !ok {
		return ctx, system.ErrMissingDBContext
	}
	return ctx, db.AutoMigrate(&adapter.CasbinRule{})
}

func (i *initCasbin) TableCreated(ctx context.Context) bool {
	db, ok := ctx.Value("db").(*gorm.DB)
	if !ok {
		return false
	}
	return db.Migrator().HasTable(&adapter.CasbinRule{})
}

func (i *initCasbin) InitializeData(ctx context.Context) (context.Context, error) {
	db, ok := ctx.Value("db").(*gorm.DB)
	if !ok {
		return ctx, system.ErrMissingDBContext
	}

	entities := []adapter.CasbinRule{
		{Ptype: "p", V0: "888", V1: "/falco/dashboard", V2: "POST"},
		{Ptype: "p", V0: "888", V1: "/falco/host/create", V2: "POST"},
		{Ptype: "p", V0: "888", V1: "/falco/host/update", V2: "PUT"},
		{Ptype: "p", V0: "888", V1: "/falco/host/delete", V2: "DELETE"},
		{Ptype: "p", V0: "888", V1: "/falco/host/list", V2: "POST"},
		{Ptype: "p", V0: "888", V1: "/falco/agent/list", V2: "POST"},
		{Ptype: "p", V0: "888", V1: "/falco/task/create", V2: "POST"},
		{Ptype: "p", V0: "888", V1: "/falco/task/install", V2: "POST"},
		{Ptype: "p", V0: "888", V1: "/falco/task/upgrade", V2: "POST"},
		{Ptype: "p", V0: "888", V1: "/falco/task/rollback", V2: "POST"},
		{Ptype: "p", V0: "888", V1: "/falco/task/reload", V2: "POST"},
		{Ptype: "p", V0: "888", V1: "/falco/task/restart", V2: "POST"},
		{Ptype: "p", V0: "888", V1: "/falco/task/list", V2: "POST"},
		{Ptype: "p", V0: "888", V1: "/falco/rule/package/create", V2: "POST"},
		{Ptype: "p", V0: "888", V1: "/falco/rule/package/list", V2: "POST"},
		{Ptype: "p", V0: "888", V1: "/falco/rule/publish", V2: "POST"},
		{Ptype: "p", V0: "888", V1: "/falco/rule/publish/list", V2: "POST"},
		{Ptype: "p", V0: "888", V1: "/falco/event/list", V2: "POST"},
		{Ptype: "p", V0: "888", V1: "/falco/settings", V2: "GET"},
		{Ptype: "p", V0: "888", V1: "/falco/settings", V2: "PUT"},

		{Ptype: "p", V0: "9528", V1: "/falco/dashboard", V2: "POST"},
		{Ptype: "p", V0: "9528", V1: "/falco/host/list", V2: "POST"},
		{Ptype: "p", V0: "9528", V1: "/falco/agent/list", V2: "POST"},
		{Ptype: "p", V0: "9528", V1: "/falco/task/list", V2: "POST"},
		{Ptype: "p", V0: "9528", V1: "/falco/rule/package/list", V2: "POST"},
		{Ptype: "p", V0: "9528", V1: "/falco/rule/publish/list", V2: "POST"},
		{Ptype: "p", V0: "9528", V1: "/falco/event/list", V2: "POST"},
		{Ptype: "p", V0: "9528", V1: "/falco/settings", V2: "GET"},
	}

	if err := db.Create(&entities).Error; err != nil {
		return ctx, errors.Wrap(err, "Falco Casbin 初始化失败")
	}
	return context.WithValue(ctx, i.InitializerName(), entities), nil
}

func (i *initCasbin) DataInserted(ctx context.Context) bool {
	db, ok := ctx.Value("db").(*gorm.DB)
	if !ok {
		return false
	}
	if errors.Is(db.Where(adapter.CasbinRule{Ptype: "p", V0: "888", V1: "/falco/dashboard", V2: "POST"}).
		First(&adapter.CasbinRule{}).Error, gorm.ErrRecordNotFound) {
		return false
	}
	return true
}
