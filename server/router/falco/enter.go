package falco

import api "github.com/flipped-aurora/gin-vue-admin/server/api/v1"

type RouterGroup struct {
	FalcoRouter
}

var (
	falcoApi = api.ApiGroupApp.FalcoApiGroup.FalcoApi
)
