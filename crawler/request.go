package crawler

import (
	"bytes"
	"net/http"
	"os"
	"path"

	"github.com/PuerkitoBio/goquery"
)

const (
	PageURL int = iota
	AssetURL
)

// RequestTask ...
type RequestTask struct {
	URL         string
	URLType     int
	Refer       string
	Depth       int
	FailedTimes int
}

// Config ...
type Config struct {
	PageQueueSize  int
	AssetQueueSize int

	SiteDBPath string
	SitePath   string
	StartPage  string
	MaxDepth   int

	PageWorkerCount  int
	AssetWorkerCount int

	NoJs     string
	NoCSS    string
	NoImages string
	NoFonts  string
}

func getURL(url, refer string) (resp *http.Response, err error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "")
	req.Header.Set("Referer", refer)

	resp, err = client.Do(req)
	if err != nil {
		logger.Errorf("请求失败: url: %s, refer: %s, error: %s", url, refer, err.Error())
		return
	}
	return
}

// getPageCharset 解析页面, 从中获取页面编码信息
func getPageCharset(body []byte) (charset string, err error) {
	bodyReader := bytes.NewReader(body)
	dom, err := goquery.NewDocumentFromReader(bodyReader)
	if err != nil {
		logger.Errorf("生成dom树失败: %s", err.Error())
		return
	}
	var metaInfo string
	var exist bool
	metaInfo, exist = dom.Find("meta[charset]").Attr("charset")
	if exist {
		charset = metaInfo
		return
	}
	metaInfo, exist = dom.Find("meta[http-equiv]").Attr("content")
	if exist {
		// FindStringSubmatch返回值为切片, 第一个成员为模式匹配到的子串,
		// 之后的成员分别是各分组匹配到的子串.
		matchArray := charsetPattern.FindStringSubmatch(metaInfo)
		if len(matchArray) > 0 {
			charset = matchArray[1]
			return
		}
	}
	charset = "utf-8"
	return
}

// URLFilter ...
func URLFilter(fullURL string, urlType int, mainSite string) (boolean bool) {

	return true
}

// WriteToLocalFile ...
func WriteToLocalFile(baseDir string, fileDir string, fileName string, fileContent []byte) (err error) {
	fileDir = path.Join(baseDir, fileDir)
	err = os.MkdirAll(fileDir, os.ModePerm)
	filePath := path.Join(fileDir, fileName)
	file, err := os.Create(filePath)
	defer file.Close()

	_, err = file.Write(fileContent)
	if err != nil {
		logger.Errorf("写入文件失败: %s", err.Error())
	}
	return
}
