package crawler

import (
	"bytes"
	"io/ioutil"
	"net/url"
	"os"
	"strings"

	"gitee.com/generals-space/site-mirror-go.git/util"
	"github.com/PuerkitoBio/goquery"
)

var logger = util.NewLogger(os.Stdout)

// Crawler ...
type Crawler struct {
	PageQueue  chan *RequestTask // 页面任务队列
	AssetQueue chan *RequestTask // 静态资源任务队列

	Config   *Config
	MainSite string
}

// NewCrawler 创建Crawler对象
func NewCrawler(config *Config) (crawler *Crawler, err error) {
	pageQueue := make(chan *RequestTask, config.PageQueueSize)
	assetQueue := make(chan *RequestTask, config.AssetQueueSize)
	urlObj, err := url.Parse(config.StartPage)
	if err != nil {
		logger.Errorf("解析起始地址失败: url: %s, %s", config.StartPage, err.Error())
		return
	}
	mainSite := urlObj.Host // Host成员带端口.
	crawler = &Crawler{
		PageQueue:  pageQueue,
		AssetQueue: assetQueue,

		Config:   config,
		MainSite: mainSite,
	}
	return
}

// Start 启动n个协程
func (crawler *Crawler) Start() {
	req := &RequestTask{
		URL:         crawler.Config.StartPage,
		URLType:     PageURL,
		Refer:       "",
		Depth:       1,
		FailedTimes: 0,
	}
	crawler.PageQueue <- req

	var x int
	for {
		x++
		if x > crawler.Config.PageWorkerCount {
			break
		}
		go crawler.GetHTMLPage(x)
	}

	var y int
	for {
		y++
		if y > crawler.Config.PageWorkerCount {
			break
		}
		go crawler.GetStaticAsset(y)
	}
}

// GetHTMLPage 工作协程, 从队列中获取任务, 请求html页面并解析
func (crawler *Crawler) GetHTMLPage(num int) {
	for req := range crawler.PageQueue {
		logger.Infof("取得页面任务: %+v", req)
		if 0 < crawler.Config.MaxDepth && crawler.Config.MaxDepth < req.Depth {
			logger.Infof("当前页面已达到最大深度, 不再抓取新页面: url: %s, refer: %s, depth: %d", req.URL, req.Refer, req.Depth)
		}
		resp, err := getURL(req.URL, req.Refer)
		if err != nil {
			logger.Errorf("请求页面失败: url: %s, refer: %s, error: %s, 重新入队列", req.URL, req.Refer, err.Error())
			continue
		}
		defer resp.Body.Close()
		bodyContent, err := ioutil.ReadAll(resp.Body)
		charsetName, err := getPageCharset(bodyContent)
		if err != nil {
			logger.Errorf("获取页面编码失败: url: %s, refer: %s, error: %s, 重新入队列", req.URL, req.Refer, err.Error())
		}
		logger.Debugf("页面编码: %s, url: %s, refer: %s", charsetName, req.URL, req.Refer)
		// 解析页面
		charset, exist := CharsetMap[strings.ToLower(charsetName)]
		if !exist {
			logger.Debugf("未找到匹配的编码: %s, url: %s, refer: %s", charsetName, req.URL, req.Refer)
		}
		utf8Coutent, err := DecodeToUTF8(bodyContent, charset)
		if err != nil {
			logger.Errorf("页面解码失败: %s", err.Error())
		}
		utf8Reader := bytes.NewReader(utf8Coutent)
		htmlDom, err := goquery.NewDocumentFromReader(utf8Reader)
		if err != nil {
			logger.Errorf("生成dom树失败: %s", err.Error())
			continue
		}

		if 0 < crawler.Config.MaxDepth && crawler.Config.MaxDepth < req.Depth+1 {
			logger.Infof("当前页面已达到最大深度, 不再抓取新页面: url: %s, refer: %s, depth: %d", req.URL, req.Refer, req.Depth)
		} else {
			crawler.ParseLinkingPages(htmlDom, req)
		}
		crawler.ParseLinkingAssets(htmlDom, req)

		htmlString, err := htmlDom.Html()
		if err != nil {
			logger.Errorf("生成dom树失败: %s", err.Error())
			continue
		}
		htmlString = ReplaceHTMLCharacterEntities(htmlString, charset)
		fileContent, err := EncodeFromUTF8([]byte(htmlString), charset)
		if err != nil {
			logger.Errorf("编码失败: %s", err.Error())
			continue
		}
		fileDir, fileName, err := TransToLocalPath(crawler.MainSite, req.URL, PageURL)
		if err != nil {
			continue
		}
		err = WriteToLocalFile(crawler.Config.SitePath, fileDir, fileName, fileContent)
		if err != nil {
			logger.Errorf("写入文件失败: %s", err.Error())
			continue
		}
	}
}

// GetStaticAsset 工作协程, 从队列中获取任务, 获取静态资源并存储
func (crawler *Crawler) GetStaticAsset(num int) {
	for req := range crawler.AssetQueue {
		logger.Infof("取得静态文件任务: %+v", req)
		if 0 < crawler.Config.MaxDepth && crawler.Config.MaxDepth < req.Depth {
			logger.Infof("当前页面已达到最大深度, 不再抓取新页面: url: %s, refer: %s, depth: %d", req.URL, req.Refer, req.Depth)
		}
		resp, err := getURL(req.URL, req.Refer)
		if err != nil {
			logger.Errorf("请求页面失败: url: %s, refer: %s, error: %s, 重新入队列", req.URL, req.Refer, err.Error())
			continue
		}
		defer resp.Body.Close()
		bodyContent, err := ioutil.ReadAll(resp.Body)
		// 如果是css文件, 解析其中的链接, 否则直接存储.
		field, exist := resp.Header["Content-Type"]
		if exist && field[0] == "text/css" {
			bodyContent, err = crawler.parseCSSFile(bodyContent, req)
			if err != nil {
				logger.Errorf("解析css文件失败: %s", err.Error())
				continue
			}
		}
		fileDir, fileName, err := TransToLocalPath(crawler.MainSite, req.URL, AssetURL)
		if err != nil {
			continue
		}
		fileContent := bodyContent
		err = WriteToLocalFile(crawler.Config.SitePath, fileDir, fileName, fileContent)
		if err != nil {
			logger.Errorf("写入文件失败: %s", err.Error())
			continue
		}
	}
}

// EnqueuePage 页面任务入队列.
// 入队列前查询数据库记录, 如已有记录则不再接受.
func (crawler *Crawler) EnqueuePage(req *RequestTask) {
	crawler.PageQueue <- req
}

// EnqueueAsset 页面任务入队列.
// 入队列前查询数据库记录, 如已有记录则不再接受.
func (crawler *Crawler) EnqueueAsset(req *RequestTask) {
	crawler.AssetQueue <- req
}
