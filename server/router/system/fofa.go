package system

import "github.com/gin-gonic/gin"

type FofaRouter struct{}

// InitFofaRouter 初始化 FOFA 路由
func (f *FofaRouter) InitFofaRouter(Router *gin.RouterGroup) {
	fofaRouter := Router.Group("fofa")
	{
		fofaRouter.GET("search", fofaApi.Search) // FOFA 搜索
	}
}
