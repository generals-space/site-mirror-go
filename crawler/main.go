package crawler

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"

	"gitee.com/generals-space/site-mirror-go.git/util"
)

var logger = util.NewLogger(os.Stdout)

// Crawler ...
type Crawler struct {
	PageQueue  chan *RequestTask // 页面任务队列
	AssetQueue chan *RequestTask // 静态资源任务队列

	PageWorkerCount  int
	AssetWorkerCount int

	Config *Config
	// PageWorkerPool
	// AssetWorkerPool
}

// NewCrawler 创建Crawler对象
func NewCrawler(config *Config) (crawler *Crawler, err error) {
	pageQueue := make(chan *RequestTask, config.PageQueueSize)
	assetQueue := make(chan *RequestTask, config.AssetQueueSize)
	crawler = &Crawler{
		PageQueue:  pageQueue,
		AssetQueue: assetQueue,

		PageWorkerCount:  config.PageQueueSize,
		AssetWorkerCount: config.AssetQueueSize,

		Config: config,
	}
	return
}

// Start 启动n个协程
func (crawler *Crawler) Start() {
	req := &RequestTask{
		URL:         crawler.Config.StartPage,
		URLType:     PageURL,
		Refer:       "",
		Depth:       0,
		FailedTimes: 0,
	}
	crawler.PageQueue <- req

	var x, y int
	for {
		x++
		if x > crawler.PageWorkerCount {
			break
		}
		go crawler.GetHTMLPage(x)
	}
	for {
		y++
		if y > crawler.AssetWorkerCount {
			break
		}
		go crawler.GetStaticAsset(y)
	}
}

// GetHTMLPage 工作协程, 从队列中获取任务, 请求html页面并解析
func (crawler *Crawler) GetHTMLPage(num int) {
	for req := range crawler.PageQueue {
		fmt.Printf("%+v\n", req)
		resp, err := getURL(req.URL, req.Refer)

		if err != nil {
			logger.Errorf("请求页面失败: url: %s, refer: %s, error: %s, 重新入队列", req.URL, req.Refer, err.Error())
			continue
		}
		body, err := ioutil.ReadAll(resp.Body)
		charsetName, err := getPageCharset(body)
		if err != nil {
			logger.Errorf("获取页面编码失败: url: %s, refer: %s, error: %s, 重新入队列", req.URL, req.Refer, err.Error())
		}
		logger.Debugf("页面编码: %s, url: %s, refer: %s", charsetName, req.URL, req.Refer)
		// 解析页面

		charset, exist := CharsetMap[charsetName]
		if !exist {
			logger.Debugf("未找到匹配的编码: %s, url: %s, refer: %s", charsetName, req.URL, req.Refer)
		}
		bodyReader := bytes.NewReader(body)
		utf8Reader := charset.NewDecoder().Reader(bodyReader)
		buf, err := ioutil.ReadAll(utf8Reader)
		if err != nil {
			logger.Errorf("读取页面内容失败: url: %s, refer: %s, error: %s", req.URL, req.Refer, err.Error())
		}
		logger.Debugf("以UTF-8编码方式打开: %s", string(buf))

	}
}

// GetStaticAsset 工作协程, 从队列中获取任务, 获取静态资源并存储
func (crawler *Crawler) GetStaticAsset(num int) {
	for req := range crawler.AssetQueue {
		fmt.Printf("%+v\n", req)
	}
}
