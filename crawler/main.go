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

// Start 启动n个工作协程
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
			logger.Infof("当前页面已达到最大深度, 不再抓取: req: %+v", req)
			continue
		}
		if req.FailedTimes > crawler.Config.MaxRetryTimes {
			logger.Infof("当前页面失败次数过多, 不再尝试: req: %+v", req)
			continue
		}

		resp, err := getURL(req.URL, req.Refer, crawler.Config.UserAgent)
		if err != nil {
			logger.Errorf("请求页面失败, 重新入队列: req: %+v, error: %s", req, err.Error())
			req.FailedTimes++
			crawler.EnqueuePage(req)
			continue
		} else if resp.StatusCode == 404 {
			// 抓取失败一般是5xx或403, 405等, 出现404基本上就没有重试的意义了, 可以直接放弃
			continue
		}
		defer resp.Body.Close()
		// 编码处理
		bodyContent, err := ioutil.ReadAll(resp.Body)
		charsetName, err := getPageCharset(bodyContent)
		if err != nil {
			logger.Errorf("获取页面编码失败: req: %+v, error: %s", req, err.Error())
			continue
		}
		charsetName = strings.ToLower(charsetName)
		logger.Debugf("当前页面编码: %s, req: %+v", charsetName, req)
		charset, exist := CharsetMap[charsetName]
		if !exist {
			logger.Debugf("未找到匹配的编码: req: %+v, error: %s", req, err.Error())
			continue
		}
		utf8Coutent, err := DecodeToUTF8(bodyContent, charset)
		if err != nil {
			logger.Errorf("页面解码失败: req: %+v, error: %s", req, err.Error())
			continue
		}
		utf8Reader := bytes.NewReader(utf8Coutent)
		htmlDom, err := goquery.NewDocumentFromReader(utf8Reader)
		if err != nil {
			logger.Errorf("生成dom树失败: req: %+v, error: %s", req, err.Error())
			continue
		}

		if 0 < crawler.Config.MaxDepth && crawler.Config.MaxDepth < req.Depth+1 {
			logger.Infof("当前页面已达到最大深度, 不再解析新页面: %+v", req)
		} else {
			crawler.ParseLinkingPages(htmlDom, req)
		}
		crawler.ParseLinkingAssets(htmlDom, req)

		htmlString, err := htmlDom.Html()
		if err != nil {
			logger.Errorf("获取页面Html()值失败: req: %+v, error: %s", req, err.Error())
			continue
		}
		htmlString = ReplaceHTMLCharacterEntities(htmlString, charset)
		fileContent, err := EncodeFromUTF8([]byte(htmlString), charset)
		if err != nil {
			logger.Errorf("页面编码失败: req: %+v, error: %s", req, err.Error())
			continue
		}
		fileDir, fileName, err := TransToLocalPath(crawler.MainSite, req.URL, PageURL)
		if err != nil {
			logger.Errorf("转换为本地链接失败: req: %+v, error: %s", req, err.Error())
			continue
		}
		err = WriteToLocalFile(crawler.Config.SitePath, fileDir, fileName, fileContent)
		if err != nil {
			logger.Errorf("写入文件失败: req: %+v, error: %s", req, err.Error())
			continue
		}
	}
}

// GetStaticAsset 工作协程, 从队列中获取任务, 获取静态资源并存储
func (crawler *Crawler) GetStaticAsset(num int) {
	for req := range crawler.AssetQueue {
		logger.Infof("取得静态文件任务: %+v", req)
		if req.FailedTimes > crawler.Config.MaxRetryTimes {
			logger.Infof("当前页面失败次数过多, 不再尝试: req: %+v", req)
			continue
		}

		resp, err := getURL(req.URL, req.Refer, crawler.Config.UserAgent)
		if err != nil {
			logger.Errorf("请求静态资源失败, 重新入队列: req: %+v, error: %s", req, err.Error())
			req.FailedTimes++
			crawler.EnqueueAsset(req)
			continue
		} else if resp.StatusCode == 404 {
			// 抓取失败一般是5xx或403, 405等, 出现404基本上就没有重试的意义了, 可以直接放弃
			continue
		}
		defer resp.Body.Close()
		bodyContent, err := ioutil.ReadAll(resp.Body)
		// 如果是css文件, 解析其中的链接, 否则直接存储.
		field, exist := resp.Header["Content-Type"]
		if exist && field[0] == "text/css" {
			bodyContent, err = crawler.parseCSSFile(bodyContent, req)
			if err != nil {
				logger.Errorf("解析css文件失败: req: %+v, error: %s", req, err.Error())
				continue
			}
		}
		fileDir, fileName, err := TransToLocalPath(crawler.MainSite, req.URL, AssetURL)
		if err != nil {
			logger.Errorf("转换为本地链接失败: req: %+v, error: %s", req, err.Error())
			continue
		}

		fileContent := bodyContent
		err = WriteToLocalFile(crawler.Config.SitePath, fileDir, fileName, fileContent)
		if err != nil {
			logger.Errorf("写入文件失败: req: %+v, error: %s", req, err.Error())
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
