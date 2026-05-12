package system

import (
	"strings"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	commonRes "github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	systemReq "github.com/flipped-aurora/gin-vue-admin/server/model/system/request"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type FofaApi struct{}

// Search 网络空间搜索
// @Tags SpaceSearch
// @Summary 网络空间搜索
// @Security ApiKeyAuth
// @Produce application/json
// @Param query query systemReq.FofaSearch true "网络空间搜索参数"
// @Success 200 {object} commonRes.Response{data=systemRes.FofaSearchResult,msg=string} "搜索成功"
// @Router /fofa/search [get]
func (fofaApi *FofaApi) Search(c *gin.Context) {
	var searchInfo systemReq.FofaSearch
	if err := c.ShouldBindQuery(&searchInfo); err != nil {
		commonRes.FailWithMessage(err.Error(), c)
		return
	}

	if strings.TrimSpace(searchInfo.Query) == "" {
		commonRes.FailWithMessage("请输入查询语法", c)
		return
	}

	result, err := fofaService.Search(searchInfo)
	if err != nil {
		global.GVA_LOG.Error("网络空间搜索失败", zap.Error(err))
		commonRes.FailWithMessage("网络空间搜索失败: "+err.Error(), c)
		return
	}

	commonRes.OkWithDetailed(result, "搜索成功", c)
}
