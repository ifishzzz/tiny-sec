package request

import commonRequest "github.com/flipped-aurora/gin-vue-admin/server/model/common/request"

type FofaSearch struct {
	Query string `json:"query" form:"query"`
	Full  bool   `json:"full" form:"full"`
	commonRequest.PageInfo
}
