package crawler

import (
	"bytes"
	"net/http"
	"regexp"

	"github.com/PuerkitoBio/goquery"
)

const (
	PageURL int = iota
	AssetURL
)

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
	SiteDBPath     string
	StartPage      string
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

// CharsetPattern meta[http-equiv]元素, content属性中charset截取的正则模式.
var CharsetPattern = `charset\s*=\s*(\S*)\s*;?`

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
		// 普通的MatchString可直接接受模式字符串, 无需Compile,
		// 但是只能作为判断是否匹配, 无法从中获取其他信息.
		pattern := regexp.MustCompile(CharsetPattern)
		// FindStringSubmatch返回值为切片, 第一个成员为模式匹配到的子串,
		// 之后的成员分别是各分组匹配到的子串.
		matchArray := pattern.FindStringSubmatch(metaInfo)
		if len(matchArray) > 0 {
			charset = matchArray[1]
			return
		}
	}
	charset = "utf-8"
	return
}
