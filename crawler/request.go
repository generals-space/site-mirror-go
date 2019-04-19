package crawler

import (
	"bytes"
	"net/http"
	"net/url"
	"os"
	"path"
	"regexp"
	"strings"

	"gitee.com/generals-space/site-mirror-go.git/model"
	"github.com/PuerkitoBio/goquery"
)

func getURL(url, refer, ua string) (resp *http.Response, err error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", ua)
	req.Header.Set("Referer", refer)

	resp, err = client.Do(req)
	if err != nil {
		logger.Errorf("请求失败: url: %s, refer: %s, error: %s", url, refer, err.Error())
		return
	}
	return
}

func joinURL(baseURL, subURL string) (fullURL, fullURLWithoutFrag string) {
	baseURLObj, _ := url.Parse(baseURL)
	subURLObj, _ := url.Parse(subURL)
	fullURLObj := baseURLObj.ResolveReference(subURLObj)
	fullURL = fullURLObj.String()
	fullURLObj.Fragment = ""
	fullURLWithoutFrag = fullURLObj.String()
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
		// FindStringSubmatch返回值为切片, 第一个成员为模式匹配到的子串, 之后的成员分别是各分组匹配到的子串.
		// ta基本等效于FindStringSubmatch(metaInfo, 1), 只查询1个匹配项.
		matchedArray := charsetPattern.FindStringSubmatch(metaInfo)
		for _, matchedItem := range matchedArray[1:] {
			if matchedItem != "" {
				charset = matchedItem
				return
			}
		}
	}
	charset = "utf-8"
	return
}

// URLFilter ...
func URLFilter(fullURL string, urlType int, config *Config) (boolean bool) {
	urlObj, err := url.Parse(fullURL)
	if err != nil {
		logger.Errorf("解析地址失败: url: %s, %s", fullURL, err.Error())
		return
	}
	if urlType == model.URLTypePage && urlObj.Host != config.MainSite {
		logger.Infof("不抓取站外页面: %s", fullURL)
		return
	}
	if urlType == model.URLTypeAsset && urlObj.Host != config.MainSite && config.OutsiteAsset == false {
		logger.Infof("不抓取站外资源: %s", fullURL)
		return
	}
	if urlType == model.URLTypeAsset && strings.HasSuffix(fullURL, ".js") && config.NoJs == true {
		logger.Infof("不抓取js资源: %s", fullURL)
		return
	}
	if urlType == model.URLTypeAsset && strings.HasSuffix(fullURL, ".css") && config.NoCSS == true {
		logger.Infof("不抓取css资源: %s", fullURL)
		return
	}
	if urlType == model.URLTypeAsset && imagePattern.MatchString(fullURL) && config.NoImages == true {
		logger.Infof("不抓取图片资源: %s", fullURL)
		return
	}
	if urlType == model.URLTypeAsset && fontPattern.MatchString(fullURL) && config.NoFonts == true {
		logger.Infof("不抓取字体资源: %s", fullURL)
		return
	}
	for _, rule := range config.BlackList {
		pattern := regexp.MustCompile(rule)
		if pattern.MatchString(fullURL) {
			logger.Infof("不抓取黑名单中的url: %s", fullURL)
			return
		}
	}
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
