package falco

import (
	"context"

	sysModel "github.com/flipped-aurora/gin-vue-admin/server/model/system"
	"github.com/flipped-aurora/gin-vue-admin/server/service/system"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

const initOrderFalcoMenuAuthority = initOrderFalcoMenu + 1

type initMenuAuthority struct{}

func init() {
	system.RegisterInit(initOrderFalcoMenuAuthority, &initMenuAuthority{})
}

func (i *initMenuAuthority) InitializerName() string {
	return "falco_menu_authorities"
}

func (i *initMenuAuthority) MigrateTable(ctx context.Context) (context.Context, error) {
	return ctx, nil
}

func (i *initMenuAuthority) TableCreated(ctx context.Context) bool {
	return false
}

func (i *initMenuAuthority) InitializeData(ctx context.Context) (context.Context, error) {
	db, ok := ctx.Value("db").(*gorm.DB)
	if !ok {
		return ctx, system.ErrMissingDBContext
	}

	var falcoMenus []sysModel.SysBaseMenu
	if err := db.Where("name IN ?", []string{
		"falco",
		"falcoDashboard",
		"falcoHosts",
		"falcoAgents",
		"falcoRules",
		"falcoTasks",
		"falcoEvents",
		"falcoSettings",
	}).Find(&falcoMenus).Error; err != nil {
		return ctx, errors.Wrap(err, "查询 Falco 菜单失败")
	}

	if len(falcoMenus) == 0 {
		return ctx, errors.New("未找到 Falco 菜单，无法建立默认授权")
	}

	if err := appendMenusByAuthorityID(db, 888, falcoMenus); err != nil {
		return ctx, err
	}
	if err := appendMenusByAuthorityID(db, 9528, falcoMenus); err != nil {
		return ctx, err
	}

	return context.WithValue(ctx, i.InitializerName(), falcoMenus), nil
}

func (i *initMenuAuthority) DataInserted(ctx context.Context) bool {
	db, ok := ctx.Value("db").(*gorm.DB)
	if !ok {
		return false
	}

	var auth sysModel.SysAuthority
	if err := db.Where("authority_id = ?", 9528).Preload("SysBaseMenus").First(&auth).Error; err != nil {
		return false
	}

	for _, menu := range auth.SysBaseMenus {
		if menu.Name == "falco" {
			return true
		}
	}
	return false
}

func appendMenusByAuthorityID(db *gorm.DB, authorityID uint, menus []sysModel.SysBaseMenu) error {
	var auth sysModel.SysAuthority
	if err := db.Where("authority_id = ?", authorityID).First(&auth).Error; err != nil {
		return errors.Wrapf(err, "查询角色 %d 失败", authorityID)
	}
	if err := db.Model(&auth).Association("SysBaseMenus").Append(menus); err != nil {
		return errors.Wrapf(err, "为角色 %d 追加 Falco 菜单失败", authorityID)
	}
	return nil
}
