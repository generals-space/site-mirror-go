package crawler

// Config ...
type Config struct {
	PageQueueSize  int
	AssetQueueSize int

	SiteDBPath string
	SitePath   string

	StartPage string
	UserAgent string
	// 爬取页面的深度, 从1开始计, 爬到第N层为止.
	// 1表示只抓取单页, 0表示无限制
	MaxDepth int
	// 请求出错最大重试次数(超时也算出错)
	MaxRetryTimes int

	PageWorkerCount  int
	AssetWorkerCount int

	NoJs     string
	NoCSS    string
	NoImages string
	NoFonts  string
}
