package response

type FofaResultItem struct {
	Host     string `json:"host"`
	IP       string `json:"ip"`
	Port     int    `json:"port"`
	Protocol string `json:"protocol"`
	Title    string `json:"title"`
	Domain   string `json:"domain"`
	Server   string `json:"server"`
	Country  string `json:"country"`
	City     string `json:"city"`
}

type FofaSearchResult struct {
	List     []FofaResultItem `json:"list"`
	Total    int64            `json:"total"`
	Page     int              `json:"page"`
	PageSize int              `json:"pageSize"`
	Engine   string           `json:"engine"`
	Query    string           `json:"query"`
	Mode     string           `json:"mode"`
}
