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
	searchEngineFofa  = "fofa"
	searchEngineQuake = "quake"

	fofaAPIEndpoint  = "https://fofa.info/api/v1/search/all"
	fofaEmailKey     = "fofa_email"
	fofaAPIKeyKey    = "fofa_key"
	quakeAPIEndpoint = "https://quake.360.net/api/v3/search/quake_service"
	quakeAPIKeyKey   = "quake_key"
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

type quakeAPIResponse struct {
	Code    quakeCode       `json:"code"`
	Message string          `json:"message"`
	Meta    quakeMeta       `json:"meta"`
	Data    json.RawMessage `json:"data"`
}

type quakeCode string

func (c *quakeCode) UnmarshalJSON(data []byte) error {
	text := strings.TrimSpace(string(data))
	if text == "" || text == "null" {
		*c = ""
		return nil
	}

	if len(text) >= 2 && text[0] == '"' && text[len(text)-1] == '"' {
		text = text[1 : len(text)-1]
	}

	*c = quakeCode(text)
	return nil
}

type quakeMeta struct {
	Pagination struct {
		Total int64 `json:"total"`
	} `json:"pagination"`
}

type quakeRecord struct {
	IP       string `json:"ip"`
	Port     int    `json:"port"`
	Location struct {
		CountryCN string `json:"country_cn"`
		Country   string `json:"country"`
		CityCN    string `json:"city_cn"`
		City      string `json:"city"`
	} `json:"location"`
	Service struct {
		Name    string `json:"name"`
		Product string `json:"product"`
		HTTP    struct {
			Host  string `json:"host"`
			Title string `json:"title"`
		} `json:"http"`
	} `json:"service"`
}

func (fofaService *FofaService) Search(info systemReq.FofaSearch) (systemRes.FofaSearchResult, error) {
	switch normalizeSearchEngine(info.Engine) {
	case searchEngineQuake:
		return fofaService.searchQuake(info)
	default:
		return fofaService.searchFofa(info)
	}
}

func (fofaService *FofaService) searchFofa(info systemReq.FofaSearch) (result systemRes.FofaSearchResult, err error) {
	query := strings.TrimSpace(info.Query)
	if query == "" {
		return result, fmt.Errorf("请输入查询语法")
	}

	page, pageSize := normalizePagination(info.Page, info.PageSize, 10000)

	email, key, err := fofaService.getFofaCredentials()
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
		Engine:   searchEngineFofa,
		Query:    apiResp.Query,
		Mode:     apiResp.Mode,
	}
	if result.Query == "" {
		result.Query = query
	}
	return result, nil
}

func (fofaService *FofaService) searchQuake(info systemReq.FofaSearch) (result systemRes.FofaSearchResult, err error) {
	query := strings.TrimSpace(info.Query)
	if query == "" {
		return result, fmt.Errorf("请输入查询语法")
	}

	page, pageSize := normalizePagination(info.Page, info.PageSize, 100)
	start := (page - 1) * pageSize

	key, err := fofaService.getQuakeKey()
	if err != nil {
		return result, err
	}

	requestBody := map[string]any{
		"query":  query,
		"start":  start,
		"size":   pageSize,
		"latest": !info.Full,
	}

	resp, err := gvaRequest.HttpRequestWithTimeout(
		quakeAPIEndpoint,
		http.MethodPost,
		map[string]string{
			"Accept":       "application/json",
			"X-QuakeToken": key,
		},
		nil,
		requestBody,
		30*time.Second,
	)
	if err != nil {
		return result, fmt.Errorf("Quake 请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return result, fmt.Errorf("读取 Quake 响应失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return result, fmt.Errorf("Quake 请求失败，状态码 %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var apiResp quakeAPIResponse
	if err = json.Unmarshal(body, &apiResp); err != nil {
		return result, fmt.Errorf("解析 Quake 响应失败: %w", err)
	}

	if !apiResp.Code.IsSuccess() {
		errMsg := strings.TrimSpace(apiResp.Message)
		if errMsg == "" {
			errMsg = "Quake 返回错误(" + apiResp.Code.String() + ")"
		}
		return result, fmt.Errorf("%s", errMsg)
	}

	var records []quakeRecord
	if len(apiResp.Data) > 0 && string(apiResp.Data) != "null" {
		if err = json.Unmarshal(apiResp.Data, &records); err != nil {
			return result, fmt.Errorf("解析 Quake 数据失败: %w", err)
		}
	}

	items := make([]systemRes.FofaResultItem, 0, len(records))
	for _, row := range records {
		items = append(items, mapQuakeResultItem(row))
	}

	mode := "latest"
	if info.Full {
		mode = "full"
	}

	result = systemRes.FofaSearchResult{
		List:     items,
		Total:    apiResp.Meta.Pagination.Total,
		Page:     page,
		PageSize: pageSize,
		Engine:   searchEngineQuake,
		Query:    query,
		Mode:     mode,
	}
	return result, nil
}

func (fofaService *FofaService) getFofaCredentials() (email string, key string, err error) {
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

func (fofaService *FofaService) getQuakeKey() (string, error) {
	var keyParam system.SysParams
	if err := global.GVA_DB.Where("`key` = ?", quakeAPIKeyKey).First(&keyParam).Error; err != nil {
		return "", fmt.Errorf("请先在参数管理中配置 %s", quakeAPIKeyKey)
	}

	key := strings.TrimSpace(keyParam.Value)
	if key == "" {
		return "", fmt.Errorf("请先在参数管理中填写 %s 的值", quakeAPIKeyKey)
	}
	return key, nil
}

func (c quakeCode) IsSuccess() bool {
	code := strings.TrimSpace(string(c))
	return code == "" || code == "0"
}

func (c quakeCode) String() string {
	return strings.TrimSpace(string(c))
}

func normalizeSearchEngine(engine string) string {
	switch strings.ToLower(strings.TrimSpace(engine)) {
	case searchEngineQuake:
		return searchEngineQuake
	default:
		return searchEngineFofa
	}
}

func normalizePagination(page int, pageSize int, maxPageSize int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	if maxPageSize > 0 && pageSize > maxPageSize {
		pageSize = maxPageSize
	}
	return page, pageSize
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

func mapQuakeResultItem(row quakeRecord) systemRes.FofaResultItem {
	host := strings.TrimSpace(row.Service.HTTP.Host)
	protocol := "http"
	switch {
	case strings.HasPrefix(host, "https://"):
		protocol = "https"
	case strings.HasPrefix(host, "http://"):
		protocol = "http"
	case row.Port == 443:
		protocol = "https"
	}

	return systemRes.FofaResultItem{
		Host:     host,
		IP:       strings.TrimSpace(row.IP),
		Port:     row.Port,
		Protocol: protocol,
		Title:    strings.TrimSpace(row.Service.HTTP.Title),
		Server:   firstNonEmptyString(strings.TrimSpace(row.Service.Product), strings.TrimSpace(row.Service.Name)),
		Country:  firstNonEmptyString(strings.TrimSpace(row.Location.CountryCN), strings.TrimSpace(row.Location.Country)),
		City:     firstNonEmptyString(strings.TrimSpace(row.Location.CityCN), strings.TrimSpace(row.Location.City)),
	}
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
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
