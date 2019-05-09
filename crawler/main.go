package crawler

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"github.com/jinzhu/gorm"

	"gitee.com/generals-space/site-mirror-go.git/model"
	"gitee.com/generals-space/site-mirror-go.git/util"
)

var logger *util.Logger

// Crawler ...
type Crawler struct {
	PageQueue  chan *model.URLRecord // 页面任务队列
	AssetQueue chan *model.URLRecord // 静态资源任务队列

	Config        *Config
	DBClient      *gorm.DB
	DBClientMutex *sync.Mutex
}

// NewCrawler 创建Crawler对象
func NewCrawler(config *Config, _logger *util.Logger) (crawler *Crawler, err error) {
	logger = _logger
	pageQueue := make(chan *model.URLRecord, config.PageWorkerCount*config.LinkRatioInSinglePage)
	assetQueue := make(chan *model.URLRecord, config.AssetWorkerCount*config.LinkRatioInSinglePage)
	urlObj, err := url.Parse(config.StartPage)
	if err != nil {
		logger.Errorf("解析起始地址失败: url: %s, %s", config.StartPage, err.Error())
		return
	}
	mainSite := urlObj.Host // Host成员带端口.
	config.MainSite = mainSite

	dbClient, err := model.GetDB(config.SiteDBPath)
	if err != nil {
		logger.Errorf("初始化数据库失败: site db: %s, %s", config.SiteDBPath, err.Error())
		return
	}
	crawler = &Crawler{
		PageQueue:  pageQueue,
		AssetQueue: assetQueue,

		Config:        config,
		DBClient:      dbClient,
		DBClientMutex: &sync.Mutex{},
	}

	err = crawler.LoadTaskQueue()
	if err != nil {
		logger.Errorf("加载任务队列失败: %s", err.Error())
		return
	}
	return
}

// Start 启动n个工作协程
func (crawler *Crawler) Start() {
	req := &model.URLRecord{
		URL:         crawler.Config.StartPage,
		URLType:     model.URLTypePage,
		Refer:       "",
		Depth:       1,
		FailedTimes: 0,
	}
	crawler.EnqueuePage(req)

	for i := 0; i < crawler.Config.PageWorkerCount; i++ {
		go crawler.GetHTMLPage(i)
	}
	for i := 0; i < crawler.Config.AssetWorkerCount; i++ {
		go crawler.GetStaticAsset(i)
	}
}

// getAndRead 发起请求获取页面或静态资源, 返回响应体内容.
func (crawler *Crawler) getAndRead(req *model.URLRecord) (body []byte, header http.Header, err error) {
	crawler.DBClientMutex.Lock()
	err = model.UpdateURLRecordStatus(crawler.DBClient, req.URL, model.URLTaskStatusPending)
	crawler.DBClientMutex.Unlock()
	if err != nil {
		logger.Infof("更新任务队列记录失败: req: %+v, error: %s", req, err.Error())
		return
	}

	if req.FailedTimes > crawler.Config.MaxRetryTimes {
		logger.Infof("失败次数过多, 不再尝试: req: %+v", req)
		return
	}

	if req.URLType == model.URLTypePage && 0 < crawler.Config.MaxDepth && crawler.Config.MaxDepth < req.Depth {
		logger.Infof("当前页面已达到最大深度, 不再抓取: req: %+v", req)
		return
	}

	resp, err := getURL(req.URL, req.Refer, crawler.Config.UserAgent)
	if err != nil {
		logger.Errorf("请求失败, 重新入队列: req: %+v, error: %s", req, err.Error())
		req.FailedTimes++
		if req.URLType == model.URLTypePage {
			crawler.EnqueuePage(req)
		} else {
			crawler.EnqueueAsset(req)
		}
		return
	} else if resp.StatusCode == 404 {
		// 抓取失败一般是5xx或403, 405等, 出现404基本上就没有重试的意义了, 可以直接放弃
		crawler.DBClientMutex.Lock()
		err = model.UpdateURLRecordStatus(crawler.DBClient, req.URL, model.URLTaskStatusFailed)
		crawler.DBClientMutex.Unlock()
		if err != nil {
			logger.Errorf("更新任务记录状态失败: req: %+v, error: %s", req, err.Error())
		}
		return
	}
	defer resp.Body.Close()

	header = resp.Header
	body, err = ioutil.ReadAll(resp.Body)

	return
}

// GetHTMLPage 工作协程, 从队列中获取任务, 请求html页面并解析
func (crawler *Crawler) GetHTMLPage(num int) {
	for req := range crawler.PageQueue {
		logger.Infof("取得页面任务: %+v", req)

		respBody, _, err := crawler.getAndRead(req)

		// 编码处理
		charsetName, err := getPageCharset(respBody)
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
		utf8Coutent, err := DecodeToUTF8(respBody, charset)
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

		logger.Debugf("准备进行页面解析: req: %+v", req)

		if 0 < crawler.Config.MaxDepth && crawler.Config.MaxDepth < req.Depth+1 {
			logger.Infof("当前页面已达到最大深度, 不再解析新页面: %+v", req)
		} else {
			crawler.ParseLinkingPages(htmlDom, req)
		}
		crawler.ParseLinkingAssets(htmlDom, req)

		logger.Debugf("页面解析完成, 准备写入本地文件: req: %+v", req)

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
		fileDir, fileName, err := TransToLocalPath(crawler.Config.MainSite, req.URL, model.URLTypePage)
		if err != nil {
			logger.Errorf("转换为本地链接失败: req: %+v, error: %s", req, err.Error())
			continue
		}
		err = WriteToLocalFile(crawler.Config.SitePath, fileDir, fileName, fileContent)
		if err != nil {
			logger.Errorf("写入文件失败: req: %+v, error: %s", req, err.Error())
			continue
		}

		logger.Debugf("页面任务写入本地文件成功: req: %+v", req)

		crawler.DBClientMutex.Lock()
		err = model.UpdateURLRecordStatus(crawler.DBClient, req.URL, model.URLTaskStatusSuccess)
		crawler.DBClientMutex.Unlock()
		if err != nil {
			logger.Errorf("更新任务记录状态失败: req: %+v, error: %s", req, err.Error())
			continue
		}
		logger.Debugf("页面任务完成: req: %+v", req)
	}
}

// GetStaticAsset 工作协程, 从队列中获取任务, 获取静态资源并存储
func (crawler *Crawler) GetStaticAsset(num int) {
	for req := range crawler.AssetQueue {
		logger.Infof("取得静态资源任务: %+v", req)

		respBody, respHeader, err := crawler.getAndRead(req)

		// 如果是css文件, 解析其中的链接, 否则直接存储.
		field, exist := respHeader["Content-Type"]
		if exist && field[0] == "text/css" {
			respBody, err = crawler.parseCSSFile(respBody, req)
			if err != nil {
				logger.Errorf("解析css文件失败: req: %+v, error: %s", req, err.Error())
				continue
			}
		}
		fileDir, fileName, err := TransToLocalPath(crawler.Config.MainSite, req.URL, model.URLTypeAsset)
		if err != nil {
			logger.Errorf("转换为本地链接失败: req: %+v, error: %s", req, err.Error())
			continue
		}

		err = WriteToLocalFile(crawler.Config.SitePath, fileDir, fileName, respBody)
		if err != nil {
			logger.Errorf("写入文件失败: req: %+v, error: %s", req, err.Error())
			continue
		}
		logger.Debugf("静态资源任务写入本地文件成功: req: %+v", req)

		crawler.DBClientMutex.Lock()
		err = model.UpdateURLRecordStatus(crawler.DBClient, req.URL, model.URLTaskStatusSuccess)
		crawler.DBClientMutex.Unlock()
		if err != nil {
			logger.Errorf("更新任务记录状态失败: req: %+v, error: %s", req, err.Error())
			continue
		}
		logger.Debugf("静态资源任务完成: req: %+v", req)
	}
}
