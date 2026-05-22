package falco

import "github.com/flipped-aurora/gin-vue-admin/server/service"

type ApiGroup struct {
	FalcoApi
}

var (
	falcoService = service.ServiceGroupApp.FalcoServiceGroup.FalcoService
)
