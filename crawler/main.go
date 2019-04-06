package crawler

import (
	"bytes"
	"io/ioutil"
	"net/url"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/jinzhu/gorm"

	"gitee.com/generals-space/site-mirror-go.git/model"
	"gitee.com/generals-space/site-mirror-go.git/util"
)

var logger = util.NewLogger(os.Stdout)

// Crawler ...
type Crawler struct {
	PageQueue  chan *model.URLTask // 页面任务队列
	AssetQueue chan *model.URLTask // 静态资源任务队列

	Config   *Config
	MainSite string
	DBClient *gorm.DB
}

// NewCrawler 创建Crawler对象
func NewCrawler(config *Config) (crawler *Crawler, err error) {
	pageQueue := make(chan *model.URLTask, config.PageQueueSize)
	assetQueue := make(chan *model.URLTask, config.AssetQueueSize)
	urlObj, err := url.Parse(config.StartPage)
	if err != nil {
		logger.Errorf("解析起始地址失败: url: %s, %s", config.StartPage, err.Error())
		return
	}
	mainSite := urlObj.Host // Host成员带端口.
	dbClient, err := model.GetDB(config.SiteDBPath)
	if err != nil {
		logger.Errorf("初始化数据库失败: site db: %s, %s", config.SiteDBPath, err.Error())
		return
	}

	crawler = &Crawler{
		PageQueue:  pageQueue,
		AssetQueue: assetQueue,

		Config:   config,
		MainSite: mainSite,
		DBClient: dbClient,
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
	req := &model.URLTask{
		URL:         crawler.Config.StartPage,
		URLType:     PageURL,
		Refer:       "",
		Depth:       1,
		FailedTimes: 0,
	}
	crawler.EnqueuePage(req)

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
		err := model.DelPageTask(crawler.DBClient, req)
		if err != nil {
			logger.Infof("删除页面任务队列记录失败: req: %+v, error: %s", req, err.Error())
			continue
		}
		if 0 < crawler.Config.MaxDepth && crawler.Config.MaxDepth < req.Depth {
			logger.Infof("已达到最大深度, 不再抓取: req: %+v", req)
			continue
		}
		if req.FailedTimes > crawler.Config.MaxRetryTimes {
			logger.Infof("失败次数过多, 不再尝试: req: %+v", req)
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
			err = model.UpdateURLRecordStatus(crawler.DBClient, req.URL, model.URLTaskStatusFailed)
			if err != nil {
				logger.Errorf("更新任务记录状态失败: req: %+v, error: %s", req, err.Error())
			}
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
		err = model.UpdateURLRecordStatus(crawler.DBClient, req.URL, model.URLTaskStatusSuccess)
		if err != nil {
			logger.Errorf("更新任务记录状态失败: req: %+v, error: %s", req, err.Error())
			continue
		}
	}
}

// GetStaticAsset 工作协程, 从队列中获取任务, 获取静态资源并存储
func (crawler *Crawler) GetStaticAsset(num int) {
	for req := range crawler.AssetQueue {
		logger.Infof("取得静态资源任务: %+v", req)
		err := model.DelAssetTask(crawler.DBClient, req)
		if err != nil {
			logger.Infof("删除页面任务队列记录失败: req: %+v, error: %s", req, err.Error())
			continue
		}
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
			err = model.UpdateURLRecordStatus(crawler.DBClient, req.URL, model.URLTaskStatusFailed)
			if err != nil {
				logger.Errorf("更新任务记录状态失败: req: %+v, error: %s", req, err.Error())
			}
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
		err = model.UpdateURLRecordStatus(crawler.DBClient, req.URL, model.URLTaskStatusSuccess)
		if err != nil {
			logger.Errorf("更新任务记录状态失败: req: %+v, error: %s", req, err.Error())
			continue
		}
	}
}
