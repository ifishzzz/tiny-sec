import service from '@/utils/request'

// @Tags SpaceSearch
// @Summary 网络空间搜索
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param query query object true "网络空间搜索参数"
// @Success 200 {object} object "{"success":true,"data":{},"msg":"搜索成功"}"
// @Router /fofa/search [get]
export const spaceSearch = (params) => {
  return service({
    url: '/fofa/search',
    method: 'get',
    params
  })
}
