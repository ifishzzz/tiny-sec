import service from '@/utils/request'

// @Tags Fofa
// @Summary FOFA 资产搜索
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param query query object true "FOFA 查询参数"
// @Success 200 {object} object "{"success":true,"data":{},"msg":"搜索成功"}"
// @Router /fofa/search [get]
export const fofaSearch = (params) => {
  return service({
    url: '/fofa/search',
    method: 'get',
    params
  })
}
