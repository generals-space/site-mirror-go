package crawler

// Config ...
type Config struct {
	// 单个页面中可能包含链接的最大数量
	// 用LinkRatioInSinglePage*PageWorkerCount得到PageQueueSize,
	// 这是为了防止由于队列满而造成worker阻塞引起的列锁,
	// 但仍然可能由于递归抓取而写满队列
	LinkRatioInSinglePage int
	PageWorkerCount       int
	AssetWorkerCount      int

	SiteDBPath string
	SitePath   string

	StartPage string
	MainSite  string
	UserAgent string
	// 爬取页面的深度, 从1开始计, 爬到第N层为止.
	// 1表示只抓取单页, 0表示无限制
	MaxDepth int
	// 请求出错最大重试次数(超时也算出错)
	MaxRetryTimes int

	OutsiteAsset bool
	NoJs         bool
	NoCSS        bool
	NoImages     bool
	NoFonts      bool
	BlackList    []string
}

// NewConfig 获取默认配置
func NewConfig() (config *Config) {
	config = &Config{
		LinkRatioInSinglePage: 5000,
		PageWorkerCount:       10,
		AssetWorkerCount:      10,

		SiteDBPath: "site.db",
		SitePath:   "sites",

		OutsiteAsset: true,
		NoJs:         true,
		NoCSS:        false,
		NoImages:     false,
		NoFonts:      false,
		BlackList:    []string{},
	}

	return
}
