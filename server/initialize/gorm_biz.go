package initialize

import (
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	falcoModel "github.com/flipped-aurora/gin-vue-admin/server/model/falco"
)

func bizModel() error {
	return global.GVA_DB.AutoMigrate(
		&falcoModel.FalcoHost{},
		&falcoModel.FalcoAgent{},
		&falcoModel.FalcoInstallTask{},
		&falcoModel.FalcoRulePackage{},
		&falcoModel.FalcoRulePublish{},
		&falcoModel.FalcoRuntimeStatus{},
		&falcoModel.FalcoEvent{},
	)
}
