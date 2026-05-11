package system

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/system"
	systemReq "github.com/flipped-aurora/gin-vue-admin/server/model/system/request"
	systemRes "github.com/flipped-aurora/gin-vue-admin/server/model/system/response"
	gvaRequest "github.com/flipped-aurora/gin-vue-admin/server/utils/request"
)

const (
	fofaAPIEndpoint = "https://fofa.info/api/v1/search/all"
	fofaEmailKey    = "fofa_email"
	fofaAPIKeyKey   = "fofa_key"
)

var fofaFields = []string{"host", "ip", "port", "protocol", "title", "domain", "server", "country", "city"}

type FofaService struct{}

type fofaAPIResponse struct {
	Error   bool    `json:"error"`
	Errmsg  string  `json:"errmsg"`
	Message string  `json:"message"`
	Mode    string  `json:"mode"`
	Query   string  `json:"query"`
	Page    int     `json:"page"`
	Size    int64   `json:"size"`
	Results [][]any `json:"results"`
}

func (fofaService *FofaService) Search(info systemReq.FofaSearch) (result systemRes.FofaSearchResult, err error) {
	query := strings.TrimSpace(info.Query)
	if query == "" {
		return result, fmt.Errorf("请输入 FOFA 查询语法")
	}

	page := info.Page
	if page <= 0 {
		page = 1
	}

	pageSize := info.PageSize
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 10000 {
		pageSize = 10000
	}

	email, key, err := fofaService.getCredentials()
	if err != nil {
		return result, err
	}

	params := map[string]string{
		"email":   email,
		"key":     key,
		"qbase64": base64.StdEncoding.EncodeToString([]byte(query)),
		"fields":  strings.Join(fofaFields, ","),
		"page":    strconv.Itoa(page),
		"size":    strconv.Itoa(pageSize),
		"full":    strconv.FormatBool(info.Full),
	}

	resp, err := gvaRequest.HttpRequestWithTimeout(
		fofaAPIEndpoint,
		http.MethodGet,
		map[string]string{"Accept": "application/json"},
		params,
		nil,
		30*time.Second,
	)
	if err != nil {
		return result, fmt.Errorf("FOFA 请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return result, fmt.Errorf("读取 FOFA 响应失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return result, fmt.Errorf("FOFA 请求失败，状态码 %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var apiResp fofaAPIResponse
	if err = json.Unmarshal(body, &apiResp); err != nil {
		return result, fmt.Errorf("解析 FOFA 响应失败: %w", err)
	}

	if apiResp.Error {
		errMsg := strings.TrimSpace(apiResp.Errmsg)
		if errMsg == "" {
			errMsg = strings.TrimSpace(apiResp.Message)
		}
		if errMsg == "" {
			errMsg = "FOFA 返回错误"
		}
		return result, fmt.Errorf("%s", errMsg)
	}

	items := make([]systemRes.FofaResultItem, 0, len(apiResp.Results))
	for _, row := range apiResp.Results {
		items = append(items, mapFofaResultItem(row))
	}

	result = systemRes.FofaSearchResult{
		List:     items,
		Total:    apiResp.Size,
		Page:     page,
		PageSize: pageSize,
		Query:    apiResp.Query,
		Mode:     apiResp.Mode,
	}
	return result, nil
}

func (fofaService *FofaService) getCredentials() (email string, key string, err error) {
	var emailParam system.SysParams
	if err = global.GVA_DB.Where("`key` = ?", fofaEmailKey).First(&emailParam).Error; err != nil {
		return "", "", fmt.Errorf("请先在参数管理中配置 %s", fofaEmailKey)
	}

	var keyParam system.SysParams
	if err = global.GVA_DB.Where("`key` = ?", fofaAPIKeyKey).First(&keyParam).Error; err != nil {
		return "", "", fmt.Errorf("请先在参数管理中配置 %s", fofaAPIKeyKey)
	}

	email = strings.TrimSpace(emailParam.Value)
	key = strings.TrimSpace(keyParam.Value)
	if email == "" || key == "" {
		return "", "", fmt.Errorf("请先在参数管理中填写 %s 和 %s 的值", fofaEmailKey, fofaAPIKeyKey)
	}

	return email, key, nil
}

func mapFofaResultItem(row []any) systemRes.FofaResultItem {
	item := systemRes.FofaResultItem{}
	for index, field := range fofaFields {
		if index >= len(row) {
			break
		}

		value := valueToString(row[index])
		switch field {
		case "host":
			item.Host = value
		case "ip":
			item.IP = value
		case "port":
			item.Port = valueToInt(row[index])
		case "protocol":
			item.Protocol = value
		case "title":
			item.Title = value
		case "domain":
			item.Domain = value
		case "server":
			item.Server = value
		case "country":
			item.Country = value
		case "city":
			item.City = value
		}
	}
	return item
}

func valueToString(value any) string {
	switch v := value.(type) {
	case nil:
		return ""
	case string:
		return v
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case float32:
		return strconv.FormatFloat(float64(v), 'f', -1, 64)
	case int:
		return strconv.Itoa(v)
	case int64:
		return strconv.FormatInt(v, 10)
	case json.Number:
		return v.String()
	case []string:
		return strings.Join(v, ", ")
	case []any:
		parts := make([]string, 0, len(v))
		for _, item := range v {
			parts = append(parts, valueToString(item))
		}
		return strings.Join(parts, ", ")
	default:
		return fmt.Sprint(v)
	}
}

func valueToInt(value any) int {
	switch v := value.(type) {
	case int:
		return v
	case int64:
		return int(v)
	case float64:
		return int(v)
	case float32:
		return int(v)
	case json.Number:
		i, err := v.Int64()
		if err == nil {
			return int(i)
		}
	case string:
		i, err := strconv.Atoi(v)
		if err == nil {
			return i
		}
	}
	return 0
}
